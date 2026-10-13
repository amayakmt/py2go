package runtime

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// The native part of the json module: the decoder, and the escaping of
// strings for the encoder.

const jsonErrType = "json__decoder_pm_JSONDecodeError"

func jsonError(msg, doc string, pos int) *PyError {
	runes := []rune(doc)
	if pos > len(runes) {
		pos = len(runes)
	}
	lineno := 1 + strings.Count(string(runes[:pos]), "\n")
	last := -1
	for i := pos - 1; i >= 0; i-- {
		if runes[i] == '\n' {
			last = i
			break
		}
	}
	colno := pos - last
	if last < 0 {
		colno = pos + 1
	}
	e := newError(jsonErrType, msg+": line "+strconv.Itoa(lineno)+" column "+strconv.Itoa(colno)+" (char "+strconv.Itoa(pos)+")")
	e.Attrs = map[string]any{"msg": msg, "doc": doc, "pos": int64(pos), "lineno": int64(lineno), "colno": int64(colno)}
	return e
}

type jsonParser struct {
	doc        []rune
	text       string
	pos        int
	objectHook any
	pairsHook  any
	parseFloat any
	parseInt   any
	parseConst any
}

func (p *jsonParser) skipWS() {
	for p.pos < len(p.doc) {
		switch p.doc[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *jsonParser) fail(msg string, pos int) {
	panic(jsonError(msg, p.text, pos))
}

func (p *jsonParser) value(idx int) any {
	if idx >= len(p.doc) {
		p.fail("Expecting value", idx)
	}
	c := p.doc[idx]
	switch {
	case c == '"':
		s, end := p.str(idx + 1)
		p.pos = end
		return s
	case c == '{':
		return p.object(idx + 1)
	case c == '[':
		return p.array(idx + 1)
	case c == 'n' && p.lit(idx, "null"):
		p.pos = idx + 4
		return nil
	case c == 't' && p.lit(idx, "true"):
		p.pos = idx + 4
		return true
	case c == 'f' && p.lit(idx, "false"):
		p.pos = idx + 5
		return false
	case c == 'N' && p.lit(idx, "NaN"):
		p.pos = idx + 3
		return p.constant("NaN", math.NaN())
	case c == 'I' && p.lit(idx, "Infinity"):
		p.pos = idx + 8
		return p.constant("Infinity", math.Inf(1))
	case c == '-' && p.lit(idx, "-Infinity"):
		p.pos = idx + 9
		return p.constant("-Infinity", math.Inf(-1))
	}
	if v, end, ok := p.number(idx); ok {
		p.pos = end
		return v
	}
	p.fail("Expecting value", idx)
	return nil
}

func (p *jsonParser) constant(name string, v float64) any {
	if p.parseConst != nil {
		return DynCall(p.parseConst, name)
	}
	return v
}

func (p *jsonParser) lit(idx int, s string) bool {
	rs := []rune(s)
	if idx+len(rs) > len(p.doc) {
		return false
	}
	for i, r := range rs {
		if p.doc[idx+i] != r {
			return false
		}
	}
	return true
}

func (p *jsonParser) number(idx int) (any, int, bool) {
	i := idx
	n := len(p.doc)
	if i < n && p.doc[i] == '-' {
		i++
	}
	if i >= n {
		return nil, 0, false
	}
	if p.doc[i] == '0' {
		i++
	} else if p.doc[i] >= '1' && p.doc[i] <= '9' {
		for i < n && p.doc[i] >= '0' && p.doc[i] <= '9' {
			i++
		}
	} else {
		return nil, 0, false
	}
	isFloat := false
	if i+1 < n && p.doc[i] == '.' && p.doc[i+1] >= '0' && p.doc[i+1] <= '9' {
		isFloat = true
		i++
		for i < n && p.doc[i] >= '0' && p.doc[i] <= '9' {
			i++
		}
	}
	if i < n && (p.doc[i] == 'e' || p.doc[i] == 'E') {
		j := i + 1
		if j < n && (p.doc[j] == '+' || p.doc[j] == '-') {
			j++
		}
		if j < n && p.doc[j] >= '0' && p.doc[j] <= '9' {
			for j < n && p.doc[j] >= '0' && p.doc[j] <= '9' {
				j++
			}
			isFloat = true
			i = j
		}
	}
	text := string(p.doc[idx:i])
	if isFloat {
		if p.parseFloat != nil {
			return DynCall(p.parseFloat, text), i, true
		}
		f, _ := strconv.ParseFloat(text, 64)
		return f, i, true
	}
	if p.parseInt != nil {
		return DynCall(p.parseInt, text), i, true
	}
	b, _ := new(big.Int).SetString(text, 10)
	return NormBig(b), i, true
}

func (p *jsonParser) str(idx int) (string, int) {
	start := idx - 1
	var sb strings.Builder
	for {
		if idx >= len(p.doc) {
			p.fail("Unterminated string starting at", start)
		}
		c := p.doc[idx]
		switch {
		case c == '"':
			return sb.String(), idx + 1
		case c == '\\':
			idx++
			if idx >= len(p.doc) {
				p.fail("Unterminated string starting at", start)
			}
			e := p.doc[idx]
			switch e {
			case '"', '\\', '/':
				sb.WriteRune(e)
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case 'u':
				r, ok := p.hex4(idx + 1)
				if !ok {
					p.fail("Invalid \\uXXXX escape", idx)
				}
				idx += 4
				if r >= 0xd800 && r <= 0xdbff && idx+6 < len(p.doc)+0 && p.doc[idx+1] == '\\' && p.doc[idx+2] == 'u' {
					if r2, ok2 := p.hex4(idx + 3); ok2 && r2 >= 0xdc00 && r2 <= 0xdfff {
						r = 0x10000 + (r-0xd800)<<10 + (r2 - 0xdc00)
						idx += 6
					}
				}
				if r >= 0xd800 && r <= 0xdfff {
					sb.WriteRune(utf8.RuneError)
				} else {
					sb.WriteRune(rune(r))
				}
			default:
				p.fail("Invalid \\escape", idx-1)
			}
			idx++
		case c < 0x20:
			p.fail("Invalid control character at", idx)
		default:
			sb.WriteRune(c)
			idx++
		}
	}
}

func (p *jsonParser) hex4(at int) (rune, bool) {
	if at+4 > len(p.doc) {
		return 0, false
	}
	v, err := strconv.ParseUint(string(p.doc[at:at+4]), 16, 32)
	if err != nil || strings.ContainsAny(string(p.doc[at:at+4]), "+-_") {
		return 0, false
	}
	return rune(v), true
}

func (p *jsonParser) array(idx int) any {
	items := []any{}
	p.pos = idx
	p.skipWS()
	if p.pos < len(p.doc) && p.doc[p.pos] == ']' {
		p.pos++
		return NewListOf(items)
	}
	for {
		v := p.value(p.pos)
		items = append(items, v)
		p.skipWS()
		if p.pos >= len(p.doc) {
			p.fail("Expecting ',' delimiter", p.pos)
		}
		switch p.doc[p.pos] {
		case ']':
			p.pos++
			return NewListOf(items)
		case ',':
			comma := p.pos
			p.pos++
			p.skipWS()
			if p.pos < len(p.doc) && p.doc[p.pos] == ']' {
				p.fail("Illegal trailing comma before end of array", comma)
			}
		default:
			p.fail("Expecting ',' delimiter", p.pos)
		}
	}
}

// NewListOf makes a list of objects from a slice.
func NewListOf(items []any) *List[any] { return &List[any]{items: items} }

func (p *jsonParser) object(idx int) any {
	var pairs []jsonPair
	p.pos = idx
	p.skipWS()
	if p.pos < len(p.doc) && p.doc[p.pos] == '}' {
		p.pos++
		return p.finishObject(pairs)
	}
	for {
		if p.pos >= len(p.doc) || p.doc[p.pos] != '"' {
			p.fail("Expecting property name enclosed in double quotes", p.pos)
		}
		key, end := p.str(p.pos + 1)
		p.pos = end
		p.skipWS()
		if p.pos >= len(p.doc) || p.doc[p.pos] != ':' {
			p.fail("Expecting ':' delimiter", p.pos)
		}
		p.pos++
		p.skipWS()
		v := p.value(p.pos)
		pairs = append(pairs, jsonPair{key, v})
		p.skipWS()
		if p.pos >= len(p.doc) {
			p.fail("Expecting ',' delimiter", p.pos)
		}
		switch p.doc[p.pos] {
		case '}':
			p.pos++
			return p.finishObject(pairs)
		case ',':
			comma := p.pos
			p.pos++
			p.skipWS()
			if p.pos < len(p.doc) && p.doc[p.pos] == '}' {
				p.fail("Illegal trailing comma before end of object", comma)
			}
		default:
			p.fail("Expecting ',' delimiter", p.pos)
		}
	}
}

type jsonPair struct {
	k string
	v any
}

func (p *jsonParser) finishObject(pairs []jsonPair) any {
	if p.pairsHook != nil {
		items := make([]any, len(pairs))
		for i, kv := range pairs {
			items[i] = NewVarTuple[any](kv.k, kv.v)
		}
		return DynCall(p.pairsHook, NewListOf(items))
	}
	d := NewDict[any, any]()
	for _, kv := range pairs {
		d.Set(kv.k, kv.v)
	}
	if p.objectHook != nil {
		return DynCall(p.objectHook, d)
	}
	return d
}

func init() {
	RegisterNative("json.loads", func(args []any, kw []KW) any {
		doc := Unbox(argAt(args, 0))
		var text string
		switch x := doc.(type) {
		case string:
			text = x
		default:
			bs, ok := bytesOf(x)
			if !ok {
				panic(newError("TypeError", "the JSON object must be str, bytes or bytearray, not "+TypeName(x)))
			}
			text = decodeJSONBytes(bs)
		}
		if strings.HasPrefix(text, "\xef\xbb\xbf") {
			panic(jsonError("Unexpected UTF-8 BOM (decode using utf-8-sig)", text, 0))
		}
		p := &jsonParser{doc: []rune(text), text: text}
		for _, name := range []string{"object_hook", "object_pairs_hook", "parse_float", "parse_int", "parse_constant"} {
			if v, ok := kwValue(kw, name); ok && !IsNone(v) {
				switch name {
				case "object_hook":
					p.objectHook = v
				case "object_pairs_hook":
					p.pairsHook = v
				case "parse_float":
					p.parseFloat = v
				case "parse_int":
					p.parseInt = v
				case "parse_constant":
					p.parseConst = v
				}
			}
		}
		p.skipWS()
		v := p.value(p.pos)
		p.skipWS()
		if p.pos != len(p.doc) {
			p.fail("Extra data", p.pos)
		}
		return v
	})
	RegisterNative("json.escape", func(args []any, kw []KW) any {
		nativeArgs("escape", args, 2, 2)
		return jsonQuote(Str(args[0]), Truth(args[1]))
	})
}

func decodeJSONBytes(bs string) string {
	b := []byte(bs)
	switch {
	case len(b) >= 4 && (b[0] == 0 && b[1] == 0 && (b[2] != 0 || b[3] != 0) || b[0] != 0 && b[1] == 0 && b[2] == 0 && b[3] == 0):
		if b[0] == 0 {
			return DecodeBytes(bs, "utf-32-be", "strict")
		}
		return DecodeBytes(bs, "utf-32-le", "strict")
	case len(b) >= 2 && (b[0] == 0 && b[1] != 0):
		return DecodeBytes(bs, "utf-16-be", "strict")
	case len(b) >= 2 && (b[0] != 0 && b[1] == 0):
		return DecodeBytes(bs, "utf-16-le", "strict")
	case len(b) >= 2 && (b[0] == 0xff && b[1] == 0xfe || b[0] == 0xfe && b[1] == 0xff):
		return DecodeBytes(bs, "utf-16", "strict")
	case strings.HasPrefix(bs, "\xef\xbb\xbf"):
		return DecodeBytes(bs, "utf-8-sig", "strict")
	}
	return DecodeBytes(bs, "utf-8", "strict")
}

// jsonQuote is the JSON string literal for s.
func jsonQuote(s string, ascii bool) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\b':
			b.WriteString(`\b`)
		case r == '\f':
			b.WriteString(`\f`)
		case r < 0x20:
			b.WriteString(`\u00` + string("0123456789abcdef"[r>>4]) + string("0123456789abcdef"[r&15]))
		case r < 0x7f || !ascii && r != utf8.RuneError:
			b.WriteRune(r)
		case r == 0x7f:
			b.WriteRune(r)
		case r > 0xffff:
			r -= 0x10000
			hi, lo := 0xd800+(r>>10), 0xdc00+(r&0x3ff)
			b.WriteString(`\u` + hex4(int(hi)) + `\u` + hex4(int(lo)))
		default:
			b.WriteString(`\u` + hex4(int(r)))
		}
	}
	b.WriteByte('"')
	return b.String()
}

func hex4(n int) string {
	const d = "0123456789abcdef"
	return string([]byte{d[n>>12&15], d[n>>8&15], d[n>>4&15], d[n&15]})
}

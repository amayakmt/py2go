package runtime

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// The core codecs of str.encode and bytes.decode: utf-8, ascii, latin-1,
// utf-16, utf-32 and unicode-escape, with the error handlers strict, ignore,
// replace, backslashreplace and xmlcharrefreplace.

// canonicalCodec maps the spellings of an encoding name to the one this file
// implements.
func canonicalCodec(name string) string {
	n := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), "_", "-"))
	switch n {
	case "utf-8", "utf8", "u8", "utf", "utf8-ucs2", "utf8-ucs4", "cp65001", "locale":
		return "utf-8"
	case "utf-8-sig":
		return "utf-8-sig"
	case "ascii", "us-ascii", "646", "ansi-x3.4-1968", "iso646-us", "cp367":
		return "ascii"
	case "latin-1", "latin1", "iso-8859-1", "iso8859-1", "8859", "l1", "latin", "cp819", "iso-ir-100":
		return "latin-1"
	case "utf-16", "utf16", "u16":
		return "utf-16"
	case "utf-16-le", "utf-16le":
		return "utf-16-le"
	case "utf-16-be", "utf-16be":
		return "utf-16-be"
	case "utf-32", "utf32", "u32":
		return "utf-32"
	case "utf-32-le", "utf-32le":
		return "utf-32-le"
	case "utf-32-be", "utf-32be":
		return "utf-32-be"
	case "unicode-escape", "unicodeescape":
		return "unicode-escape"
	case "raw-unicode-escape", "rawunicodeescape":
		return "raw-unicode-escape"
	}
	if c, ok := sbcsAliases[n]; ok {
		return c
	}
	panic(newError("LookupError", "unknown encoding: "+name))
}

func checkHandler(errs string) {
	switch errs {
	case "strict", "ignore", "replace", "backslashreplace", "xmlcharrefreplace", "surrogateescape", "surrogatepass":
	default:
		panic(newError("LookupError", "unknown error handler name '"+errs+"'"))
	}
}

// charEscape spells a character as Python's backslashreplace does.
func charEscape(r rune) string {
	switch {
	case r < 0x100:
		return fmt.Sprintf(`\x%02x`, r)
	case r < 0x10000:
		return fmt.Sprintf(`\u%04x`, r)
	}
	return fmt.Sprintf(`\U%08x`, r)
}

// EncodeStr is s.encode(encoding, errors), as the bytes it holds.
func EncodeStr(s, encoding, errs string) Bytes {
	codec := canonicalCodec(encoding)
	switch codec {
	case "utf-8":
		return Bytes(encodeSurrogates(s, errs))
	case "utf-8-sig":
		return Bytes("\xef\xbb\xbf" + encodeSurrogates(s, errs))
	case "ascii":
		return Bytes(encodeLimited(s, "ascii", 0x80, errs))
	case "latin-1":
		return Bytes(encodeLimited(s, "latin-1", 0x100, errs))
	case "utf-16", "utf-16-le", "utf-16-be":
		var out []byte
		var order binary.AppendByteOrder = binary.LittleEndian
		if codec == "utf-16-be" {
			order = binary.BigEndian
		}
		if codec == "utf-16" {
			out = append(out, 0xff, 0xfe)
		}
		for _, u := range utf16.Encode([]rune(s)) {
			out = order.AppendUint16(out, u)
		}
		return Bytes(out)
	case "utf-32", "utf-32-le", "utf-32-be":
		var out []byte
		var order binary.AppendByteOrder = binary.LittleEndian
		if codec == "utf-32-be" {
			order = binary.BigEndian
		}
		if codec == "utf-32" {
			out = append(out, 0xff, 0xfe, 0, 0)
		}
		for _, r := range s {
			out = order.AppendUint32(out, uint32(r))
		}
		return Bytes(out)
	case "unicode-escape":
		var b strings.Builder
		for _, r := range s {
			switch {
			case r == '\\':
				b.WriteString(`\\`)
			case r == '\t':
				b.WriteString(`\t`)
			case r == '\n':
				b.WriteString(`\n`)
			case r == '\r':
				b.WriteString(`\r`)
			case r < 0x20 || r == 0x7f || r >= 0x80:
				b.WriteString(charEscape(r))
			default:
				b.WriteRune(r)
			}
		}
		return Bytes(b.String())
	case "raw-unicode-escape":
		var b strings.Builder
		for _, r := range s {
			switch {
			case r < 0x100:
				b.WriteByte(byte(r))
			case r < 0x10000:
				fmt.Fprintf(&b, `\u%04x`, r)
			default:
				fmt.Fprintf(&b, `\U%08x`, r)
			}
		}
		return Bytes(b.String())
	}
	if table, ok := sbcsTables[codec]; ok {
		return Bytes(sbcsEncode(s, codec, table, errs))
	}
	panic(newError("LookupError", "unknown encoding: "+encoding))
}

// sbcsEncode encodes to a single-byte code page.
func sbcsEncode(s, codec, table, errs string) string {
	rev := map[rune]byte{}
	i := 0
	for _, r := range table {
		if r != 0xfffe {
			if _, dup := rev[r]; !dup {
				rev[r] = byte(i)
			}
		}
		i++
	}
	var out []byte
	runes := []rune(s)
	for pos, r := range runes {
		if b, ok := rev[r]; ok {
			out = append(out, b)
			continue
		}
		checkHandler(errs)
		switch errs {
		case "ignore":
		case "replace":
			out = append(out, '?')
		case "backslashreplace":
			out = append(out, charEscape(r)...)
		case "xmlcharrefreplace":
			out = append(out, fmt.Sprintf("&#%d;", r)...)
		default:
			panic(unicodeErr("UnicodeEncodeError", codec, s, pos, pos+1, "character maps to <undefined>",
				fmt.Sprintf("'charmap' codec can't encode character '%s' in position %d: character maps to <undefined>", charEscapeRepr(r), pos)))
		}
	}
	return string(out)
}

// wtf8 spells a surrogate code point as the three bytes Go strings hold it in.
func wtf8(r int) string {
	return string([]byte{byte(0xe0 | r>>12), byte(0x80 | (r>>6)&0x3f), byte(0x80 | r&0x3f)})
}

// encodeSurrogates handles the lone surrogates a str may hold when it is
// encoded as UTF-8: surrogateescape turns U+DC80..U+DCFF back into bytes,
// surrogatepass keeps them, and strict refuses them.
func encodeSurrogates(s, errs string) string {
	if strings.IndexByte(s, 0xed) < 0 {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0xed && i+2 < len(s) && s[i+1]&0xe0 == 0xa0 {
			r := int(s[i]&0x0f)<<12 | int(s[i+1]&0x3f)<<6 | int(s[i+2]&0x3f)
			switch {
			case errs == "surrogateescape" && r >= 0xdc80 && r <= 0xdcff:
				b.WriteByte(byte(r - 0xdc00))
				i += 2
				continue
			case errs == "surrogatepass":
			case errs == "ignore":
				i += 2
				continue
			case errs == "replace":
				b.WriteByte('?')
				i += 2
				continue
			default:
				pos := utf8.RuneCountInString(s[:i])
				panic(unicodeErr("UnicodeEncodeError", "utf-8", s, pos, pos+1, "surrogates not allowed",
					fmt.Sprintf("'utf-8' codec can't encode character '\\u%04x' in position %d: surrogates not allowed", r, pos)))
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// encodeLimited encodes to a codec that holds the code points below limit.
func encodeLimited(s, name string, limit rune, errs string) string {
	var out []byte
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r < limit {
			out = append(out, byte(r))
			continue
		}
		j := i
		for j < len(runes) && runes[j] >= limit {
			j++
		}
		checkHandler(errs)
		switch errs {
		case "ignore":
		case "replace":
			out = append(out, []byte(strings.Repeat("?", j-i))...)
		case "backslashreplace":
			for _, c := range runes[i:j] {
				out = append(out, charEscape(c)...)
			}
		case "xmlcharrefreplace":
			for _, c := range runes[i:j] {
				out = append(out, ("&#" + strconv.Itoa(int(c)) + ";")...)
			}
		default:
			what := "character " + quote(string(r))
			if j-i > 1 {
				what = fmt.Sprintf("characters in position %d-%d", i, j-1)
			} else {
				what = fmt.Sprintf("character '%s' in position %d", strings.Trim(charEscapeRepr(r), "'"), i)
			}
			panic(unicodeErr("UnicodeEncodeError", name, s, i, j, fmt.Sprintf("ordinal not in range(%d)", limit),
				fmt.Sprintf("'%s' codec can't encode %s: ordinal not in range(%d)", name, what, limit)))
		}
		i = j - 1
	}
	return string(out)
}

func charEscapeRepr(r rune) string {
	if r >= 0x20 && r < 0x7f {
		return string(r)
	}
	return charEscape(r)
}

// DecodeBytes is b.decode(encoding, errors).
func DecodeBytes(s, encoding, errs string) string {
	codec := canonicalCodec(encoding)
	switch codec {
	case "utf-8", "utf-8-sig":
		if codec == "utf-8-sig" {
			s = strings.TrimPrefix(s, "\xef\xbb\xbf")
		}
		return decodeUTF8(s, errs)
	case "ascii":
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			if s[i] < 0x80 {
				b.WriteByte(s[i])
				continue
			}
			checkHandler(errs)
			switch errs {
			case "ignore":
			case "replace":
				b.WriteRune(utf8.RuneError)
			case "backslashreplace":
				fmt.Fprintf(&b, `\x%02x`, s[i])
			case "surrogateescape":
				b.WriteString(wtf8(0xdc00 + int(s[i])))
			default:
				panic(unicodeErr("UnicodeDecodeError", "ascii", Bytes(s), i, i+1, "ordinal not in range(128)",
					fmt.Sprintf("'ascii' codec can't decode byte 0x%02x in position %d: ordinal not in range(128)", s[i], i)))
			}
		}
		return b.String()
	case "latin-1":
		return bytesToText(s)
	case "utf-16", "utf-16-le", "utf-16-be":
		var order binary.ByteOrder = binary.LittleEndian
		data := []byte(s)
		off := 0
		if codec == "utf-16" && len(data) >= 2 {
			switch {
			case data[0] == 0xff && data[1] == 0xfe:
				data, off = data[2:], 2
			case data[0] == 0xfe && data[1] == 0xff:
				data, order, off = data[2:], binary.BigEndian, 2
			}
		}
		if codec == "utf-16-be" {
			order = binary.BigEndian
		}
		var b strings.Builder
		name := "utf-16-" + map[bool]string{false: "le", true: "be"}[order == binary.ByteOrder(binary.BigEndian)]
		fail := func(start, end int, reason string) {
			decodeFailure(&b, name, data, off, start, end, reason, errs)
		}
		i := 0
		for ; i+2 <= len(data); i += 2 {
			u := order.Uint16(data[i:])
			switch {
			case u < 0xd800 || u > 0xdfff:
				b.WriteRune(rune(u))
			case u >= 0xdc00:
				fail(i, i+2, "illegal encoding")
			case i+4 > len(data):
				fail(i, len(data)&^1, "unexpected end of data")
				i = len(data) &^ 1
			default:
				u2 := order.Uint16(data[i+2:])
				if u2 < 0xdc00 || u2 > 0xdfff {
					fail(i, i+2, "illegal UTF-16 surrogate")
				} else {
					b.WriteRune(utf16.DecodeRune(rune(u), rune(u2)))
					i += 2
				}
			}
		}
		if i < len(data) {
			fail(i, len(data), "truncated data")
		}
		return b.String()
	case "utf-32", "utf-32-le", "utf-32-be":
		var order binary.ByteOrder = binary.LittleEndian
		data := []byte(s)
		off := 0
		if codec == "utf-32" && len(data) >= 4 {
			switch {
			case data[0] == 0xff && data[1] == 0xfe && data[2] == 0 && data[3] == 0:
				data, off = data[4:], 4
			case data[0] == 0 && data[1] == 0 && data[2] == 0xfe && data[3] == 0xff:
				data, order, off = data[4:], binary.BigEndian, 4
			}
		}
		if codec == "utf-32-be" {
			order = binary.BigEndian
		}
		var b strings.Builder
		name := "utf-32-" + map[bool]string{false: "le", true: "be"}[order == binary.ByteOrder(binary.BigEndian)]
		fail := func(start, end int, reason string) {
			decodeFailure(&b, name, data, off, start, end, reason, errs)
		}
		i := 0
		for ; i+4 <= len(data); i += 4 {
			u := order.Uint32(data[i:])
			switch {
			case u > 0x10ffff:
				fail(i, i+4, "code point not in range(0x110000)")
			case u >= 0xd800 && u <= 0xdfff:
				fail(i, i+4, "code point in surrogate code point range(0xd800, 0xe000)")
			default:
				b.WriteRune(rune(u))
			}
		}
		if i < len(data) {
			fail(i, len(data), "truncated data")
		}
		return b.String()
	case "unicode-escape", "raw-unicode-escape":
		return decodeEscapes(s, codec == "raw-unicode-escape")
	}
	if table, ok := sbcsTables[codec]; ok {
		var b strings.Builder
		runes := []rune(table)
		for i := 0; i < len(s); i++ {
			r := runes[s[i]]
			if r != 0xfffe {
				b.WriteRune(r)
				continue
			}
			decodeFailure(&b, "charmap", []byte(s), 0, i, i+1, "character maps to <undefined>", errs)
		}
		return b.String()
	}
	panic(newError("LookupError", "unknown encoding: "+encoding))
}

// decodeUTF8 follows CPython's decoder: the same errors, the same extent of
// each one, the same replacement.
func decodeUTF8(s, errs string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		c := s[i]
		if c < 0x80 {
			b.WriteByte(c)
			i++
			continue
		}
		need, lo, hi := 0, byte(0x80), byte(0xbf)
		switch {
		case c >= 0xc2 && c <= 0xdf:
			need = 1
		case c >= 0xe0 && c <= 0xef:
			need = 2
			if c == 0xe0 {
				lo = 0xa0
			}
			if c == 0xed {
				hi = 0x9f
			}
		case c >= 0xf0 && c <= 0xf4:
			need = 3
			if c == 0xf0 {
				lo = 0x90
			}
			if c == 0xf4 {
				hi = 0x8f
			}
		}
		end, reason := i+1, "invalid start byte"
		if need > 0 {
			ok := true
			for k := 1; k <= need; k++ {
				if i+k >= len(s) {
					reason, ok = "unexpected end of data", false
					break
				}
				cb := s[i+k]
				l, h := byte(0x80), byte(0xbf)
				if k == 1 {
					l, h = lo, hi
				}
				if cb < l || cb > h {
					reason, ok = "invalid continuation byte", false
					break
				}
				end = i + k + 1
			}
			if ok {
				b.WriteString(s[i : i+need+1])
				i += need + 1
				continue
			}
		}
		checkHandler(errs)
		switch errs {
		case "ignore":
		case "replace":
			b.WriteRune(utf8.RuneError)
		case "backslashreplace":
			for k := i; k < end; k++ {
				fmt.Fprintf(&b, `\x%02x`, s[k])
			}
		case "surrogateescape":
			for k := i; k < end; k++ {
				if s[k] < 0x80 {
					b.WriteByte(s[k])
				} else {
					b.WriteString(wtf8(0xdc00 + int(s[k])))
				}
			}
		default:
			if end-i == 1 {
				panic(unicodeErr("UnicodeDecodeError", "utf-8", Bytes(s), i, end, reason,
					fmt.Sprintf("'utf-8' codec can't decode byte 0x%02x in position %d: %s", s[i], i, reason)))
			}
			panic(unicodeErr("UnicodeDecodeError", "utf-8", Bytes(s), i, end, reason,
				fmt.Sprintf("'utf-8' codec can't decode bytes in position %d-%d: %s", i, end-1, reason)))
		}
		i = end
	}
	return b.String()
}

// decodeEscapes reads \xNN, \uNNNN, \UNNNNNNNN, \NNN, \n and the like.
func decodeEscapes(s string, raw bool) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && i+1 >= len(s) && !raw {
			panic(decodeError("unicodeescape", []byte(s[i:]), i, "\\ at end of string"))
		}
		if c != '\\' || i+1 >= len(s) {
			if c < 0x80 || raw {
				b.WriteRune(rune(c))
			} else {
				b.WriteRune(rune(c))
			}
			continue
		}
		i++
		hexRun := func(n int) {
			// i is the letter after the backslash. The digits that follow
			// are read up to n, and a short run is a truncated escape.
			k := 0
			var v uint64
			for k < n && i+1+k < len(s) {
				d, ok := hexVal(s[i+1+k])
				if !ok {
					break
				}
				v = v<<4 | uint64(d)
				k++
			}
			if k < n {
				reason := map[int]string{2: "truncated \\xXX escape", 4: "truncated \\uXXXX escape", 8: "truncated \\UXXXXXXXX escape"}[n]
				panic(decodeError("unicodeescape", []byte(s[i-1:i+1+k]), i-1, reason))
			}
			if v > 0x10ffff {
				panic(decodeError("unicodeescape", []byte(s[i-1:i+1+k]), i-1, "illegal Unicode character"))
			}
			b.WriteRune(rune(v))
			i += n
		}
		if raw {
			switch s[i] {
			case 'u':
				hexRun(4)
			case 'U':
				hexRun(8)
			default:
				b.WriteByte('\\')
				b.WriteByte(s[i])
			}
			continue
		}
		switch s[i] {
		case '\\':
			b.WriteByte('\\')
		case '\'':
			b.WriteByte('\'')
		case '"':
			b.WriteByte('"')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case 'a':
			b.WriteByte(7)
		case 'b':
			b.WriteByte(8)
		case 'f':
			b.WriteByte(12)
		case 'v':
			b.WriteByte(11)
		case '\n':
		case 'x':
			hexRun(2)
		case 'u':
			hexRun(4)
		case 'U':
			hexRun(8)
		case '0', '1', '2', '3', '4', '5', '6', '7':
			j := i
			for j < len(s) && j < i+3 && s[j] >= '0' && s[j] <= '7' {
				j++
			}
			v, _ := strconv.ParseUint(s[i:j], 8, 32)
			b.WriteRune(rune(v))
			i = j - 1
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// decodeFailure handles a decoding error of a fixed-width codec: strict
// raises it, the other handlers write their replacement. start and end are
// offsets in data, which begins off bytes into the input.
func decodeFailure(b *strings.Builder, codec string, data []byte, off, start, end int, reason, errs string) {
	checkHandler(errs)
	switch errs {
	case "ignore":
	case "replace":
		b.WriteRune(utf8.RuneError)
	case "backslashreplace":
		for _, c := range data[start:end] {
			fmt.Fprintf(b, `\x%02x`, c)
		}
	case "surrogateescape":
		for _, c := range data[start:end] {
			if c < 0x80 {
				b.WriteByte(c)
			} else {
				b.WriteString(wtf8(0xdc00 + int(c)))
			}
		}
	default:
		panic(decodeError(codec, data[start:end], off+start, reason))
	}
}

// decodeError is the UnicodeDecodeError with CPython's message.
func decodeError(codec string, bad []byte, start int, reason string) *PyError {
	if len(bad) == 1 {
		return unicodeErr("UnicodeDecodeError", codec, Bytes(bad), 0, 1, reason,
			fmt.Sprintf("'%s' codec can't decode byte 0x%02x in position %d: %s", codec, bad[0], start, reason))
	}
	return unicodeErr("UnicodeDecodeError", codec, Bytes(bad), 0, len(bad), reason,
		fmt.Sprintf("'%s' codec can't decode bytes in position %d-%d: %s", codec, start, start+len(bad)-1, reason))
}

// unicodeErr builds a UnicodeDecodeError or UnicodeEncodeError with the
// attributes CPython gives it.
func unicodeErr(class, encoding string, object any, start, end int, reason, msg string) *PyError {
	e := newError(class, msg)
	e.Attrs = map[string]any{"encoding": encoding, "object": object, "start": int64(start), "end": int64(end), "reason": reason}
	return e
}

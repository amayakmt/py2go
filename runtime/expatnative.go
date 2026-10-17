package runtime

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// A small XML 1.0 parser that reports what expat reports: the events of a
// document, in order, or the first error with expat's code, message, line
// and column. pyexpat (stdlib/pyexpat.py) is written over it.

type xmlErr struct {
	code      int
	line, col int
}

var expatMessages = map[int]string{
	1: "out of memory", 2: "syntax error", 3: "no element found", 4: "not well-formed (invalid token)",
	5: "unclosed token", 6: "partial character", 7: "mismatched tag", 8: "duplicate attribute",
	9: "junk after document element", 10: "illegal parameter entity reference", 11: "undefined entity",
	12: "recursive entity reference", 13: "asynchronous entity", 14: "reference to invalid character number",
	15: "reference to binary entity", 16: "reference to external entity in attribute",
	17: "XML or text declaration not at start of entity", 18: "unknown encoding",
	19: "encoding specified in XML declaration is incorrect", 20: "unclosed CDATA section",
	21: "error in processing external entity reference", 22: "document is not standalone",
	23: "unexpected parser state - please send a bug report", 24: "entity declared in parameter entity",
	25: "requested feature requires XML_DTD support in Expat", 26: "cannot change setting once parsing has begun",
	27: "unbound prefix", 28: "must not undeclare prefix", 29: "incomplete markup in parameter entity",
	30: "XML declaration not well-formed", 31: "text declaration not well-formed", 32: "illegal character(s) in public id",
	33: "parser suspended", 34: "parser not suspended", 35: "parsing aborted", 36: "parsing finished",
	37: "cannot suspend in external parameter entity", 38: "reserved prefix (xml) must not be undeclared or bound to another namespace name",
	39: "reserved prefix (xmlns) must not be declared or undeclared", 40: "prefix must not be bound to one of the reserved namespace names",
}

type xmlParser struct {
	src      []rune
	pos      int
	line     int
	lineBase int // position where the current line starts
	events   *List[any]
	nsSep    string
	ns       bool
	entities map[string]string
	nsStack  []map[string]string
	err      *xmlErr
	tokStart int // the start of the tag being read, or -1
}

func (p *xmlParser) col() int { return p.pos - p.lineBase }

func (p *xmlParser) fail(code int) {
	if code == 5 && p.tokStart >= 0 {
		// An unclosed token is reported where the token starts.
		p.failAt(5, p.tokStart)
	}
	if p.err == nil {
		p.err = &xmlErr{code, p.line, p.col()}
	}
	panic(p.err)
}

func (p *xmlParser) failAt(code, pos int) {
	// The line and column of an earlier position.
	line, base := 1, 0
	for i := 0; i < pos && i < len(p.src); i++ {
		if p.src[i] == '\n' {
			line++
			base = i + 1
		}
	}
	if p.err == nil {
		p.err = &xmlErr{code, line, pos - base}
	}
	panic(p.err)
}

func (p *xmlParser) eof() bool { return p.pos >= len(p.src) }

func (p *xmlParser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *xmlParser) advance() rune {
	r := p.src[p.pos]
	p.pos++
	if r == '\n' {
		p.line++
		p.lineBase = p.pos
	}
	return r
}

func (p *xmlParser) startsWith(s string) bool {
	i := p.pos
	for _, r := range s {
		if i >= len(p.src) || p.src[i] != r {
			return false
		}
		i++
	}
	return true
}

func (p *xmlParser) skip(s string) {
	for range []rune(s) {
		p.advance()
	}
}

func isXMLSpace(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }

func (p *xmlParser) skipSpace() bool {
	any := false
	for !p.eof() && isXMLSpace(p.peek()) {
		p.advance()
		any = true
	}
	return any
}

func isNameStart(r rune) bool {
	return r == ':' || r == '_' || unicode.IsLetter(r) || r >= 0xC0 && r != 0xD7 && r != 0xF7 && r < 0x2FF ||
		r >= 0x370 && r <= 0x1FFF && r != 0x37E || r >= 0x2070 && r <= 0x218F || r >= 0x3001 && r <= 0xD7FF || r >= 0xF900
}

func isNameChar(r rune) bool {
	return isNameStart(r) || r == '-' || r == '.' || unicode.IsDigit(r) || r == 0xB7 || r >= 0x300 && r <= 0x36F || r == 0x203F || r == 0x2040
}

func (p *xmlParser) name() string {
	start := p.pos
	if p.eof() {
		p.fail(5)
	}
	if !isNameStart(p.peek()) {
		p.fail(4)
	}
	for !p.eof() && isNameChar(p.peek()) {
		p.advance()
	}
	return string(p.src[start:p.pos])
}

func (p *xmlParser) emit(vals ...any) { p.events.Append(NewVarTuple[any](vals...)) }

// reference reads &name; or &#n; after the '&', and returns its text.
func (p *xmlParser) reference(inAttr bool) string {
	start := p.pos - 1
	if p.peek() == '#' {
		p.advance()
		base := 10
		if p.peek() == 'x' {
			p.advance()
			base = 16
		}
		ds := p.pos
		for !p.eof() && p.peek() != ';' {
			r := p.peek()
			if !(r >= '0' && r <= '9' || base == 16 && (r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F')) {
				p.failAt(4, p.pos)
			}
			p.advance()
		}
		if p.eof() {
			p.fail(5)
		}
		n, err := strconv.ParseInt(string(p.src[ds:p.pos]), base, 32)
		p.advance()
		if err != nil || n == 0 || n > 0x10FFFF || n >= 0xD800 && n <= 0xDFFF || n < 0x20 && n != 9 && n != 10 && n != 13 {
			p.failAt(14, start)
		}
		return string(rune(n))
	}
	nm := p.name()
	if p.peek() != ';' {
		if p.eof() {
			p.fail(5)
		}
		p.fail(4)
	}
	p.advance()
	switch nm {
	case "lt":
		return "<"
	case "gt":
		return ">"
	case "amp":
		return "&"
	case "apos":
		return "'"
	case "quot":
		return "\""
	}
	if v, ok := p.entities[nm]; ok {
		return v
	}
	p.failAt(11, start)
	return ""
}

func (p *xmlParser) quoted() string {
	q := p.peek()
	if q != '"' && q != '\'' {
		if p.eof() {
			p.fail(5)
		}
		p.fail(4)
	}
	p.advance()
	start := p.pos
	for !p.eof() && p.peek() != q {
		p.advance()
	}
	if p.eof() {
		p.fail(5)
	}
	s := string(p.src[start:p.pos])
	p.advance()
	return s
}

func (p *xmlParser) attrValue() string {
	q := p.peek()
	if q != '"' && q != '\'' {
		if p.eof() {
			p.fail(5)
		}
		p.fail(4)
	}
	p.advance()
	var b strings.Builder
	for {
		if p.eof() {
			p.fail(5)
		}
		r := p.peek()
		if r == q {
			p.advance()
			return b.String()
		}
		switch r {
		case '<':
			p.fail(4)
		case '&':
			p.advance()
			b.WriteString(p.reference(true))
			continue
		case '\t', '\n', '\r':
			p.advance()
			b.WriteByte(' ')
			continue
		}
		b.WriteRune(p.advance())
	}
}

func (p *xmlParser) comment() {
	p.skip("<!--")
	start := p.pos
	for {
		if p.eof() {
			p.fail(5)
		}
		if p.startsWith("--") {
			if !p.startsWith("-->") {
				p.fail(4)
			}
			text := string(p.src[start:p.pos])
			p.skip("-->")
			p.emit("comment", text)
			return
		}
		p.advance()
	}
}

func (p *xmlParser) pi() {
	tagStart := p.pos
	p.skip("<?")
	target := p.name()
	if strings.EqualFold(target, "xml") {
		p.failAt(17, tagStart)
	}
	data := ""
	if p.skipSpace() {
		start := p.pos
		for !p.eof() && !p.startsWith("?>") {
			p.advance()
		}
		data = string(p.src[start:p.pos])
	}
	if !p.startsWith("?>") {
		if p.eof() {
			p.fail(5)
		}
		p.fail(4)
	}
	p.skip("?>")
	p.emit("pi", target, data)
}

func (p *xmlParser) xmlDecl() {
	p.skip("<?xml")
	version, encoding, standalone := "", any(nil), int64(-1)
	for {
		p.skipSpace()
		if p.startsWith("?>") {
			p.skip("?>")
			break
		}
		if p.eof() {
			p.fail(5)
		}
		k := p.name()
		p.skipSpace()
		if p.peek() != '=' {
			p.fail(30)
		}
		p.advance()
		p.skipSpace()
		v := p.quoted()
		switch k {
		case "version":
			version = v
		case "encoding":
			encoding = v
		case "standalone":
			if v == "yes" {
				standalone = 1
			} else {
				standalone = 0
			}
		default:
			p.fail(30)
		}
	}
	p.emit("xmldecl", version, encoding, standalone)
}

func (p *xmlParser) doctype() {
	p.skip("<!DOCTYPE")
	if !p.skipSpace() {
		p.fail(2)
	}
	name := p.name()
	p.skipSpace()
	var sysid, pubid any
	if p.startsWith("SYSTEM") {
		p.skip("SYSTEM")
		p.skipSpace()
		sysid = p.quoted()
	} else if p.startsWith("PUBLIC") {
		p.skip("PUBLIC")
		p.skipSpace()
		pubid = p.quoted()
		p.skipSpace()
		sysid = p.quoted()
	}
	p.skipSpace()
	internal := false
	if p.peek() == '[' {
		internal = true
		p.emit("startdoctype", name, sysid, pubid, true)
		p.advance()
		p.internalSubset()
		p.skipSpace()
	} else {
		p.emit("startdoctype", name, sysid, pubid, false)
	}
	_ = internal
	if p.peek() != '>' {
		if p.eof() {
			p.fail(5)
		}
		p.fail(2)
	}
	p.advance()
	p.emit("enddoctype")
}

// internalSubset reads up to the closing ']', keeping the general entities it
// declares with a literal value.
func (p *xmlParser) internalSubset() {
	for {
		p.skipSpace()
		if p.eof() {
			p.fail(5)
		}
		if p.peek() == ']' {
			p.advance()
			return
		}
		if p.startsWith("<!--") {
			p.comment()
			continue
		}
		if p.startsWith("<?") {
			p.pi()
			continue
		}
		if p.startsWith("<!ENTITY") {
			p.skip("<!ENTITY")
			p.skipSpace()
			param := false
			if p.peek() == '%' {
				param = true
				p.advance()
				p.skipSpace()
			}
			nm := p.name()
			p.skipSpace()
			if p.peek() == '"' || p.peek() == '\'' {
				v := p.quoted()
				if !param {
					if _, dup := p.entities[nm]; !dup {
						p.entities[nm] = v
					}
				}
			}
			for !p.eof() && p.peek() != '>' {
				if p.peek() == '"' || p.peek() == '\'' {
					p.quoted()
					continue
				}
				p.advance()
			}
			if p.eof() {
				p.fail(5)
			}
			p.advance()
			continue
		}
		if p.startsWith("<!") {
			// ELEMENT, ATTLIST, NOTATION: skipped, quotes respected.
			for !p.eof() && p.peek() != '>' {
				if p.peek() == '"' || p.peek() == '\'' {
					p.quoted()
					continue
				}
				p.advance()
			}
			if p.eof() {
				p.fail(5)
			}
			p.advance()
			continue
		}
		if p.peek() == '%' {
			for !p.eof() && p.peek() != ';' {
				p.advance()
			}
			if !p.eof() {
				p.advance()
			}
			continue
		}
		p.fail(2)
	}
}

func (p *xmlParser) resolve(qname string, isAttr bool, start int) string {
	if !p.ns {
		return qname
	}
	prefix, local := "", qname
	if i := strings.IndexByte(qname, ':'); i >= 0 {
		prefix, local = qname[:i], qname[i+1:]
	}
	if prefix == "" && isAttr {
		return local
	}
	if prefix == "xml" {
		return "http://www.w3.org/XML/1998/namespace" + p.nsSep + local
	}
	for i := len(p.nsStack) - 1; i >= 0; i-- {
		if uri, ok := p.nsStack[i][prefix]; ok {
			if uri == "" {
				if prefix != "" {
					p.failAt(27, start)
				}
				return local
			}
			return uri + p.nsSep + local
		}
	}
	if prefix != "" {
		p.failAt(27, start)
	}
	return local
}

// element reads one element, after its '<'.
func (p *xmlParser) element() {
	tagStart := p.pos
	p.tokStart = tagStart
	p.advance() // '<'
	qname := p.name()
	type attr struct {
		name, value string
		pos         int
	}
	var attrs []attr
	seen := map[string]bool{}
	for {
		hadSpace := p.skipSpace()
		if p.eof() {
			p.fail(5)
		}
		if p.peek() == '>' || p.startsWith("/>") {
			break
		}
		if !hadSpace {
			p.fail(4)
		}
		apos := p.pos
		an := p.name()
		p.skipSpace()
		if p.peek() != '=' {
			if p.eof() {
				p.fail(5)
			}
			p.fail(4)
		}
		p.advance()
		p.skipSpace()
		av := p.attrValue()
		if seen[an] {
			p.failAt(8, apos)
		}
		seen[an] = true
		attrs = append(attrs, attr{an, av, apos})
	}
	empty := p.startsWith("/>")
	if empty {
		p.skip("/>")
	} else {
		p.advance()
	}
	p.tokStart = -1
	// Namespace declarations of this element.
	var declared []string
	if p.ns {
		scope := map[string]string{}
		var rest []attr
		for _, a := range attrs {
			switch {
			case a.name == "xmlns":
				scope[""] = a.value
				declared = append(declared, "")
			case strings.HasPrefix(a.name, "xmlns:"):
				scope[a.name[6:]] = a.value
				declared = append(declared, a.name[6:])
			default:
				rest = append(rest, a)
			}
		}
		attrs = rest
		p.nsStack = append(p.nsStack, scope)
		for _, pre := range declared {
			var pv any = pre
			if pre == "" {
				pv = nil
			}
			p.emit("startns", pv, scope[pre])
		}
	}
	name := p.resolve(qname, false, tagStart)
	flat := NewList[any]()
	for _, a := range attrs {
		flat.Append(p.resolve(a.name, true, a.pos))
		flat.Append(a.value)
	}
	p.emit("start", name, flat)
	if !empty {
		p.content(qname)
	}
	p.emit("end", name)
	if p.ns {
		p.nsStack = p.nsStack[:len(p.nsStack)-1]
		for i := len(declared) - 1; i >= 0; i-- {
			var pv any = declared[i]
			if declared[i] == "" {
				pv = nil
			}
			p.emit("endns", pv)
		}
	}
}

// content reads up to the end tag of the element named qname.
func (p *xmlParser) content(qname string) {
	var text strings.Builder
	flush := func() {
		if text.Len() > 0 {
			p.emit("data", text.String())
			text.Reset()
		}
	}
	for {
		if p.eof() {
			flush()
			p.fail(3)
		}
		r := p.peek()
		switch {
		case p.startsWith("</"):
			flush()
			tagStart := p.pos
			p.skip("</")
			nm := p.name()
			p.skipSpace()
			if p.peek() != '>' {
				if p.eof() {
					p.fail(5)
				}
				p.fail(4)
			}
			if nm != qname {
				p.failAt(7, tagStart+2)
			}
			p.advance()
			return
		case p.startsWith("<!--"):
			flush()
			p.comment()
		case p.startsWith("<![CDATA["):
			flush()
			p.skip("<![CDATA[")
			start := p.pos
			for !p.eof() && !p.startsWith("]]>") {
				p.advance()
			}
			if p.eof() {
				p.fail(20)
			}
			p.emit("startcdata")
			if p.pos > start {
				p.emit("data", string(p.src[start:p.pos]))
			}
			p.emit("endcdata")
			p.skip("]]>")
		case p.startsWith("<?"):
			flush()
			p.pi()
		case r == '<':
			flush()
			if p.pos+1 >= len(p.src) {
				p.fail(5)
			}
			p.element()
		case r == '&':
			// expat reports the text of each reference as a call of its own.
			flush()
			p.advance()
			text.WriteString(p.reference(false))
			flush()
		case r == '\r' || r == '\n':
			// and each line end.
			flush()
			p.advance()
			if r == '\r' && p.peek() == '\n' {
				p.advance()
			}
			text.WriteByte('\n')
			flush()
		default:
			if p.startsWith("]]>") {
				p.fail(4)
			}
			text.WriteRune(p.advance())
		}
	}
}

func (p *xmlParser) misc(afterRoot bool) bool {
	p.skipSpace()
	switch {
	case p.eof():
		return false
	case p.startsWith("<!--"):
		p.comment()
		return true
	case p.startsWith("<?xml") && p.pos+5 < len(p.src) && isXMLSpace(p.src[p.pos+5]):
		p.fail(17)
	case p.startsWith("<?"):
		p.pi()
		return true
	}
	return false
}

func (p *xmlParser) document() {
	if p.startsWith("<?xml") && (p.pos+5 < len(p.src) && isXMLSpace(p.src[p.pos+5]) || p.startsWith("<?xml?")) {
		p.xmlDecl()
	}
	for p.misc(false) {
	}
	if p.startsWith("<!DOCTYPE") {
		p.doctype()
		for p.misc(false) {
		}
	}
	if p.eof() {
		p.fail(3)
	}
	if p.peek() != '<' {
		// Text before the root: expat reads it as one token and fails after it.
		for !p.eof() && p.peek() != '<' && !isXMLSpace(p.peek()) {
			p.advance()
		}
		p.fail(4)
	}
	if p.pos+1 >= len(p.src) {
		p.fail(5)
	}
	if !isNameStart(p.src[p.pos+1]) {
		p.fail(4)
	}
	p.element()
	for p.misc(true) {
	}
	if !p.eof() {
		p.fail(9)
	}
}

func init() {
	// pyexpat.parse(text, namespace_separator) is (events, error): error is
	// None or (code, message, line, column).
	RegisterNative("pyexpat.parse", func(args []any, kw []KW) any {
		var text string
		switch x := Unbox(args[0]).(type) {
		case string:
			text = x
		default:
			s, _ := bytesOf(x)
			if utf8.ValidString(s) {
				text = s
			} else {
				text = DecodeLatin1(s)
			}
		}
		p := &xmlParser{src: []rune(text), line: 1, events: NewList[any](), entities: map[string]string{}, tokStart: -1}
		if sep := Unbox(args[1]); sep != nil {
			p.ns = true
			p.nsSep = sep.(string)
		}
		var result any
		func() {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(*xmlErr); ok {
						result = NewVarTuple[any](p.events, NewVarTuple[any](int64(e.code), expatMessages[e.code], int64(e.line), int64(e.col)))
						return
					}
					panic(r)
				}
			}()
			p.document()
			result = NewVarTuple[any](p.events, nil)
		}()
		return result
	})
	RegisterNative("pyexpat.error_string", func(args []any, kw []KW) any {
		if m, ok := expatMessages[int(DynInt(args[0]))]; ok {
			return m
		}
		return nil
	})
}

// DecodeLatin1 reads bytes as ISO-8859-1.
func DecodeLatin1(s string) string {
	r := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		r[i] = rune(s[i])
	}
	return string(r)
}

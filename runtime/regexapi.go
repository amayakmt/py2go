package runtime

import (
	"reflect"
	"strconv"
	"strings"
)

// The objects of the re module: compiled patterns and match objects.

// RePat is the object re.compile returns.
type RePat struct {
	p *RePattern
}

// ReMatch is a match object.
type ReMatch struct {
	pat      *RePat
	text     []rune
	subject  any // the string or bytes matched
	caps     []int
	pos, end int
	last     int
}

func (re *RePat) isBytes() bool { return re.p.isBytes }

func (re *RePat) PyClassName() string { return "Pattern" }

func (re *RePat) PyRepr() string {
	s := "re.compile(" + Repr(re.patternValue())
	if f := re.flagsText(); f != "" {
		s += ", " + f
	}
	return s + ")"
}

func (re *RePat) patternValue() any {
	if re.p.isBytes {
		return Bytes(re.p.Pattern)
	}
	return re.p.Pattern
}

func (re *RePat) flagsText() string {
	f := re.p.Flags
	if !re.p.isBytes {
		f &^= ReUNICODE
	}
	var parts []string
	for _, e := range []struct {
		bit  int
		name string
	}{{ReIGNORECASE, "re.IGNORECASE"}, {ReLOCALE, "re.LOCALE"}, {ReMULTILINE, "re.MULTILINE"}, {ReDOTALL, "re.DOTALL"}, {ReUNICODE, "re.UNICODE"}, {ReVERBOSE, "re.VERBOSE"}, {ReDEBUG, "re.DEBUG"}, {ReASCII, "re.ASCII"}} {
		if f&e.bit != 0 {
			parts = append(parts, e.name)
		}
	}
	return strings.Join(parts, "|")
}

func (re *RePat) PyEq(other any) (bool, bool) {
	o, ok := Unbox(other).(*RePat)
	if !ok {
		return false, true
	}
	return re.p.Pattern == o.p.Pattern && re.p.Flags == o.p.Flags && re.p.isBytes == o.p.isBytes, true
}

func (re *RePat) PyHash() int64 { return hashStr(re.p.Pattern) ^ int64(re.p.Flags) }

var rePatNames = names("match fullmatch search findall finditer sub subn split scanner")

func (re *RePat) PyDynTable() map[string]bool { return rePatNames }

func (re *RePat) PyGetAttr(name string) (any, bool) {
	switch name {
	case "pattern":
		return re.patternValue(), true
	case "flags":
		return int64(re.p.Flags), true
	case "groups":
		return int64(re.p.Groups), true
	case "groupindex":
		d := NewDict[any, any]()
		for _, nm := range groupNamesInOrder(re.p.GroupNames) {
			d.Set(nm, int64(re.p.GroupNames[nm]))
		}
		return d, true
	}
	if rePatNames[name] {
		return func(args ...any) any {
			r, _ := re.DynMethod(name, args, nil)
			return r
		}, true
	}
	return nil, false
}

func groupNamesInOrder(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// By group number, as the dict is ordered by definition.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && m[out[j]] < m[out[j-1]]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// subject reads a str or bytes argument for this pattern.
func (re *RePat) subjectOf(v any) ([]rune, any) {
	v = Unbox(v)
	if re.p.isBytes {
		bs, ok := bytesOf(v)
		if !ok {
			panic(newError("TypeError", "cannot use a bytes pattern on a string-like object"))
		}
		rs := make([]rune, len(bs))
		for i := 0; i < len(bs); i++ {
			rs[i] = rune(bs[i])
		}
		return rs, v
	}
	s, ok := v.(string)
	if !ok {
		if _, isB := bytesOf(v); isB {
			panic(newError("TypeError", "cannot use a string pattern on a bytes-like object"))
		}
		panic(newError("TypeError", "expected string or bytes-like object, got '"+TypeName(v)+"'"))
	}
	return []rune(s), v
}

func (re *RePat) slice(text []rune, a, b int) any {
	if re.p.isBytes {
		out := make([]byte, 0, b-a)
		for _, r := range text[a:b] {
			out = append(out, byte(r))
		}
		return Bytes(out)
	}
	return string(text[a:b])
}

func (re *RePat) newMatch(text []rune, subject any, d *ReMatchData, pos, end int) *ReMatch {
	return &ReMatch{pat: re, text: text, subject: subject, caps: d.Caps, pos: pos, end: end, last: d.Last}
}

func regionOf(args []any, kw []KW, from int, n int) (int, int) {
	pos, end := int(argInt(args, from, 0)), int(argInt(args, from+1, int64(n)))
	if v, ok := kwValue(kw, "pos"); ok {
		pos = int(DynInt(v))
	}
	if v, ok := kwValue(kw, "endpos"); ok && !IsNone(v) {
		end = int(DynInt(v))
	}
	if pos < 0 {
		pos = 0
	}
	if end > n {
		end = n
	}
	if end < pos {
		end = pos
	}
	return pos, end
}

func (re *RePat) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "match", "fullmatch", "search":
		text, subject := re.subjectOf(argAt1(args, kw, 0, "string"))
		pos, end := regionOf(args, kw, 1, len(text))
		var d *ReMatchData
		switch name {
		case "match":
			d = re.p.matchAt(text, pos, end, false)
		case "fullmatch":
			d = re.p.matchAt(text, pos, end, true)
		default:
			d = re.p.searchFrom(text, pos, end, false)
		}
		if d == nil {
			return nil, true
		}
		return re.newMatch(text, subject, d, pos, end), true
	case "findall":
		text, subject := re.subjectOf(argAt1(args, kw, 0, "string"))
		pos, end := regionOf(args, kw, 1, len(text))
		out := NewList[any]()
		re.iterate(text, subject, pos, end, func(m *ReMatch) bool {
			switch re.p.Groups {
			case 0:
				out.Append(m.groupValue(0))
			case 1:
				out.Append(m.groupOrEmpty(1))
			default:
				items := make([]any, re.p.Groups)
				for i := range items {
					items[i] = m.groupOrEmpty(i + 1)
				}
				out.Append(NewVarTuple(items...))
			}
			return true
		})
		return out, true
	case "finditer":
		text, subject := re.subjectOf(argAt1(args, kw, 0, "string"))
		pos, end := regionOf(args, kw, 1, len(text))
		var matches []any
		re.iterate(text, subject, pos, end, func(m *ReMatch) bool {
			matches = append(matches, m)
			return true
		})
		i := 0
		return NewIterator(func() (any, bool) {
			if i >= len(matches) {
				return nil, false
			}
			i++
			return matches[i-1], true
		}), true
	case "sub", "subn":
		repl := argAt1(args, kw, 0, "repl")
		text, subject := re.subjectOf(argAt1(args, kw, 1, "string"))
		count := int(argInt(args, 2, 0))
		if v, ok := kwValue(kw, "count"); ok {
			count = int(DynInt(v))
		}
		res, n := re.substitute(repl, text, subject, count)
		if name == "subn" {
			return NewVarTuple[any](res, int64(n)), true
		}
		return res, true
	case "split":
		text, subject := re.subjectOf(argAt1(args, kw, 0, "string"))
		maxsplit := int(argInt(args, 1, 0))
		if v, ok := kwValue(kw, "maxsplit"); ok {
			maxsplit = int(DynInt(v))
		}
		return re.split(text, subject, maxsplit), true
	}
	return nil, false
}

// argAt1 reads a positional-or-keyword argument.
func argAt1(args []any, kw []KW, i int, name string) any {
	if i < len(args) {
		return args[i]
	}
	if v, ok := kwValue(kw, name); ok {
		return v
	}
	panic(newError("TypeError", "missing required argument '"+name+"' (pos "+itoa(i+1)+")"))
}

// iterate calls f for each non-overlapping match, with Python's rule for
// empty matches: one next to the previous match is allowed, but the same
// empty match is not found twice.
func (re *RePat) iterate(text []rune, subject any, pos, end int, f func(*ReMatch) bool) {
	mustAdvance := false
	for pos <= end {
		d := re.p.searchFrom(text, pos, end, mustAdvance)
		if d == nil {
			return
		}
		m := re.newMatch(text, subject, d, 0, end)
		if !f(m) {
			return
		}
		s, e := d.Caps[0], d.Caps[1]
		pos = e
		mustAdvance = e == s
	}
}

func (re *RePat) substitute(repl any, text []rune, subject any, count int) (any, int) {
	repl = Unbox(repl)
	var tpl *reTemplate
	var fn any
	switch r := repl.(type) {
	case string:
		if re.p.isBytes {
			panic(newError("TypeError", "sequence item 0: expected a bytes-like object, str found"))
		}
		tpl = parseTemplate(r, re)
	default:
		if bs, ok := bytesOf(repl); ok {
			if !re.p.isBytes {
				panic(newError("TypeError", "sequence item 0: expected str instance, bytes found"))
			}
			tpl = parseTemplate(bytesToText(bs), re)
		} else if _, callable := repl.(interface{ PyCall([]any, []KW) (any, bool) }); callable || isFuncValue(repl) {
			fn = repl
		} else {
			panic(newError("TypeError", "decoding to str: need a bytes-like object, "+TypeName(repl)+" found"))
		}
	}
	var out []rune
	last := 0
	n := 0
	appendAny := func(v any) {
		switch x := Unbox(v).(type) {
		case string:
			out = append(out, []rune(x)...)
		default:
			if bs, ok := bytesOf(x); ok {
				for i := 0; i < len(bs); i++ {
					out = append(out, rune(bs[i]))
				}
				return
			}
			panic(newError("TypeError", "expected str instance, "+TypeName(x)+" found"))
		}
	}
	re.iterate(text, subject, 0, len(text), func(m *ReMatch) bool {
		if count > 0 && n >= count {
			return false
		}
		out = append(out, text[last:m.caps[0]]...)
		if fn != nil {
			appendAny(DynCall(fn, m))
		} else {
			out = append(out, []rune(tpl.expand(m))...)
		}
		last = m.caps[1]
		n++
		return true
	})
	out = append(out, text[last:]...)
	if re.p.isBytes {
		bs := make([]byte, len(out))
		for i, r := range out {
			bs[i] = byte(r)
		}
		return Bytes(bs), n
	}
	return string(out), n
}

func (re *RePat) split(text []rune, subject any, maxsplit int) any {
	out := NewList[any]()
	last := 0
	n := 0
	re.iterate(text, subject, 0, len(text), func(m *ReMatch) bool {
		if maxsplit > 0 && n >= maxsplit {
			return false
		}
		out.Append(re.slice(text, last, m.caps[0]))
		for g := 1; g <= re.p.Groups; g++ {
			out.Append(m.groupValue(g))
		}
		last = m.caps[1]
		n++
		return true
	})
	out.Append(re.slice(text, last, len(text)))
	return out
}

// ---- templates ----

type reTemplate struct {
	parts []any // string literal or int group number
}

func parseTemplate(t string, re *RePat) *reTemplate {
	tpl := &reTemplate{}
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			tpl.parts = append(tpl.parts, lit.String())
			lit.Reset()
		}
	}
	rs := []rune(t)
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		if c != '\\' {
			lit.WriteRune(c)
			continue
		}
		if i+1 >= len(rs) {
			panic(reError("bad escape (end of pattern)", t, i))
		}
		i++
		e := rs[i]
		switch {
		case e == 'g':
			if i+1 >= len(rs) || rs[i+1] != '<' {
				panic(reError("missing <", t, i+1))
			}
			j := i + 2
			for j < len(rs) && rs[j] != '>' {
				j++
			}
			if j >= len(rs) {
				panic(reError("missing >, unterminated name", t, i+2))
			}
			name := string(rs[i+2 : j])
			if name == "" {
				panic(reError("missing group name", t, i+2))
			}
			g, err := strconv.Atoi(name)
			if err != nil {
				idx, ok := re.p.GroupNames[name]
				if !ok {
					panic(newError("IndexError", "unknown group name '"+name+"'"))
				}
				g = idx
			} else if g > re.p.Groups {
				panic(reError("invalid group reference "+strconv.Itoa(g), t, i+2))
			}
			flush()
			tpl.parts = append(tpl.parts, g)
			i = j
		case e >= '1' && e <= '9':
			g := int(e - '0')
			if i+1 < len(rs) && rs[i+1] >= '0' && rs[i+1] <= '9' {
				g = g*10 + int(rs[i+1]-'0')
				i++
			}
			if g > re.p.Groups {
				panic(reError("invalid group reference "+strconv.Itoa(g), t, i))
			}
			flush()
			tpl.parts = append(tpl.parts, g)
		case e == '0':
			v, n := 0, 0
			for n < 2 && i+1 < len(rs) && rs[i+1] >= '0' && rs[i+1] <= '7' {
				v = v*8 + int(rs[i+1]-'0')
				i++
				n++
			}
			lit.WriteRune(rune(v))
		case e == 'n':
			lit.WriteByte('\n')
		case e == 't':
			lit.WriteByte('\t')
		case e == 'r':
			lit.WriteByte('\r')
		case e == 'f':
			lit.WriteByte('\f')
		case e == 'v':
			lit.WriteByte('\v')
		case e == 'a':
			lit.WriteByte(7)
		case e == 'b':
			lit.WriteByte(8)
		case e == '\\':
			lit.WriteByte('\\')
		case e < 128 && (e >= 'a' && e <= 'z' || e >= 'A' && e <= 'Z'):
			panic(reError("bad escape \\"+string(e), t, i-1))
		default:
			lit.WriteByte('\\')
			lit.WriteRune(e)
		}
	}
	flush()
	return tpl
}

func (t *reTemplate) expand(m *ReMatch) string {
	var b strings.Builder
	for _, p := range t.parts {
		switch x := p.(type) {
		case string:
			b.WriteString(x)
		case int:
			if m.caps[2*x] >= 0 {
				b.WriteString(string(m.text[m.caps[2*x]:m.caps[2*x+1]]))
			}
		}
	}
	return b.String()
}

// ---- match objects ----

func (m *ReMatch) PyClassName() string { return "Match" }

func (m *ReMatch) groupValue(g int) any {
	if m.caps[2*g] < 0 {
		return nil
	}
	return m.pat.slice(m.text, m.caps[2*g], m.caps[2*g+1])
}

func (m *ReMatch) groupOrEmpty(g int) any {
	if v := m.groupValue(g); v != nil {
		return v
	}
	if m.pat.p.isBytes {
		return Bytes("")
	}
	return ""
}

func (m *ReMatch) PyRepr() string {
	return "<re.Match object; span=(" + strconv.Itoa(m.caps[0]) + ", " + strconv.Itoa(m.caps[1]) + "), match=" + Repr(m.groupValue(0)) + ">"
}

func (m *ReMatch) PyTruth() bool { return true }

var reMatchNames = names("group groups groupdict start end span expand __getitem__")

func (m *ReMatch) PyDynTable() map[string]bool { return reMatchNames }

func (m *ReMatch) PyGetAttr(name string) (any, bool) {
	switch name {
	case "string":
		return m.subject, true
	case "re":
		return m.pat, true
	case "pos":
		return int64(m.pos), true
	case "endpos":
		return int64(m.end), true
	case "lastindex":
		if m.last == 0 {
			return nil, true
		}
		return int64(m.last), true
	case "lastgroup":
		v, _ := m.PyGetAttr("lastindex")
		if v == nil {
			return nil, true
		}
		if nm := m.pat.p.groupName(int(v.(int64))); nm != "" {
			return nm, true
		}
		return nil, true
	case "regs":
		items := make([]any, 0, len(m.caps)/2)
		for g := 0; g*2 < len(m.caps); g++ {
			items = append(items, NewVarTuple[any](int64(m.caps[2*g]), int64(m.caps[2*g+1])))
		}
		return NewVarTuple(items...), true
	}
	if reMatchNames[name] {
		return func(args ...any) any {
			r, _ := m.DynMethod(name, args, nil)
			return r
		}, true
	}
	return nil, false
}

// groupIndex resolves a group name or number.
func (m *ReMatch) groupIndex(g any) int {
	g = Unbox(g)
	switch x := g.(type) {
	case string:
		if idx, ok := m.pat.p.GroupNames[x]; ok {
			return idx
		}
		panic(newError("IndexError", "no such group"))
	}
	n, ok := intOf(g)
	if !ok || isBig(g) || n < 0 || int(n) > m.pat.p.Groups {
		panic(newError("IndexError", "no such group"))
	}
	return int(n)
}

func (m *ReMatch) GetItemAny(key any) any { return m.groupValue(m.groupIndex(key)) }

func (m *ReMatch) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "group":
		if len(args) == 0 {
			return m.groupValue(0), true
		}
		if len(args) == 1 {
			return m.groupValue(m.groupIndex(args[0])), true
		}
		items := make([]any, len(args))
		for i, a := range args {
			items[i] = m.groupValue(m.groupIndex(a))
		}
		return NewVarTuple(items...), true
	case "__getitem__":
		return m.GetItemAny(argAt(args, 0)), true
	case "groups":
		var def any
		if len(args) > 0 {
			def = args[0]
		} else if v, ok := kwValue(kw, "default"); ok {
			def = v
		}
		items := make([]any, m.pat.p.Groups)
		for g := range items {
			if v := m.groupValue(g + 1); v != nil {
				items[g] = v
			} else {
				items[g] = def
			}
		}
		return NewVarTuple(items...), true
	case "groupdict":
		var def any
		if len(args) > 0 {
			def = args[0]
		} else if v, ok := kwValue(kw, "default"); ok {
			def = v
		}
		d := NewDict[any, any]()
		for _, nm := range groupNamesInOrder(m.pat.p.GroupNames) {
			g := m.pat.p.GroupNames[nm]
			if v := m.groupValue(g); v != nil {
				d.Set(nm, v)
			} else {
				d.Set(nm, def)
			}
		}
		return d, true
	case "start", "end", "span":
		g := 0
		if len(args) > 0 {
			g = m.groupIndex(args[0])
		}
		s, e := int64(m.caps[2*g]), int64(m.caps[2*g+1])
		switch name {
		case "start":
			return s, true
		case "end":
			return e, true
		}
		return NewVarTuple[any](s, e), true
	case "expand":
		t := argAt(args, 0)
		if bs, ok := bytesOf(t); ok {
			return Bytes(parseTemplate(bytesToText(bs), m.pat).expand(m)), true
		}
		return parseTemplate(Str(t), m.pat).expand(m), true
	}
	return nil, false
}

// ---- the module's natives ----

var reCache = map[string]*RePat{}

func init() {
	RegisterNative("re.compile", func(args []any, kw []KW) any {
		nativeArgs("compile", args, 1, 2)
		flags := 0
		if len(args) > 1 {
			flags = int(DynInt(args[1]))
		}
		pat := Unbox(args[0])
		if rp, ok := pat.(*RePat); ok {
			if flags != 0 {
				panic(valueError("cannot process flags argument with a compiled pattern"))
			}
			return rp
		}
		isBytes := false
		var src string
		switch x := pat.(type) {
		case string:
			src = x
		default:
			bs, ok := bytesOf(x)
			if !ok {
				panic(newError("TypeError", "first argument must be string or compiled pattern"))
			}
			src, isBytes = bs, true
		}
		key := strconv.Itoa(flags) + map[bool]string{true: "b", false: "s"}[isBytes] + src
		if rp, ok := reCache[key]; ok {
			return rp
		}
		rp := &RePat{p: CompilePattern(src, flags, isBytes)}
		if len(reCache) > 512 {
			reCache = map[string]*RePat{}
		}
		reCache[key] = rp
		return rp
	})
	RegisterNative("re.escape", func(args []any, kw []KW) any {
		nativeArgs("escape", args, 1, 1)
		special := "()[]{}?*+-|^$\\.&~# \t\n\r\v\f"
		escape := func(rs []rune) []rune {
			var out []rune
			for _, r := range rs {
				if r < 128 && strings.ContainsRune(special, r) {
					out = append(out, '\\')
				}
				out = append(out, r)
			}
			return out
		}
		switch x := Unbox(args[0]).(type) {
		case string:
			return string(escape([]rune(x)))
		default:
			bs, ok := bytesOf(x)
			if !ok {
				panic(newError("TypeError", "expected str or bytes-like object"))
			}
			var out []byte
			for i := 0; i < len(bs); i++ {
				if strings.IndexByte(special, bs[i]) >= 0 {
					out = append(out, '\\')
				}
				out = append(out, bs[i])
			}
			return Bytes(out)
		}
	})
	RegisterNative("re.purge", func(args []any, kw []KW) any {
		reCache = map[string]*RePat{}
		return nil
	})
}

func isFuncValue(v any) bool {
	_, ok := v.(*Callable)
	if ok {
		return true
	}
	return reflect.ValueOf(v).Kind() == reflect.Func
}

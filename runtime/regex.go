package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// A backtracking regular expression engine with the syntax and the matching
// rules of Python's re module: groups, backreferences, lazy and possessive
// quantifiers, lookahead and lookbehind, atomic groups, conditionals, and the
// flags I, M, S, X and A. Strings are matched by code point.

// Flag values, as in the re module.
const (
	ReIGNORECASE = 2
	ReLOCALE     = 4
	ReMULTILINE  = 8
	ReDOTALL     = 16
	ReUNICODE    = 32
	ReVERBOSE    = 64
	ReDEBUG      = 128
	ReASCII      = 256
)

type reKind int

const (
	rLit reKind = iota
	rAny
	rSet
	rCat
	rAlt
	rRepeat
	rGroup
	rBackref
	rBOL
	rEOL
	rBOS
	rEOS
	rWordB
	rNotWordB
	rLook
	rAtomic
	rCond
	rEmpty
)

type reNode struct {
	kind     reKind
	r        rune
	fold     bool // ignore case
	dotall   bool
	multi    bool
	ascii    bool
	set      *reSet
	subs     []*reNode
	min, max int // max < 0: unbounded
	greedy   bool
	possess  bool
	group    int // group number for rGroup, rBackref, rCond
	ahead    bool
	neg      bool
	width    int // lookbehind width
}

type reSet struct {
	negate bool
	ranges [][2]rune
	cats   []rune // 'd', 'D', 's', 'S', 'w', 'W'
	fold   bool
	ascii  bool
}

// RePattern is a compiled pattern.
type RePattern struct {
	Pattern    string
	Flags      int
	Groups     int
	GroupNames map[string]int
	names      []string // by group number, "" when unnamed
	root       *reNode
	isBytes    bool
}

// ReError is the error for a pattern that does not compile.
func reError(msg string, pattern string, pos int) *PyError {
	text := msg
	if pos >= 0 {
		text += " at position " + strconv.Itoa(pos)
	}
	e := newError("re_pm_PatternError", text)
	e.Attrs = map[string]any{"msg": msg, "pattern": pattern, "pos": int64(pos), "lineno": nil, "colno": nil}
	if pos < 0 {
		e.Attrs["pos"] = nil
	}
	return e
}

type reParser struct {
	src     []rune
	pos     int
	flags   int
	groups  int
	names   map[string]int
	byNum   []string
	pattern string
	open    []int // groups open at this point (for backrefs to them)
}

// CompilePattern parses a pattern.
func CompilePattern(pattern string, flags int, isBytes bool) *RePattern {
	p := &reParser{src: []rune(pattern), flags: flags, names: map[string]int{}, byNum: []string{""}, pattern: pattern}
	injectedASCII := false
	if isBytes {
		injectedASCII = p.flags&ReASCII == 0
		p.src = nil
		for i := 0; i < len(pattern); i++ {
			p.src = append(p.src, rune(pattern[i]))
		}
		p.flags |= ReASCII
	}
	// Global flags such as (?i) at the start.
	root := p.parseAlt(0)
	if p.pos < len(p.src) {
		if p.src[p.pos] == ')' {
			panic(reError("unbalanced parenthesis", pattern, p.pos))
		}
		panic(reError("unexpected end of pattern", pattern, p.pos))
	}
	resolveRefs(root, p)
	reported := p.flags
	if injectedASCII {
		reported &^= ReASCII
	}
	re := &RePattern{Pattern: pattern, Flags: reported, Groups: p.groups, GroupNames: p.names, names: p.byNum, root: root, isBytes: isBytes}
	if !isBytes && reported&ReASCII == 0 {
		re.Flags |= ReUNICODE
	}
	return re
}

func resolveRefs(n *reNode, p *reParser) {
	if n == nil {
		return
	}
	if n.kind == rBackref && n.group > p.groups {
		panic(reError("invalid group reference "+strconv.Itoa(n.group), p.pattern, -1))
	}
	if n.kind == rLook && !n.ahead {
		w, ok := fixedWidth(n.subs[0])
		if !ok {
			panic(reError("look-behind requires fixed-width pattern", p.pattern, -1))
		}
		n.width = w
	}
	for _, s := range n.subs {
		resolveRefs(s, p)
	}
}

func fixedWidth(n *reNode) (int, bool) {
	switch n.kind {
	case rLit, rAny, rSet:
		return 1, true
	case rCat:
		t := 0
		for _, s := range n.subs {
			w, ok := fixedWidth(s)
			if !ok {
				return 0, false
			}
			t += w
		}
		return t, true
	case rAlt:
		w0 := -1
		for _, s := range n.subs {
			w, ok := fixedWidth(s)
			if !ok || (w0 >= 0 && w != w0) {
				return 0, false
			}
			w0 = w
		}
		if w0 < 0 {
			return 0, true
		}
		return w0, true
	case rRepeat:
		w, ok := fixedWidth(n.subs[0])
		if !ok || n.min != n.max {
			return 0, false
		}
		return w * n.min, true
	case rGroup, rAtomic:
		return fixedWidth(n.subs[0])
	case rBackref:
		return 0, false
	case rCond:
		return 0, false
	}
	return 0, true
}

func (p *reParser) more() bool { return p.pos < len(p.src) }
func (p *reParser) peek() rune { return p.src[p.pos] }

func (p *reParser) parseAlt(depth int) *reNode {
	var alts []*reNode
	alts = append(alts, p.parseCat(depth))
	for p.more() && p.peek() == '|' {
		p.pos++
		alts = append(alts, p.parseCat(depth))
	}
	if len(alts) == 1 {
		return alts[0]
	}
	return &reNode{kind: rAlt, subs: alts}
}

func (p *reParser) verbose() bool { return p.flags&ReVERBOSE != 0 }

func (p *reParser) parseCat(depth int) *reNode {
	var items []*reNode
	for p.more() {
		c := p.peek()
		if c == '|' {
			break
		}
		if c == ')' {
			if depth == 0 {
				panic(reError("unbalanced parenthesis", p.pattern, p.pos))
			}
			break
		}
		if p.verbose() {
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v' {
				p.pos++
				continue
			}
			if c == '#' {
				for p.more() && p.peek() != '\n' {
					p.pos++
				}
				continue
			}
		}
		start := p.pos
		atom := p.parseAtom(depth)
		if atom == nil {
			continue
		}
		atom = p.parseQuant(atom, start)
		items = append(items, atom)
	}
	if len(items) == 1 {
		return items[0]
	}
	if len(items) == 0 {
		return &reNode{kind: rEmpty}
	}
	return &reNode{kind: rCat, subs: items}
}

func (p *reParser) parseQuant(atom *reNode, atomStart int) *reNode {
	for p.more() {
		if p.verbose() {
			// whitespace between an atom and its quantifier
			save := p.pos
			for p.more() && strings.ContainsRune(" \t\n\r\f\v", p.peek()) {
				p.pos++
			}
			if !p.more() || !strings.ContainsRune("*+?{", p.peek()) {
				p.pos = save
				return atom
			}
		}
		c := p.peek()
		min, max := 0, -1
		bracePos := atomStart
		switch c {
		case '*':
			p.pos++
		case '+':
			min = 1
			p.pos++
		case '?':
			max = 1
			p.pos++
		case '{':
			save := p.pos
			bracePos = save + 1
			p.pos++
			lo, hi, ok := p.parseBraces()
			if !ok {
				p.pos = save
				return atom
			}
			min, max = lo, hi
		default:
			return atom
		}
		if atom.kind == rBOL || atom.kind == rEOL || atom.kind == rBOS || atom.kind == rEOS || atom.kind == rWordB || atom.kind == rNotWordB || atom.kind == rEmpty && false {
			panic(reError("nothing to repeat", p.pattern, atomStart))
		}
		if atom.kind == rRepeat && !atom.possess && atom.greedy && false {
			panic(reError("multiple repeat", p.pattern, p.pos-1))
		}
		greedy, possess := true, false
		if p.more() && p.peek() == '?' {
			greedy = false
			p.pos++
		} else if p.more() && p.peek() == '+' {
			possess = true
			p.pos++
		}
		if max >= 0 && min > max {
			panic(reError("min repeat greater than max repeat", p.pattern, bracePos))
		}
		atom = &reNode{kind: rRepeat, subs: []*reNode{atom}, min: min, max: max, greedy: greedy, possess: possess}
		if p.more() && strings.ContainsRune("*+?", p.peek()) {
			panic(reError("multiple repeat", p.pattern, p.pos))
		}
		if p.more() && p.peek() == '{' {
			save := p.pos
			p.pos++
			if _, _, ok := p.parseBraces(); ok {
				panic(reError("multiple repeat", p.pattern, save))
			}
			p.pos = save
		}
		return atom
	}
	return atom
}

// parseBraces reads "m,n}" after "{". It reports false when the text is not a
// quantifier, in which case the brace is a literal.
func (p *reParser) parseBraces() (int, int, bool) {
	readInt := func() (int, bool) {
		start := p.pos
		for p.more() && p.peek() >= '0' && p.peek() <= '9' {
			p.pos++
		}
		if p.pos == start {
			return 0, false
		}
		n, _ := strconv.Atoi(string(p.src[start:p.pos]))
		return n, true
	}
	lo, hasLo := readInt()
	hi := lo
	if p.more() && p.peek() == ',' {
		p.pos++
		h, hasHi := readInt()
		if hasHi {
			hi = h
		} else {
			hi = -1
		}
		if !hasLo {
			lo = 0
		}
	} else if !hasLo {
		return 0, 0, false
	}
	if !p.more() || p.peek() != '}' {
		return 0, 0, false
	}
	p.pos++
	return lo, hi, true
}

func (p *reParser) litNode(r rune) *reNode {
	return &reNode{kind: rLit, r: r, fold: p.flags&ReIGNORECASE != 0}
}

func (p *reParser) parseAtom(depth int) *reNode {
	c := p.peek()
	switch c {
	case '.':
		p.pos++
		return &reNode{kind: rAny, dotall: p.flags&ReDOTALL != 0}
	case '^':
		p.pos++
		return &reNode{kind: rBOL, multi: p.flags&ReMULTILINE != 0}
	case '$':
		p.pos++
		return &reNode{kind: rEOL, multi: p.flags&ReMULTILINE != 0}
	case '[':
		return p.parseSet()
	case '(':
		return p.parseGroup(depth)
	case '\\':
		return p.parseEscape()
	case '*', '+', '?':
		panic(reError("nothing to repeat", p.pattern, p.pos))
	case '{':
		save := p.pos
		p.pos++
		if _, _, ok := p.parseBraces(); ok {
			panic(reError("nothing to repeat", p.pattern, save))
		}
		p.pos = save + 1
		return p.litNode('{')
	}
	p.pos++
	return p.litNode(c)
}

func (p *reParser) parseGroup(depth int) *reNode {
	start := p.pos
	p.pos++ // (
	if p.more() && p.peek() == '?' {
		p.pos++
		if !p.more() {
			panic(reError("unexpected end of pattern", p.pattern, p.pos))
		}
		c := p.peek()
		switch c {
		case ':':
			p.pos++
			sub := p.parseAlt(depth + 1)
			p.expectClose(start)
			return &reNode{kind: rCat, subs: []*reNode{sub}}
		case 'P':
			p.pos++
			if !p.more() {
				panic(reError("unexpected end of pattern", p.pattern, p.pos))
			}
			switch p.peek() {
			case '<':
				p.pos++
				name := p.readName('>')
				return p.newGroup(name, depth, start)
			case '=':
				p.pos++
				name := p.readName(')')
				idx, ok := p.names[name]
				if !ok {
					panic(reError("unknown group name '"+name+"'", p.pattern, p.pos-len([]rune(name))-1))
				}
				return &reNode{kind: rBackref, group: idx, fold: p.flags&ReIGNORECASE != 0}
			}
			panic(reError("unknown extension ?P"+string(p.peek()), p.pattern, start+1))
		case '<':
			p.pos++
			if p.more() && (p.peek() == '=' || p.peek() == '!') {
				neg := p.peek() == '!'
				p.pos++
				sub := p.parseAlt(depth + 1)
				p.expectClose(start)
				return &reNode{kind: rLook, ahead: false, neg: neg, subs: []*reNode{sub}}
			}
			panic(reError("unknown extension ?<"+string(p.peekOr(' ')), p.pattern, start+1))
		case '=', '!':
			p.pos++
			sub := p.parseAlt(depth + 1)
			p.expectClose(start)
			return &reNode{kind: rLook, ahead: true, neg: c == '!', subs: []*reNode{sub}}
		case '>':
			p.pos++
			sub := p.parseAlt(depth + 1)
			p.expectClose(start)
			return &reNode{kind: rAtomic, subs: []*reNode{sub}}
		case '#':
			for p.more() && p.peek() != ')' {
				p.pos++
			}
			if !p.more() {
				panic(reError("missing ), unterminated comment", p.pattern, start))
			}
			p.pos++
			return nil
		case '(':
			p.pos++
			namePos := p.pos
			name := p.readName(')')
			idx, err := strconv.Atoi(name)
			if err == nil && idx > p.groups {
				panic(reError("invalid group reference "+name, p.pattern, namePos))
			}
			if err != nil {
				i, ok := p.names[name]
				if !ok {
					panic(reError("unknown group name '"+name+"'", p.pattern, start+3))
				}
				idx = i
			}
			yes := p.parseCat(depth + 1)
			var no *reNode
			if p.more() && p.peek() == '|' {
				p.pos++
				no = p.parseCat(depth + 1)
			}
			p.expectClose(start)
			subs := []*reNode{yes}
			if no != nil {
				subs = append(subs, no)
			}
			return &reNode{kind: rCond, group: idx, subs: subs}
		}
		// Inline flags: (?aiLmsux) or (?flags:...) or (?flags-flags:...)
		return p.parseFlags(depth, start)
	}
	return p.newGroup("", depth, start)
}

func (p *reParser) peekOr(d rune) rune {
	if p.more() {
		return p.peek()
	}
	return d
}

func (p *reParser) expectClose(start int) {
	if !p.more() || p.peek() != ')' {
		panic(reError("missing ), unterminated subpattern", p.pattern, start))
	}
	p.pos++
}

func (p *reParser) readName(end rune) string {
	start := p.pos
	for p.more() && p.peek() != end {
		p.pos++
	}
	if !p.more() {
		if end == '>' {
			panic(reError("missing >, unterminated name", p.pattern, start))
		}
		panic(reError("missing ), unterminated name", p.pattern, start))
	}
	name := string(p.src[start:p.pos])
	p.pos++
	if name == "" {
		panic(reError("missing group name", p.pattern, start))
	}
	if end == '>' && !isIdentifier(name) {
		panic(reError("bad character in group name '"+name+"'", p.pattern, start))
	}
	return name
}

func (p *reParser) newGroup(name string, depth, start int) *reNode {
	p.groups++
	idx := p.groups
	if name != "" {
		if _, dup := p.names[name]; dup {
			panic(reError("redefinition of group name '"+name+"' as group "+strconv.Itoa(idx)+"; was group "+strconv.Itoa(p.names[name]), p.pattern, start+4))
		}
		p.names[name] = idx
	}
	p.byNum = append(p.byNum, name)
	sub := p.parseAlt(depth + 1)
	p.expectClose(start)
	return &reNode{kind: rGroup, group: idx, subs: []*reNode{sub}}
}

func (p *reParser) parseFlags(depth, start int) *reNode {
	on, off := 0, 0
	neg := false
	for p.more() {
		c := p.peek()
		var f int
		switch c {
		case 'i':
			f = ReIGNORECASE
		case 'm':
			f = ReMULTILINE
		case 's':
			f = ReDOTALL
		case 'x':
			f = ReVERBOSE
		case 'a':
			f = ReASCII
		case 'u':
			f = ReUNICODE
		case 'L':
			f = ReLOCALE
		case '-':
			neg = true
			p.pos++
			continue
		case ')', ':':
			goto done
		default:
			panic(reError("unknown extension ?"+string(c), p.pattern, start+1))
		}
		if neg {
			off |= f
		} else {
			on |= f
		}
		p.pos++
	}
done:
	if !p.more() {
		panic(reError("missing -, : or )", p.pattern, p.pos))
	}
	if p.peek() == ')' {
		// Global flags.
		p.pos++
		p.flags |= on
		return nil
	}
	p.pos++ // :
	saved := p.flags
	p.flags = (p.flags | on) &^ off
	sub := p.parseAlt(depth + 1)
	p.flags = saved
	p.expectClose(start)
	return &reNode{kind: rCat, subs: []*reNode{sub}}
}

func (p *reParser) parseEscape() *reNode {
	start := p.pos
	p.pos++
	if !p.more() {
		panic(reError("bad escape (end of pattern)", p.pattern, start))
	}
	c := p.peek()
	p.pos++
	asc := p.flags&ReASCII != 0
	switch c {
	case 'd', 'D', 's', 'S', 'w', 'W':
		return &reNode{kind: rSet, set: &reSet{cats: []rune{c}, ascii: asc}}
	case 'b':
		return &reNode{kind: rWordB, ascii: asc}
	case 'B':
		return &reNode{kind: rNotWordB, ascii: asc}
	case 'A':
		return &reNode{kind: rBOS}
	case 'Z':
		return &reNode{kind: rEOS}
	case 'n':
		return p.litNode('\n')
	case 't':
		return p.litNode('\t')
	case 'r':
		return p.litNode('\r')
	case 'f':
		return p.litNode('\f')
	case 'v':
		return p.litNode('\v')
	case 'a':
		return p.litNode(7)
	case '0':
		// Octal escape.
		v := 0
		n := 0
		for n < 2 && p.more() && p.peek() >= '0' && p.peek() <= '7' {
			v = v*8 + int(p.peek()-'0')
			p.pos++
			n++
		}
		return p.litNode(rune(v))
	case 'x':
		return p.litNode(p.hexEscape(2, start))
	case 'u':
		if p.flags&ReASCII != 0 && false {
			panic(reError("bad escape \\u", p.pattern, start))
		}
		return p.litNode(p.hexEscape(4, start))
	case 'U':
		return p.litNode(p.hexEscape(8, start))
	}
	if c >= '1' && c <= '9' {
		// Three octal digits are a character; otherwise a group number of
		// one or two digits.
		if c <= '3' && p.pos+1 < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '7' && p.src[p.pos+1] >= '0' && p.src[p.pos+1] <= '7' {
			v := int(c-'0')*64 + int(p.src[p.pos]-'0')*8 + int(p.src[p.pos+1]-'0')
			p.pos += 2
			return p.litNode(rune(v))
		}
		n := int(c - '0')
		if p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
			n = n*10 + int(p.src[p.pos]-'0')
			p.pos++
		}
		if n > p.groups {
			panic(reError("invalid group reference "+strconv.Itoa(n), p.pattern, start+1))
		}
		for _, o := range p.open {
			if o == n {
				panic(reError("cannot refer to an open group", p.pattern, start))
			}
		}
		return &reNode{kind: rBackref, group: n, fold: p.flags&ReIGNORECASE != 0}
	}
	if c < 128 && (c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
		panic(reError("bad escape \\"+string(c), p.pattern, start))
	}
	return p.litNode(c)
}

func (p *reParser) hexEscape(n, start int) rune {
	if p.pos+n > len(p.src) {
		panic(reError("incomplete escape \\"+string(p.src[start+1:]), p.pattern, start))
	}
	v, err := strconv.ParseUint(string(p.src[p.pos:p.pos+n]), 16, 32)
	if err != nil {
		panic(reError("incomplete escape \\"+string(p.src[start+1:p.pos+n]), p.pattern, start))
	}
	p.pos += n
	return rune(v)
}

func (p *reParser) parseSet() *reNode {
	start := p.pos
	p.pos++
	set := &reSet{fold: p.flags&ReIGNORECASE != 0, ascii: p.flags&ReASCII != 0}
	if p.more() && p.peek() == '^' {
		set.negate = true
		p.pos++
	}
	first := true
	for {
		if !p.more() {
			panic(reError("unterminated character set", p.pattern, start))
		}
		c := p.peek()
		if c == ']' && !first {
			p.pos++
			break
		}
		first = false
		atStart := p.pos
		lo, isCat := p.setChar(set)
		if isCat {
			if p.pos+1 < len(p.src) && p.peek() == '-' && p.src[p.pos+1] != ']' {
				endText := p.pos + 2
				if p.src[p.pos+1] == '\\' && p.pos+2 < len(p.src) {
					endText++
				}
				panic(reError("bad character range "+string(p.src[atStart:endText]), p.pattern, atStart))
			}
			continue
		}
		if p.pos+1 < len(p.src) && p.peek() == '-' && p.src[p.pos+1] != ']' {
			p.pos++
			hi, cat := p.setChar(set)
			if cat {
				panic(reError("bad character range "+string(lo)+"-"+string(p.src[p.pos-2:p.pos]), p.pattern, p.pos))
			}
			if hi < lo {
				panic(reError("bad character range "+string(lo)+"-"+string(hi), p.pattern, p.pos-3))
			}
			set.ranges = append(set.ranges, [2]rune{lo, hi})
			continue
		}
		set.ranges = append(set.ranges, [2]rune{lo, lo})
	}
	return &reNode{kind: rSet, set: set}
}

// setChar reads one character of a set; a class escape such as \d is added
// to the set and reported.
func (p *reParser) setChar(set *reSet) (rune, bool) {
	c := p.peek()
	p.pos++
	if c != '\\' {
		return c, false
	}
	if !p.more() {
		panic(reError("unterminated character set", p.pattern, p.pos))
	}
	e := p.peek()
	p.pos++
	switch e {
	case 'd', 'D', 's', 'S', 'w', 'W':
		set.cats = append(set.cats, e)
		return 0, true
	case 'n':
		return '\n', false
	case 't':
		return '\t', false
	case 'r':
		return '\r', false
	case 'f':
		return '\f', false
	case 'v':
		return '\v', false
	case 'a':
		return 7, false
	case 'b':
		return 8, false
	case 'x':
		return p.hexEscape(2, p.pos-2), false
	case 'u':
		return p.hexEscape(4, p.pos-2), false
	case 'U':
		return p.hexEscape(8, p.pos-2), false
	case '0', '1', '2', '3', '4', '5', '6', '7':
		v := int(e - '0')
		n := 1
		for n < 3 && p.more() && p.peek() >= '0' && p.peek() <= '7' {
			v = v*8 + int(p.peek()-'0')
			p.pos++
			n++
		}
		return rune(v), false
	}
	if e < 128 && (e >= 'a' && e <= 'z' || e >= 'A' && e <= 'Z') {
		panic(reError("bad escape \\"+string(e), p.pattern, p.pos-2))
	}
	return e, false
}

// ---- matching ----

type reMatcher struct {
	re    *RePattern
	text  []rune
	caps  []int // 2 per group, -1 when unset
	steps int
	end   int // the end of the searched region
	last  int // the group that closed last
}

func isWord(r rune, ascii bool) bool {
	if ascii || r < 128 {
		return r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
	}
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Nd, r) || unicode.Is(unicode.Mn, r) && false || unicode.IsNumber(r)
}

func isSpaceRe(r rune, ascii bool) bool {
	if ascii || r < 128 {
		return r == ' ' || r >= 9 && r <= 13 || r >= 0x1c && r <= 0x1f
	}
	return spaceRanges.has(r)
}

func isDigitRe(r rune, ascii bool) bool {
	if ascii || r < 128 {
		return r >= '0' && r <= '9'
	}
	return decimalRanges.has(r)
}

func (s *reSet) contains(r rune) bool {
	in := s.has(r)
	if !in && s.fold {
		for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
			if s.has(f) {
				in = true
				break
			}
		}
	}
	return in != s.negate
}

func (s *reSet) has(r rune) bool {
	for _, rg := range s.ranges {
		if r >= rg[0] && r <= rg[1] {
			return true
		}
	}
	for _, c := range s.cats {
		switch c {
		case 'd':
			if isDigitRe(r, s.ascii) {
				return true
			}
		case 'D':
			if !isDigitRe(r, s.ascii) {
				return true
			}
		case 's':
			if isSpaceRe(r, s.ascii) {
				return true
			}
		case 'S':
			if !isSpaceRe(r, s.ascii) {
				return true
			}
		case 'w':
			if isWord(r, s.ascii) {
				return true
			}
		case 'W':
			if !isWord(r, s.ascii) {
				return true
			}
		}
	}
	return false
}

func foldEq(a, b rune) bool {
	if a == b {
		return true
	}
	for f := unicode.SimpleFold(a); f != a; f = unicode.SimpleFold(f) {
		if f == b {
			return true
		}
	}
	return false
}

func (m *reMatcher) wordAt(i int, ascii bool) bool {
	return i >= 0 && i < len(m.text) && isWord(m.text[i], ascii)
}

// match tries node n at pos; on success it calls k with the end position and
// returns what k returns.
func (m *reMatcher) match(n *reNode, pos int, k func(int) bool) bool {
	switch n.kind {
	case rEmpty:
		return k(pos)
	case rLit:
		if pos < m.end && (m.text[pos] == n.r || n.fold && foldEq(m.text[pos], n.r)) {
			return k(pos + 1)
		}
		return false
	case rAny:
		if pos < m.end && (n.dotall || m.text[pos] != '\n') {
			return k(pos + 1)
		}
		return false
	case rSet:
		if pos < m.end && n.set.contains(m.text[pos]) {
			return k(pos + 1)
		}
		return false
	case rCat:
		return m.cat(n.subs, 0, pos, k)
	case rAlt:
		for _, s := range n.subs {
			if m.match(s, pos, k) {
				return true
			}
		}
		return false
	case rGroup:
		i := n.group * 2
		return m.match(n.subs[0], pos, func(e int) bool {
			os, oe, ol := m.caps[i], m.caps[i+1], m.last
			m.caps[i], m.caps[i+1], m.last = pos, e, n.group
			if k(e) {
				return true
			}
			m.caps[i], m.caps[i+1], m.last = os, oe, ol
			return false
		})
	case rBackref:
		i := n.group * 2
		s, e := m.caps[i], m.caps[i+1]
		if s < 0 {
			return false
		}
		l := e - s
		if pos+l > m.end {
			return false
		}
		for j := 0; j < l; j++ {
			a, b := m.text[s+j], m.text[pos+j]
			if a != b && !(n.fold && foldEq(a, b)) {
				return false
			}
		}
		return k(pos + l)
	case rBOL:
		if pos == 0 || n.multi && m.text[pos-1] == '\n' {
			return k(pos)
		}
		return false
	case rEOL:
		if pos == m.end || (pos == m.end-1 && m.text[pos] == '\n') || n.multi && m.text[pos] == '\n' {
			return k(pos)
		}
		return false
	case rBOS:
		if pos == 0 {
			return k(pos)
		}
		return false
	case rEOS:
		if pos == m.end {
			return k(pos)
		}
		return false
	case rWordB, rNotWordB:
		a, b := m.wordAt(pos-1, n.ascii), pos < m.end && m.wordAt(pos, n.ascii)
		if (a != b) == (n.kind == rWordB) {
			return k(pos)
		}
		return false
	case rLook:
		var ok bool
		if n.ahead {
			ok = m.match(n.subs[0], pos, func(int) bool { return true })
		} else {
			if pos-n.width < 0 {
				ok = false
			} else {
				ok = m.match(n.subs[0], pos-n.width, func(e int) bool { return e == pos })
			}
		}
		if ok != n.neg {
			return k(pos)
		}
		return false
	case rAtomic:
		end := -1
		if !m.match(n.subs[0], pos, func(e int) bool { end = e; return true }) {
			return false
		}
		return k(end)
	case rCond:
		if m.caps[n.group*2] >= 0 {
			return m.match(n.subs[0], pos, k)
		}
		if len(n.subs) > 1 {
			return m.match(n.subs[1], pos, k)
		}
		return k(pos)
	case rRepeat:
		return m.repeat(n, pos, k)
	}
	return false
}

func (m *reMatcher) cat(subs []*reNode, i, pos int, k func(int) bool) bool {
	if i == len(subs) {
		return k(pos)
	}
	return m.match(subs[i], pos, func(e int) bool { return m.cat(subs, i+1, e, k) })
}

func (m *reMatcher) repeat(n *reNode, pos int, k func(int) bool) bool {
	sub := n.subs[0]
	// A single-character body is matched in a loop.
	if single(sub) && !n.possess {
		return m.repeatSingle(n, sub, pos, k)
	}
	if n.possess {
		count, cur := 0, pos
		for n.max < 0 || count < n.max {
			end := -1
			if !m.match(sub, cur, func(e int) bool { end = e; return true }) {
				break
			}
			if end == cur {
				count++
				break
			}
			cur = end
			count++
		}
		if count < n.min {
			return false
		}
		return k(cur)
	}
	var try func(count, at int) bool
	try = func(count, at int) bool {
		if n.greedy {
			if n.max < 0 || count < n.max {
				if m.match(sub, at, func(e int) bool {
					if e == at && count >= n.min {
						return false
					}
					return try(count+1, e)
				}) {
					return true
				}
			}
			return count >= n.min && k(at)
		}
		if count >= n.min && k(at) {
			return true
		}
		if n.max < 0 || count < n.max {
			return m.match(sub, at, func(e int) bool {
				if e == at && count >= n.min {
					return false
				}
				return try(count+1, e)
			})
		}
		return false
	}
	return try(0, pos)
}

func single(n *reNode) bool { return n.kind == rLit || n.kind == rAny || n.kind == rSet }

func (m *reMatcher) one(n *reNode, pos int) bool {
	if pos >= m.end {
		return false
	}
	r := m.text[pos]
	switch n.kind {
	case rLit:
		return r == n.r || n.fold && foldEq(r, n.r)
	case rAny:
		return n.dotall || r != '\n'
	}
	return n.set.contains(r)
}

func (m *reMatcher) repeatSingle(n, sub *reNode, pos int, k func(int) bool) bool {
	count := 0
	if n.greedy {
		at := pos
		for (n.max < 0 || count < n.max) && m.one(sub, at) {
			at++
			count++
		}
		for ; count >= n.min; count-- {
			if k(pos + count) {
				return true
			}
		}
		return false
	}
	at := pos
	for count < n.min {
		if !m.one(sub, at) {
			return false
		}
		at++
		count++
	}
	for {
		if k(at) {
			return true
		}
		if (n.max >= 0 && count >= n.max) || !m.one(sub, at) {
			return false
		}
		at++
		count++
	}
}

// ReMatchData is the result of a successful match.
type ReMatchData struct {
	Caps []int // 2 per group, group 0 first; -1 when unset
	Last int   // the group that closed last, 0 for none
}

// matchAt tries to match at exactly pos within [0, end).
func (re *RePattern) matchAt(text []rune, pos, end int, full bool) *ReMatchData {
	return re.matchAtReq(text, pos, end, full, false)
}

// matchAtReq is matchAt, optionally refusing an empty match.
func (re *RePattern) matchAtReq(text []rune, pos, end int, full, nonEmpty bool) *ReMatchData {
	m := &reMatcher{re: re, text: text, end: end, caps: make([]int, 2*(re.Groups+1))}
	for i := range m.caps {
		m.caps[i] = -1
	}
	stop := -1
	if !m.match(re.root, pos, func(e int) bool {
		if full && e != end || nonEmpty && e == pos {
			return false
		}
		stop = e
		return true
	}) {
		return nil
	}
	m.caps[0], m.caps[1] = pos, stop
	return &ReMatchData{Caps: m.caps, Last: m.last}
}

// searchFrom finds the first match at or after pos.
func (re *RePattern) searchFrom(text []rune, pos, end int, mustAdvance bool) *ReMatchData {
	for i := pos; i <= end; i++ {
		if d := re.matchAtReq(text, i, end, false, mustAdvance && i == pos); d != nil {
			return d
		}
	}
	return nil
}

// groupNames lists the names by group number.
func (re *RePattern) groupName(i int) string {
	if i < len(re.names) {
		return re.names[i]
	}
	return ""
}

var _ = fmt.Sprintf

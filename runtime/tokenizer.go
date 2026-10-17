package runtime

// A Python 3.13 tokenizer that gives the tokens of the tokenize module (the C
// tokenizer with extra_tokens=True): the same types, strings, positions in
// characters, and line attributes, f-string tokens included.

import (
	"strings"
	"unicode"
)

// Token types, as in Lib/token.py.
const (
	tokENDMARKER     = 0
	tokNAME          = 1
	tokNUMBER        = 2
	tokSTRING        = 3
	tokNEWLINE       = 4
	tokINDENT        = 5
	tokDEDENT        = 6
	tokOP            = 55
	tokFSTRINGSTART  = 59
	tokFSTRINGMIDDLE = 60
	tokFSTRINGEND    = 61
	tokCOMMENT       = 62
	tokNL            = 63
)

// pyToken is one token: its type and text, where it starts and ends (rows
// from 1, columns in characters) and the source line or lines it is on.
type pyToken struct {
	Type       int
	Str        string
	SRow, SCol int
	ERow, ECol int
	Line       string
}

// tokenizeError is what the tokenizer stops with: a SyntaxError (which the
// tokenize module turns into a TokenError), an IndentationError or a TabError.
type tokenizeError struct {
	Kind   string // "SyntaxError", "IndentationError", "TabError"
	Msg    string
	Row    int
	Offset int // from 1, as SyntaxError.offset
}

// fstringMode is an f-string being read.
type fstringMode struct {
	quote      rune
	quoteSize  int
	raw        bool
	startRow   int // where the FSTRING_START is
	parenBase  int // paren depth when the f-string started
	braceDepth int // replacement fields open
	// Per open replacement field: the paren depth at its '{', and whether
	// its format spec is being read.
	fieldParen []int
	inSpec     []bool
}

type tokenizer struct {
	lines []([]rune) // physical lines, each with its newline
	row   int        // current line index (0-based)
	col   int        // current column in lines[row]

	out []pyToken
	err *tokenizeError

	indents    []int // the indentation stack, in columns (tabs to 8)
	altIndents []int // the same with tabs counted as 1, for TabError
	parens     []rune
	atBOL      bool // at the start of a logical line
	contLine   bool // the previous physical line ended with a backslash

	fstrings []*fstringMode

	// implicitNewline is true when the last line had no newline: the
	// tokenizer reads it as if it had one, which shows in NEWLINE and NL.
	implicitNewline bool
}

func tokenizeSource(src string) ([]pyToken, *tokenizeError) {
	t := &tokenizer{indents: []int{0}, altIndents: []int{0}, atBOL: true}
	for len(src) > 0 {
		i := strings.IndexAny(src, "\n")
		var line string
		if i < 0 {
			line, src = src, ""
		} else {
			line, src = src[:i+1], src[i+1:]
		}
		t.lines = append(t.lines, []rune(line))
	}
	if n := len(t.lines); n > 0 {
		last := t.lines[n-1]
		if len(last) == 0 || last[len(last)-1] != '\n' && last[len(last)-1] != '\r' {
			t.implicitNewline = true
			t.lines[n-1] = append(last, '\n')
		}
	}
	t.run()
	return t.out, t.err
}

// lineText is a physical line as the token's line attribute shows it.
func (t *tokenizer) lineText(row int) string {
	if row < 0 || row >= len(t.lines) {
		return ""
	}
	l := t.lines[row]
	if row == len(t.lines)-1 && t.implicitNewline {
		l = l[:len(l)-1]
	}
	return string(l)
}

// linesText is the lines from..to joined, for multi-line strings.
func (t *tokenizer) linesText(from, to int) string {
	var b strings.Builder
	for r := from; r <= to; r++ {
		b.WriteString(t.lineText(r))
	}
	return b.String()
}

func (t *tokenizer) emit(typ int, s string, srow, scol, erow, ecol int, line string) {
	t.out = append(t.out, pyToken{typ, s, srow + 1, scol, erow + 1, ecol, line})
}

func (t *tokenizer) fail(kind, msg string, row, offset int) {
	if t.err == nil {
		t.err = &tokenizeError{kind, msg, row + 1, offset}
	}
}

func (t *tokenizer) cur() []rune { return t.lines[t.row] }

func (t *tokenizer) peek(k int) rune {
	l := t.cur()
	if t.col+k < len(l) {
		return l[t.col+k]
	}
	return 0
}

func (t *tokenizer) run() {
	for t.err == nil {
		if t.row >= len(t.lines) {
			t.finish()
			return
		}
		if len(t.fstrings) > 0 && t.inFstringLiteral() {
			t.fstringLiteral()
			continue
		}
		if t.col == 0 && t.atBOL && len(t.parens) == 0 && !t.contLine {
			if t.indentation() {
				continue
			}
		}
		t.contLine = false
		t.token()
	}
}

// finish emits the NEWLINE, DEDENTs and ENDMARKER at the end of the input.
func (t *tokenizer) finish() {
	if len(t.parens) > 0 {
		t.eofError(false)
		return
	}
	if len(t.fstrings) > 0 {
		f := t.fstrings[len(t.fstrings)-1]
		if f.quoteSize == 3 {
			t.fail("SyntaxError", "unterminated triple-quoted f-string literal (detected at line "+itoa(len(t.lines))+")", f.startRow, 1)
		} else {
			t.fail("SyntaxError", "unterminated f-string literal (detected at line "+itoa(len(t.lines))+")", f.startRow, 1)
		}
		return
	}
	end := len(t.lines)
	for len(t.indents) > 1 {
		t.indents = t.indents[:len(t.indents)-1]
		t.altIndents = t.altIndents[:len(t.altIndents)-1]
		t.emit(tokDEDENT, "", end, 0, end, 0, "")
	}
	t.emit(tokENDMARKER, "", end, 0, end, 0, "")
}

// eofError is the end of the input inside brackets or after a backslash.
// The column is where CPython's buffer ends, which is the length of the
// line in the cases checked: inside an f-string, or after a backslash on a
// first line that has its newline; 0 otherwise.
func (t *tokenizer) eofError(backslash bool) {
	row := len(t.lines) - 1
	offset := 0
	if len(t.fstrings) > 0 || backslash && row == 0 && !t.implicitNewline {
		offset = len(t.lines[row])
	}
	t.fail("SyntaxError", "unexpected EOF in multi-line statement", row, offset)
}

// indentation reads the leading whitespace of a logical line. A blank or
// comment-only line gives no INDENT or DEDENT. It reports whether the line
// was consumed (a blank line).
func (t *tokenizer) indentation() bool {
	l := t.cur()
	col, alt, i := 0, 0, 0
	for ; i < len(l); i++ {
		switch l[i] {
		case ' ':
			col++
			alt++
		case '\t':
			col = (col/8 + 1) * 8
			alt++
		case '\f':
			col, alt = 0, 0
		default:
			goto done
		}
	}
done:
	c := rune(0)
	if i < len(l) {
		c = l[i]
	}
	if c == '#' || c == '\n' || c == '\r' || c == 0 {
		// A blank line: the comment and NL only.
		t.col = i
		if c == '#' {
			t.comment()
		}
		t.newlineToken(tokNL)
		return true
	}
	top := t.indents[len(t.indents)-1]
	altTop := t.altIndents[len(t.altIndents)-1]
	switch {
	case col == top:
		if alt != altTop {
			t.fail("TabError", "inconsistent use of tabs and spaces in indentation", t.row, i+1)
			return true
		}
	case col > top:
		if alt <= altTop {
			t.fail("TabError", "inconsistent use of tabs and spaces in indentation", t.row, i+1)
			return true
		}
		t.indents = append(t.indents, col)
		t.altIndents = append(t.altIndents, alt)
		t.emit(tokINDENT, string(l[:i]), t.row, 0, t.row, i, t.lineText(t.row))
	default:
		n := len(t.indents)
		for n > 1 && col < t.indents[n-1] {
			n--
		}
		if col != t.indents[n-1] {
			t.fail("IndentationError", "unindent does not match any outer indentation level", t.row, i+1)
			return true
		}
		if alt != t.altIndents[n-1] {
			t.fail("TabError", "inconsistent use of tabs and spaces in indentation", t.row, i+1)
			return true
		}
		for len(t.indents) > n {
			t.indents = t.indents[:len(t.indents)-1]
			t.altIndents = t.altIndents[:len(t.altIndents)-1]
			t.emit(tokDEDENT, "", t.row, i, t.row, i, t.lineText(t.row))
		}
	}
	t.col = i
	t.atBOL = false
	return false
}

// newlineToken emits NEWLINE or NL for the line end at t.col and moves to
// the next line.
func (t *tokenizer) newlineToken(typ int) {
	l := t.cur()
	s := string(l[t.col:])
	if t.row == len(t.lines)-1 && t.implicitNewline {
		s = ""
	}
	t.emit(typ, s, t.row, t.col, t.row, t.col+1, t.lineText(t.row))
	t.row++
	t.col = 0
	if typ == tokNEWLINE {
		t.atBOL = true
	}
}

func (t *tokenizer) comment() {
	l := t.cur()
	start := t.col
	end := start
	for end < len(l) && l[end] != '\n' && l[end] != '\r' {
		end++
	}
	t.emit(tokCOMMENT, string(l[start:end]), t.row, start, t.row, end, t.lineText(t.row))
	t.col = end
}

func isIDStart(c rune) bool {
	return c == '_' || unicode.IsLetter(c) || unicode.Is(unicode.Nl, c) || unicode.Is(unicode.Other_ID_Start, c)
}

func isIDContinue(c rune) bool {
	return isIDStart(c) || unicode.IsDigit(c) || unicode.In(c, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc) || unicode.Is(unicode.Other_ID_Continue, c)
}

// token reads one token at t.col of a line that is not at its indentation.
func (t *tokenizer) token() {
	l := t.cur()
	// Skip spaces.
	for t.col < len(l) && (l[t.col] == ' ' || l[t.col] == '\t' || l[t.col] == '\f') {
		t.col++
	}
	if t.col >= len(l) {
		t.row++
		t.col = 0
		return
	}
	c := l[t.col]
	start := t.col
	switch {
	case c == '#':
		t.comment()
		return
	case c == '\n' || c == '\r' && t.peek(1) == '\n' || c == '\r':
		if len(t.parens) > 0 || t.atBOL {
			t.newlineToken(tokNL)
		} else {
			t.newlineToken(tokNEWLINE)
		}
		return
	case c == '\\':
		if t.peek(1) == '\n' || t.peek(1) == '\r' {
			t.row++
			t.col = 0
			t.contLine = true
			if t.row >= len(t.lines) {
				t.eofError(true)
			}
			return
		}
		t.fail("SyntaxError", "unexpected character after line continuation character", t.row, len(l))
		return
	case isIDStart(c) || c >= 128 && !unicode.IsSpace(c) && !unicode.IsPunct(c) && !unicode.IsSymbol(c):
		// A string prefix?
		i := t.col
		sawB, sawR, sawU, sawF := false, false, false, false
		for i < len(l) {
			ch := l[i]
			switch {
			case !(sawB || sawU || sawF) && (ch == 'b' || ch == 'B'):
				sawB = true
			case !(sawB || sawU || sawR || sawF) && (ch == 'u' || ch == 'U'):
				sawU = true
			case !(sawR || sawU) && (ch == 'r' || ch == 'R'):
				sawR = true
			case !(sawF || sawB || sawU) && (ch == 'f' || ch == 'F'):
				sawF = true
			default:
				goto prefixDone
			}
			i++
			if i < len(l) && (l[i] == '"' || l[i] == '\'') {
				if sawF {
					t.fstringStart(start, i, sawR)
				} else {
					t.stringLit(start, i)
				}
				return
			}
		}
	prefixDone:
		i = t.col
		for i < len(l) && (isIDContinue(l[i]) || l[i] >= 128 && !unicode.IsSpace(l[i]) && !unicode.IsPunct(l[i]) && !unicode.IsSymbol(l[i])) {
			i++
		}
		t.emit(tokNAME, string(l[start:i]), t.row, start, t.row, i, t.lineText(t.row))
		t.col = i
		t.atBOL = false
		return
	case c >= '0' && c <= '9' || c == '.' && t.peek(1) >= '0' && t.peek(1) <= '9':
		t.number()
		return
	case c == '"' || c == '\'':
		t.stringLit(start, start)
		return
	}
	t.operator()
}

func (t *tokenizer) number() {
	l := t.cur()
	start := t.col
	i := start
	digits := func(ok func(rune) bool) bool {
		// digits with single underscores between them
		if i >= len(l) || !ok(l[i]) {
			return false
		}
		for i < len(l) {
			if ok(l[i]) {
				i++
			} else if l[i] == '_' && i+1 < len(l) && ok(l[i+1]) {
				i += 2
			} else {
				break
			}
		}
		return true
	}
	dec := func(c rune) bool { return c >= '0' && c <= '9' }
	bad := func() {
		t.fail("SyntaxError", "invalid decimal literal", t.row, i+1)
	}
	if l[i] == '0' && i+1 < len(l) && strings.ContainsRune("xXoObB", l[i+1]) {
		kind := l[i+1]
		i += 2
		var ok func(rune) bool
		switch kind {
		case 'x', 'X':
			ok = func(c rune) bool { return dec(c) || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' }
		case 'o', 'O':
			ok = func(c rune) bool { return c >= '0' && c <= '7' }
		default:
			ok = func(c rune) bool { return c == '0' || c == '1' }
		}
		name := map[rune]string{'x': "hexadecimal", 'X': "hexadecimal", 'o': "octal", 'O': "octal", 'b': "binary", 'B': "binary"}[kind]
		// The column of an error is the character that stops the literal.
		if i < len(l) && l[i] == '_' {
			i++
		}
		if !digits(ok) {
			if name != "hexadecimal" && i < len(l) && dec(l[i]) {
				t.fail("SyntaxError", "invalid digit '"+string(l[i])+"' in "+name+" literal", t.row, i+1)
				return
			}
			t.fail("SyntaxError", "invalid "+name+" literal", t.row, i)
			return
		}
		if i < len(l) && l[i] == '_' {
			t.fail("SyntaxError", "invalid "+name+" literal", t.row, i+1)
			return
		}
		if name != "hexadecimal" && i < len(l) && dec(l[i]) {
			t.fail("SyntaxError", "invalid digit '"+string(l[i])+"' in "+name+" literal", t.row, i+1)
			return
		}
	} else {
		if l[i] != '.' {
			digits(dec)
			if i < len(l) && l[i] == '_' {
				bad()
				return
			}
		}
		if i < len(l) && l[i] == '.' {
			i++
			digits(dec)
			if i < len(l) && l[i] == '_' {
				bad()
				return
			}
		}
		if i < len(l) && (l[i] == 'e' || l[i] == 'E') {
			j := i + 1
			if j < len(l) && (l[j] == '+' || l[j] == '-') {
				j++
			}
			if j < len(l) && dec(l[j]) {
				i = j
				digits(dec)
				if i < len(l) && l[i] == '_' {
					bad()
					return
				}
			} else if j > i+1 {
				t.fail("SyntaxError", "invalid decimal literal", t.row, j)
				return
			}
		}
		if i < len(l) && (l[i] == 'j' || l[i] == 'J') {
			i++
		}
	}
	t.emit(tokNUMBER, string(l[start:i]), t.row, start, t.row, i, t.lineText(t.row))
	t.col = i
	t.atBOL = false
}

// stringLit reads a string literal whose prefix starts at start and whose
// quote is at q.
func (t *tokenizer) stringLit(start, q int) {
	l := t.cur()
	quote := l[q]
	size := 1
	if q+2 < len(l) && l[q+1] == quote && l[q+2] == quote {
		size = 3
	}
	srow := t.row
	i := q + size
	row := t.row
	for {
		line := t.lines[row]
		if i >= len(line) {
			if size == 1 {
				t.fail("SyntaxError", "unterminated string literal (detected at line "+itoa(row+1)+")", srow, q+1)
				return
			}
			row++
			i = 0
			if row >= len(t.lines) {
				t.fail("SyntaxError", "unterminated triple-quoted string literal (detected at line "+itoa(len(t.lines))+")", srow, q+1)
				return
			}
			continue
		}
		c := line[i]
		if c == '\\' {
			if i+1 < len(line) && (line[i+1] == '\n' || line[i+1] == '\r' && i+2 >= len(line)) && row+1 < len(t.lines) {
				// A backslash at the end of the line continues the string.
				row++
				i = 0
				continue
			}
			if i+2 < len(line) && line[i+1] == '\r' && line[i+2] == '\n' && row+1 < len(t.lines) {
				row++
				i = 0
				continue
			}
			if i+1 < len(line) {
				i += 2
				continue
			}
			i++
			continue
		}
		if c == '\n' && size == 1 {
			t.fail("SyntaxError", "unterminated string literal (detected at line "+itoa(row+1)+")", srow, q+1)
			return
		}
		if c == quote && (size == 1 || i+2 < len(line) && line[i+1] == quote && line[i+2] == quote) {
			i += size
			break
		}
		i++
	}
	var text string
	if row == srow {
		text = string(l[start:i])
	} else {
		var b strings.Builder
		b.WriteString(string(l[start:]))
		for r := srow + 1; r < row; r++ {
			b.WriteString(string(t.lines[r]))
		}
		b.WriteString(string(t.lines[row][:i]))
		text = b.String()
	}
	t.emit(tokSTRING, text, srow, start, row, i, t.linesText(srow, row))
	t.row, t.col = row, i
	t.atBOL = false
}

var threeCharOps = []string{"**=", "...", "//=", ">>=", "<<=", "!="}
var twoCharOps = []string{"!=", "%=", "&=", "**", "*=", "+=", "-=", "->", "//", "/=", ":=", "<<", "<=", "<>", "==", ">=", ">>", "@=", "^=", "|="}

const oneCharOps = "%&()*+,-./:;<=>@[]^{|}~!"

func (t *tokenizer) operator() {
	l := t.cur()
	start := t.col
	rest := string(l[start:min(start+3, len(l))])
	op := ""
	for _, o := range threeCharOps {
		if strings.HasPrefix(rest, o) {
			op = o
			break
		}
	}
	if op == "" {
		for _, o := range twoCharOps {
			if strings.HasPrefix(rest, o) {
				op = o
				break
			}
		}
	}
	c := l[start]
	if op == "" && strings.ContainsRune(oneCharOps, c) {
		op = string(c)
	}
	if op == "" {
		// The tokenize module lets any other character through as OP.
		op = string(c)
	}
	// At the top of a replacement field of an f-string, ':' starts the
	// format spec and '}' ends the field.
	if f := t.topFstring(); f != nil && f.braceDepth > 0 && len(t.parens) == f.fieldParen[f.braceDepth-1] {
		switch c {
		case ':':
			t.emit(tokOP, ":", t.row, start, t.row, start+1, t.lineText(t.row))
			t.col = start + 1
			f.inSpec[f.braceDepth-1] = true
			return
		case '}':
			t.emit(tokOP, "}", t.row, start, t.row, start+1, t.lineText(t.row))
			t.col = start + 1
			t.closeField(f)
			return
		}
	}
	switch op {
	case "(", "[", "{":
		t.parens = append(t.parens, c)
	case ")", "]", "}":
		// The tokenize module does not check that brackets match.
		if len(t.parens) > 0 {
			t.parens = t.parens[:len(t.parens)-1]
		}
	}
	end := start + len([]rune(op))
	t.emit(tokOP, op, t.row, start, t.row, end, t.lineText(t.row))
	t.col = end
	t.atBOL = false
}

// openField starts a replacement field at its '{'.
func (t *tokenizer) openField(f *fstringMode) {
	t.parens = append(t.parens, '{')
	f.braceDepth++
	f.fieldParen = append(f.fieldParen, len(t.parens))
	f.inSpec = append(f.inSpec, false)
}

// closeField ends a replacement field at its '}'.
func (t *tokenizer) closeField(f *fstringMode) {
	t.parens = t.parens[:len(t.parens)-1]
	f.braceDepth--
	f.fieldParen = f.fieldParen[:f.braceDepth]
	f.inSpec = f.inSpec[:f.braceDepth]
}

func (t *tokenizer) topFstring() *fstringMode {
	if len(t.fstrings) == 0 {
		return nil
	}
	return t.fstrings[len(t.fstrings)-1]
}

// inFstringLiteral reports whether the innermost f-string is reading text:
// outside a replacement field, or in a format spec.
func (t *tokenizer) inFstringLiteral() bool {
	f := t.topFstring()
	return f.braceDepth == 0 || f.inSpec[f.braceDepth-1] && len(t.parens) == f.fieldParen[f.braceDepth-1]
}

func (t *tokenizer) fstringStart(start, q int, raw bool) {
	l := t.cur()
	quote := l[q]
	size := 1
	if q+2 < len(l) && l[q+1] == quote && l[q+2] == quote {
		size = 3
	}
	end := q + size
	t.emit(tokFSTRINGSTART, string(l[start:end]), t.row, start, t.row, end, t.lineText(t.row))
	t.fstrings = append(t.fstrings, &fstringMode{quote: quote, quoteSize: size, raw: raw, startRow: t.row, parenBase: len(t.parens)})
	t.col = end
	t.atBOL = false
}

// fstringLiteral reads the text of an f-string up to a replacement field,
// the end of a format spec or the closing quote.
func (t *tokenizer) fstringLiteral() {
	f := t.topFstring()
	inSpec := f.braceDepth > 0
	srow, scol := t.row, t.col
	var text strings.Builder
	row, i := t.row, t.col
	// An empty text gives a token only where a format spec ends.
	emitMiddle := func(erow, ecol int, specEnd bool) {
		s := text.String()
		if s != "" || specEnd {
			t.emit(tokFSTRINGMIDDLE, s, srow, scol, erow, ecol, t.linesText(srow, erow))
		}
		text.Reset()
	}
	for {
		if row >= len(t.lines) {
			t.row = row
			t.finish()
			return
		}
		line := t.lines[row]
		if i >= len(line) {
			row++
			i = 0
			continue
		}
		c := line[i]
		// The closing quote.
		if !inSpec && c == f.quote && (f.quoteSize == 1 || i+2 < len(line) && line[i+1] == f.quote && line[i+2] == f.quote) {
			if text.Len() > 0 {
				emitMiddle(row, i, false)
			}
			t.emit(tokFSTRINGEND, strings.Repeat(string(f.quote), f.quoteSize), row, i, row, i+f.quoteSize, t.lineText(row))
			t.fstrings = t.fstrings[:len(t.fstrings)-1]
			t.row, t.col = row, i+f.quoteSize
			return
		}
		if c == '\n' && f.quoteSize == 1 {
			if inSpec {
				t.fail("SyntaxError", "f-string: newlines are not allowed in format specifiers for single quoted f-strings", row, i+1)
			} else {
				t.fail("SyntaxError", "unterminated f-string literal (detected at line "+itoa(row+1)+")", f.startRow, 1)
			}
			return
		}
		if c == '{' {
			if !inSpec && i+1 < len(line) && line[i+1] == '{' {
				text.WriteRune('{')
				emitMiddle(row, i+1, false)
				i += 2
				srow, scol = row, i
				continue
			}
			emitMiddle(row, i, false)
			t.row, t.col = row, i
			t.emit(tokOP, "{", row, i, row, i+1, t.lineText(row))
			t.col = i + 1
			t.openField(f)
			return
		}
		if c == '}' {
			if !inSpec && i+1 < len(line) && line[i+1] == '}' {
				text.WriteRune('}')
				emitMiddle(row, i+1, false)
				i += 2
				srow, scol = row, i
				continue
			}
			if inSpec {
				emitMiddle(row, i, true)
				t.row, t.col = row, i
				t.emit(tokOP, "}", row, i, row, i+1, t.lineText(row))
				t.col = i + 1
				t.closeField(f)
				return
			}
			emitMiddle(row, i, true)
			t.fail("SyntaxError", "f-string: single '}' is not allowed", row, i+1)
			return
		}
		if c == '\\' && !f.raw && i+1 < len(line) {
			if line[i+1] == 'N' && i+2 < len(line) && line[i+2] == '{' {
				j := i + 3
				for j < len(line) && line[j] != '}' {
					j++
				}
				text.WriteString(string(line[i:min(j+1, len(line))]))
				i = j + 1
				continue
			}
			text.WriteRune(c)
			text.WriteRune(line[i+1])
			i += 2
			continue
		}
		text.WriteRune(c)
		i++
	}
}

func init() {
	// tokenize.tokens(source) is the tokens of the source, as tuples
	// (type, string, (srow, scol), (erow, ecol), line), and the error the
	// tokenizer stopped with: (kind, msg, lineno, offset) or None.
	RegisterNative("tokenize.tokens", func(args []any, kw []KW) any {
		nativeArgs("tokens", args, 1, 1)
		toks, err := tokenizeSource(Str(args[0]))
		out := NewList[any]()
		for _, t := range toks {
			out.Append(NewVarTuple[any](int64(t.Type), t.Str,
				NewVarTuple[any](int64(t.SRow), int64(t.SCol)),
				NewVarTuple[any](int64(t.ERow), int64(t.ECol)), t.Line))
		}
		var e any
		if err != nil {
			e = NewVarTuple[any](err.Kind, err.Msg, int64(err.Row), int64(err.Offset))
		}
		return NewVarTuple[any](out, e)
	})
}

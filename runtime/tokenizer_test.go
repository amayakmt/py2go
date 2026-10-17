package runtime

import (
	"fmt"
	"strings"
	"testing"
)

// The expected tokens are CPython 3.13's tokenize output.
func TestTokenizerMatchesCPython(t *testing.T) {
	cases := []struct{ src, want string }{
		{"x = f\"a{b!r:>{w}}c{{d}}\"\n", `1 'x' 1,0 1,1|55 '=' 1,2 1,3|59 'f"' 1,4 1,6|60 'a' 1,6 1,7|55 '{' 1,7 1,8|1 'b' 1,8 1,9|55 '!' 1,9 1,10|1 'r' 1,10 1,11|55 ':' 1,11 1,12|60 '>' 1,12 1,13|55 '{' 1,13 1,14|1 'w' 1,14 1,15|55 '}' 1,15 1,16|60 '' 1,16 1,16|55 '}' 1,16 1,17|60 'c{' 1,17 1,19|60 'd}' 1,20 1,22|61 '"' 1,23 1,24|4 '\n' 1,24 1,25|0 '' 2,0 2,0`},
		{"if x:\n\ty = 1  # c\n", `1 'if' 1,0 1,2|1 'x' 1,3 1,4|55 ':' 1,4 1,5|4 '\n' 1,5 1,6|5 '\t' 2,0 2,1|1 'y' 2,1 2,2|55 '=' 2,3 2,4|2 '1' 2,5 2,6|62 '# c' 2,8 2,11|4 '\n' 2,11 2,12|6 '' 3,0 3,0|0 '' 3,0 3,0`},
		{"z = (1,\n     2)", `1 'z' 1,0 1,1|55 '=' 1,2 1,3|55 '(' 1,4 1,5|2 '1' 1,5 1,6|55 ',' 1,6 1,7|63 '\n' 1,7 1,8|2 '2' 2,5 2,6|55 ')' 2,6 2,7|4 '' 2,7 2,8|0 '' 3,0 3,0`},
		{"s = \"\"\"a\nb\"\"\"\n", `1 's' 1,0 1,1|55 '=' 1,2 1,3|3 '"""a\nb"""' 1,4 2,4|4 '\n' 2,4 2,5|0 '' 3,0 3,0`},
	}
	for _, c := range cases {
		toks, err := tokenizeSource(c.src)
		if err != nil {
			t.Errorf("%q: %v", c.src, err.Msg)
			continue
		}
		var got []string
		for _, tk := range toks {
			got = append(got, fmt.Sprintf("%d %s %d,%d %d,%d", tk.Type, Repr(tk.Str), tk.SRow, tk.SCol, tk.ERow, tk.ECol))
		}
		if g := strings.Join(got, "|"); g != c.want {
			t.Errorf("%q:\n got %s\nwant %s", c.src, g, c.want)
		}
	}
}

func TestTokenizerErrors(t *testing.T) {
	cases := []struct {
		src, kind, msg string
		row, offset    int
	}{
		{"x = \"abc", "SyntaxError", "unterminated string literal (detected at line 1)", 1, 5},
		{"x = (1,", "SyntaxError", "unexpected EOF in multi-line statement", 1, 0},
		{"if x:\n  y\n z", "IndentationError", "unindent does not match any outer indentation level", 3, 2},
		{"0b2", "SyntaxError", "invalid digit '2' in binary literal", 1, 3},
		{"f\"}\"", "SyntaxError", "f-string: single '}' is not allowed", 1, 3},
	}
	for _, c := range cases {
		_, err := tokenizeSource(c.src)
		if err == nil || err.Kind != c.kind || err.Msg != c.msg || err.Row != c.row || err.Offset != c.offset {
			t.Errorf("%q: got %+v, want %s %q at %d,%d", c.src, err, c.kind, c.msg, c.row, c.offset)
		}
	}
}

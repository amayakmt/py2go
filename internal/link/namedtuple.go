package link

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// A call of collections.namedtuple with a literal name and literal field
// names is replaced by the class it would make. The classes cannot be made
// when the program runs, so the shape of the class is written out here.

var identifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// expandNamedTuple turns "P = namedtuple('P', 'x y')" into a class statement.
func (r *rewriter) expandNamedTuple(as *pyast.Assign) (pyast.Node, bool, error) {
	if len(as.Targets) != 1 {
		return nil, false, nil
	}
	target, ok := as.Targets[0].(*pyast.Name)
	if !ok {
		return nil, false, nil
	}
	call, ok := as.Value.(*pyast.Call)
	if !ok {
		return nil, false, nil
	}
	fn, ok := call.Func.(*pyast.Name)
	if !ok || fn.ID != "collections"+Marker+"namedtuple" {
		return nil, false, nil
	}
	line := as.Lineno
	fail := func(msg string) (pyast.Node, bool, error) {
		return nil, true, fmt.Errorf("unsupported: namedtuple %s at line %d", msg, line)
	}
	if len(call.Args) < 2 {
		return fail("needs a literal name and field names")
	}
	typename, ok := strConst(call.Args[0])
	if !ok {
		return fail("needs a literal type name")
	}
	var fields []string
	switch f := call.Args[1].(type) {
	case *pyast.Constant:
		s, isStr := f.Value.(string)
		if !isStr {
			return fail("needs literal field names")
		}
		fields = strings.Fields(strings.ReplaceAll(s, ",", " "))
	case *pyast.List, *pyast.Tuple:
		fields = append(fields, strConsts(f)...)
		var elts []pyast.Node
		if l, ok := f.(*pyast.List); ok {
			elts = l.Elts
		} else {
			elts = f.(*pyast.Tuple).Elts
		}
		if len(fields) != len(elts) {
			return fail("needs literal field names")
		}
	default:
		return fail("needs literal field names")
	}
	var defaults []pyast.Node
	for _, k := range call.Keywords {
		switch k.Arg {
		case "defaults":
			switch d := k.Value.(type) {
			case *pyast.Tuple:
				defaults = d.Elts
			case *pyast.List:
				defaults = d.Elts
			default:
				return fail("needs literal defaults")
			}
		case "rename", "module":
		default:
			return fail("keyword '" + k.Arg + "'")
		}
	}
	seen := map[string]bool{}
	for _, f := range fields {
		if !identifier.MatchString(f) || strings.HasPrefix(f, "_") || seen[f] || isKeyword(f) {
			return nil, true, fmt.Errorf("ValueError: field names must be valid identifiers: %q", f)
		}
		seen[f] = true
	}
	if len(defaults) > len(fields) {
		return nil, true, fmt.Errorf("TypeError: Got more default values than field names")
	}
	cls, err := namedTupleClass(typename, target.ID, fields, line)
	if err != nil {
		return nil, true, err
	}
	// The defaults belong to the last parameters of __init__.
	for _, st := range cls.Body {
		if fd, ok := st.(*pyast.FunctionDef); ok && fd.Name == "__init__" {
			fd.Args.Defaults = defaults
		}
	}
	return cls, true, nil
}

func strConst(n pyast.Node) (string, bool) {
	c, ok := n.(*pyast.Constant)
	if !ok {
		return "", false
	}
	s, ok := c.Value.(string)
	return s, ok
}

var pyKeywords = map[string]bool{
	"False": true, "None": true, "True": true, "and": true, "as": true, "assert": true, "async": true, "await": true,
	"break": true, "class": true, "continue": true, "def": true, "del": true, "elif": true, "else": true,
	"except": true, "finally": true, "for": true, "from": true, "global": true, "if": true, "import": true,
	"in": true, "is": true, "lambda": true, "nonlocal": true, "not": true, "or": true, "pass": true,
	"raise": true, "return": true, "try": true, "while": true, "with": true, "yield": true,
}

func isKeyword(s string) bool { return pyKeywords[s] }

// namedTupleClass writes the class and parses it.
func namedTupleClass(typename, name string, fields []string, line int) (*pyast.ClassDef, error) {
	var b strings.Builder
	n := len(fields)
	tup := "(" + strings.Join(prefixed("self.", fields), ", ") + ",)"
	if n == 0 {
		tup = "()"
	}
	q := func(s string) string { return fmt.Sprintf("%q", s) }
	fmt.Fprintf(&b, "class %s:\n", name)
	fmt.Fprintf(&b, "    _fields = (%s)\n", quotedList(fields, q))
	fmt.Fprintf(&b, "    def __init__(self, %s):\n", strings.Join(fields, ", "))
	if n == 0 {
		b.WriteString("        pass\n")
	}
	for _, f := range fields {
		fmt.Fprintf(&b, "        self.%s = %s\n", f, f)
	}
	b.WriteString("    def __repr__(self):\n")
	parts := make([]string, n)
	for i, f := range fields {
		parts[i] = fmt.Sprintf("%s + repr(self.%s)", q(f+"="), f)
	}
	body := q(typename+"(") + " + "
	if n > 0 {
		body += strings.Join(parts, " + \", \" + ") + " + "
	}
	body += q(")")
	fmt.Fprintf(&b, "        return %s\n", body)
	fmt.Fprintf(&b, "    def _tuple(self):\n        return %s\n", tup)
	b.WriteString("    def __iter__(self):\n        return iter(self._tuple())\n")
	fmt.Fprintf(&b, "    def __len__(self):\n        return %d\n", n)
	b.WriteString("    def __getitem__(self, i):\n        return self._tuple()[i]\n")
	b.WriteString("    def __contains__(self, x):\n        return x in self._tuple()\n")
	b.WriteString("    def __eq__(self, other):\n        return self._tuple() == tuple(other)\n")
	b.WriteString("    def __lt__(self, other):\n        return self._tuple() < tuple(other)\n")
	b.WriteString("    def __le__(self, other):\n        return self._tuple() <= tuple(other)\n")
	b.WriteString("    def __gt__(self, other):\n        return self._tuple() > tuple(other)\n")
	b.WriteString("    def __ge__(self, other):\n        return self._tuple() >= tuple(other)\n")
	b.WriteString("    def __hash__(self):\n        return hash(self._tuple())\n")
	b.WriteString("    def count(self, x):\n        return self._tuple().count(x)\n")
	b.WriteString("    def index(self, x):\n        return self._tuple().index(x)\n")
	dictItems := make([]string, n)
	replItems := make([]string, n)
	for i, f := range fields {
		dictItems[i] = fmt.Sprintf("%s: self.%s", q(f), f)
		replItems[i] = fmt.Sprintf("kwds.get(%s, self.%s)", q(f), f)
	}
	fmt.Fprintf(&b, "    def _asdict(self):\n        return {%s}\n", strings.Join(dictItems, ", "))
	fmt.Fprintf(&b, "    def _replace(self, **kwds):\n        return %s(%s)\n", name, strings.Join(replItems, ", "))
	fmt.Fprintf(&b, "    @classmethod\n    def _make(cls, iterable):\n        return cls(*iterable)\n")

	dir, err := os.MkdirTemp("", "py2go-nt-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "nt.py")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return nil, err
	}
	mod, err := parseFile(path)
	if err != nil {
		return nil, fmt.Errorf("namedtuple %q at line %d: %w", typename, line, err)
	}
	cls, ok := mod.Body[0].(*pyast.ClassDef)
	if !ok {
		return nil, fmt.Errorf("namedtuple %q at line %d: no class", typename, line)
	}
	return cls, nil
}

func prefixed(p string, xs []string) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = p + x
	}
	return out
}

func quotedList(xs []string, q func(string) string) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = q(x)
	}
	if len(xs) == 1 {
		return parts[0] + ","
	}
	return strings.Join(parts, ", ")
}

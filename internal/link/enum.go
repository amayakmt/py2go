package link

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// Classes that derive from enum.Enum are rewritten: a class cannot make its
// members when the program runs, so each member becomes a variable that the
// code after the class assigns, and the uses of the class (Color.RED,
// Color(1), Color["RED"], iteration over Color) become uses of those
// variables and of functions written here.

type enumMember struct {
	name  string
	value pyast.Node
	auto  bool
}

type enumInfo struct {
	cls         string // the name of the class in the joined module
	kind        string // Enum, IntEnum, StrEnum, Flag or IntFlag
	members     []*enumMember
	byName      map[string]*enumMember
	allDeclared []string // every name, aliases included, in order
}

var enumBases = map[string]bool{"Enum": true, "IntEnum": true, "StrEnum": true, "Flag": true, "IntFlag": true}

func baseIdent(n pyast.Node) string {
	switch x := n.(type) {
	case *pyast.Name:
		return x.ID
	case *pyast.Attribute:
		return x.Attr
	}
	return ""
}

// scanEnums finds the enum classes of a module before anything is rewritten,
// so that a use that comes before the class is still understood.
func (l *linker) scanEnums(m *module) {
	local := map[string]string{} // class name -> kind, in this module
	for _, st := range m.ast.Body {
		cd, ok := st.(*pyast.ClassDef)
		if !ok {
			continue
		}
		kind := ""
		for _, b := range cd.Bases {
			id := baseIdent(b)
			if enumBases[id] {
				kind = id
			} else if k, ok := local[id]; ok {
				kind = k
			} else if info := l.enums[m.prefix+id]; info != nil {
				kind = info.kind
			}
		}
		if kind == "" {
			continue
		}
		local[cd.Name] = kind
		info := &enumInfo{cls: m.prefix + cd.Name, kind: kind, byName: map[string]*enumMember{}}
		for _, s := range cd.Body {
			as, ok := s.(*pyast.Assign)
			if !ok || len(as.Targets) != 1 {
				continue
			}
			nm, ok := as.Targets[0].(*pyast.Name)
			if !ok || strings.HasPrefix(nm.ID, "_") {
				continue
			}
			if _, isLambda := as.Value.(*pyast.Lambda); isLambda {
				continue
			}
			mem := &enumMember{name: nm.ID, value: as.Value}
			if call, ok := as.Value.(*pyast.Call); ok && baseIdent(call.Func) == "auto" {
				mem.auto = true
			}
			info.members = append(info.members, mem)
			info.byName[mem.name] = mem
			info.allDeclared = append(info.allDeclared, mem.name)
		}
		l.enums[info.cls] = info
	}
}

// enumClass rewrites the class statement of an enum, and returns it with the
// statements that follow it.
func (r *rewriter) enumClass(cd *pyast.ClassDef, info *enumInfo) (pyast.Node, error) {
	// The member assignments leave the body.
	var body []pyast.Node
	hasInit := false
	for _, st := range cd.Body {
		if as, ok := st.(*pyast.Assign); ok && len(as.Targets) == 1 {
			if nm, ok := as.Targets[0].(*pyast.Name); ok && info.byName[nm.ID] != nil {
				continue
			}
		}
		if fd, ok := st.(*pyast.FunctionDef); ok && fd.Name == "__init__" {
			hasInit = true
		}
		body = append(body, st)
	}
	cls := info.cls
	line := cd.Lineno
	flag := info.kind == "Flag" || info.kind == "IntFlag"
	generated, err := parseSnippet(enumHelperSource(cls, info, flag))
	if err != nil {
		return nil, fmt.Errorf("enum %s at line %d: %w", cd.Name, line, err)
	}
	// The helper methods go into the class, the functions after it.
	var after []pyast.Node
	for _, st := range generated.Body {
		switch x := st.(type) {
		case *pyast.ClassDef:
			body = append(body, x.Body...)
		default:
			after = append(after, st)
		}
	}
	cd.Body = body
	out := []pyast.Node{cd}

	// Values: auto() counts on, a repeated value is an alias.
	var members []*enumMember
	seen := map[string]*enumMember{}
	var nextInt int64 = 1
	maxFlag := int64(0)
	var assigned []pyast.Node
	for _, mem := range info.members {
		value := mem.value
		if mem.auto {
			switch info.kind {
			case "StrEnum":
				value = &pyast.Constant{PyType: "str", Value: strings.ToLower(mem.name), Lineno: line}
			case "Flag", "IntFlag":
				v := int64(1)
				if maxFlag > 0 {
					v = maxFlag << 1
				}
				value = &pyast.Constant{PyType: "int", Value: v, Lineno: line}
			default:
				value = &pyast.Constant{PyType: "int", Value: nextInt, Lineno: line}
			}
		}
		mem.value = value
		if c, ok := value.(*pyast.Constant); ok {
			if v, isInt := c.Value.(int64); isInt {
				nextInt = v + 1
				if v > maxFlag && flag {
					maxFlag = v
					// the next auto flag is the next power of two above
					for p := int64(1); p <= v; p <<= 1 {
						maxFlag = p
					}
				}
			}
			key := fmt.Sprintf("%s|%v", c.PyType, c.Value)
			if first := seen[key]; first != nil {
				assigned = append(assigned, &pyast.Assign{
					Targets: []pyast.Node{&pyast.Name{ID: cls + "__" + mem.name, Lineno: line}},
					Value:   &pyast.Name{ID: cls + "__" + first.name, Lineno: line},
					Lineno:  line,
				})
				info.byName[mem.name] = first
				continue
			}
			seen[key] = mem
		}
		members = append(members, mem)
		var args []pyast.Node
		if hasInit {
			if t, ok := value.(*pyast.Tuple); ok {
				args = []pyast.Node{&pyast.Starred{Value: t, Lineno: line}}
			} else {
				args = []pyast.Node{value}
			}
		}
		varName := cls + "__" + mem.name
		set := func(attr string, v pyast.Node) pyast.Node {
			return &pyast.Assign{
				Targets: []pyast.Node{&pyast.Attribute{Value: &pyast.Name{ID: varName, Lineno: line}, Attr: attr, Lineno: line}},
				Value:   v,
				Lineno:  line,
			}
		}
		assigned = append(assigned,
			&pyast.Assign{
				Targets: []pyast.Node{&pyast.Name{ID: varName, Lineno: line}},
				Value:   &pyast.Call{Func: &pyast.Name{ID: cd.Name, Lineno: line}, Args: args, Lineno: line},
				Lineno:  line,
			},
			set("_value_", value),
			set("_name_", &pyast.Constant{PyType: "str", Value: mem.name, Lineno: line}),
		)
	}
	info.members = members
	var names, pairs []pyast.Node
	for _, mem := range members {
		names = append(names, &pyast.Name{ID: cls + "__" + mem.name, Lineno: line})
	}
	var keys, vals []pyast.Node
	for _, mem := range info.orderedNames() {
		keys = append(keys, &pyast.Constant{PyType: "str", Value: mem, Lineno: line})
		vals = append(vals, &pyast.Name{ID: cls + "__" + info.byName[mem].name, Lineno: line})
	}
	_ = pairs
	assigned = append(assigned,
		&pyast.Assign{Targets: []pyast.Node{&pyast.Name{ID: cls + "__members", Lineno: line}}, Value: &pyast.List{Elts: names, Lineno: line}, Lineno: line},
		&pyast.Assign{Targets: []pyast.Node{&pyast.Name{ID: cls + "__byname", Lineno: line}}, Value: &pyast.Dict{Keys: keys, Values: vals, Lineno: line}, Lineno: line},
	)
	if flag {
		var mask pyast.Node = &pyast.Constant{PyType: "int", Value: int64(0), Lineno: line}
		for _, mem := range members {
			mask = &pyast.BinOp{Left: mask, Op: "BitOr", Right: &pyast.Attribute{Value: &pyast.Name{ID: cls + "__" + mem.name, Lineno: line}, Attr: "_value_", Lineno: line}, Lineno: line}
		}
		assigned = append(assigned,
			&pyast.Assign{Targets: []pyast.Node{&pyast.Name{ID: cls + "__mask", Lineno: line}}, Value: mask, Lineno: line},
			&pyast.Assign{Targets: []pyast.Node{&pyast.Name{ID: cls + "__composites", Lineno: line}}, Value: &pyast.Dict{Lineno: line}, Lineno: line},
		)
	}
	out = append(out, assigned...)
	out = append(out, after...)
	return &block{stmts: out}, nil
}

// orderedNames lists the member names including aliases, in definition order.
func (e *enumInfo) orderedNames() []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range e.allDeclared {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

func enumHelperSource(cls string, info *enumInfo, flag bool) string {
	var b strings.Builder
	pyName := cls
	if i := strings.LastIndex(cls, Marker); i >= 0 {
		pyName = cls[i+len(Marker):]
	}
	fmt.Fprintf(&b, "class _H:\n")
	fmt.Fprintf(&b, "    def _lookup(self, value):\n        return %s__call(value)\n", cls)
	fmt.Fprintf(&b, "    def _members(self):\n        return %s__members\n", cls)
	if flag {
		fmt.Fprintf(&b, "    def _all_bits(self):\n        return %s__mask\n", cls)
	}
	if flag {
		fmt.Fprintf(&b, `def %[1]s__call(value):
    if isinstance(value, %[3]s):
        return value
    for _m in %[1]s__members:
        if _m._value_ == value:
            return _m
    if value < 0 or value & ~%[1]s__mask:
        _n = max(value.bit_length(), %[1]s__mask.bit_length())
        raise ValueError("<flag '" + %[2]q + "'> invalid value " + str(value) + "\n    given 0b0 " + format(value, "0" + str(_n) + "b") + "\n  allowed 0b0 " + format(%[1]s__mask, "0" + str(_n) + "b"))
    if value in %[1]s__composites:
        return %[1]s__composites[value]
    _c = %[3]s()
    _c._value_ = value
    _names = []
    for _m in %[1]s__members:
        if _m._value_ and _m._value_ & value == _m._value_ and (_m._value_ & (_m._value_ - 1)) == 0:
            _names.append(_m._name_)
    _c._name_ = "|".join(_names)
    %[1]s__composites[value] = _c
    return _c
`, cls, pyName, strings.TrimSuffix(cls, ""))
	} else {
		fmt.Fprintf(&b, `def %[1]s__call(value):
    for _m in %[1]s__members:
        if _m._value_ == value:
            return _m
    raise ValueError(repr(value) + " is not a valid " + %[2]q)
`, cls, pyName)
	}
	fmt.Fprintf(&b, "def %[1]s__getitem(name):\n    return %[1]s__byname[name]\n", cls)
	return b.String()
}

// parseSnippet parses Python source with the same parser as the files.
func parseSnippet(src string) (*pyast.Module, error) {
	dir, err := os.MkdirTemp("", "py2go-snippet-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "snippet.py")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		return nil, err
	}
	return parseFile(path)
}

// enumUse rewrites a use of an enum class that was already renamed.
func (r *rewriter) enumUse(n pyast.Node) pyast.Node {
	switch x := n.(type) {
	case *pyast.Attribute:
		if nm, ok := x.Value.(*pyast.Name); ok {
			info := r.l.enums[nm.ID]
			if info == nil && r.enumCls != "" && nm.ID == r.enumCls && !r.localShadow(nm.ID) {
				info = r.enumOf
			}
			if info != nil {
				if mem := info.byName[x.Attr]; mem != nil {
					return &pyast.Name{ID: info.cls + "__" + mem.name, Lineno: x.Lineno}
				}
				if x.Attr == "__members__" {
					return &pyast.Name{ID: info.cls + "__byname", Lineno: x.Lineno}
				}
			}
		}
	case *pyast.Call:
		if nm, ok := x.Func.(*pyast.Name); ok {
			if info := r.l.enums[nm.ID]; info != nil && len(x.Args) == 1 && len(x.Keywords) == 0 {
				return &pyast.Call{Func: &pyast.Name{ID: info.cls + "__call", Lineno: x.Lineno}, Args: x.Args, Lineno: x.Lineno}
			}
			if iterBuiltins[nm.ID] {
				for i, a := range x.Args {
					x.Args[i] = r.enumIter(a)
				}
			}
		}
	case *pyast.Subscript:
		if nm, ok := x.Value.(*pyast.Name); ok {
			if info := r.l.enums[nm.ID]; info != nil {
				return &pyast.Call{Func: &pyast.Name{ID: info.cls + "__getitem", Lineno: x.Lineno}, Args: []pyast.Node{x.Index}, Lineno: x.Lineno}
			}
		}
	case *pyast.For:
		x.Iter = r.enumIter(x.Iter)
	case *pyast.Compare:
		for i, c := range x.Comparators {
			x.Comparators[i] = r.enumIter(c)
		}
	}
	return n
}

var iterBuiltins = map[string]bool{"list": true, "tuple": true, "set": true, "frozenset": true, "sorted": true, "len": true, "reversed": true,
	"enumerate": true, "iter": true, "min": true, "max": true, "sum": true, "any": true, "all": true, "zip": true, "map": true, "filter": true, "dict": true}

// enumIter replaces an enum class that is iterated by the list of its members.
func (r *rewriter) enumIter(n pyast.Node) pyast.Node {
	if nm, ok := n.(*pyast.Name); ok {
		if info := r.l.enums[nm.ID]; info != nil {
			return &pyast.Name{ID: info.cls + "__members", Lineno: nm.Lineno}
		}
	}
	return n
}

func (r *rewriter) localShadow(name string) bool { return false }

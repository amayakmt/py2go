package codegen

import (
	"reflect"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// Python programs use names that mean something in Go: default, map, var,
// string. They are renamed before anything is generated, in every place the
// program spells them (variables, parameters, keywords of calls, defs and
// classes), so the rest of the generator sees ordinary names.

// goKeywords are the Go keywords that are not Python keywords.
var goKeywords = map[string]bool{
	"case": true, "chan": true, "const": true, "default": true, "defer": true, "fallthrough": true, "func": true,
	"go": true, "goto": true, "interface": true, "package": true, "select": true,
	"struct": true, "switch": true, "var": true,
}

// goTypeNames are the Go names the generated code uses and a program must not
// shadow.
var goTypeNames = map[string]bool{
	"float64": true, "int64": true, "string": true, "true": true, "false": true,
	"nil": true, "runtime": true, "init": true, "byte": true, "rune": true, "uint": true,
	"int32": true, "uint8": true, "float32": true, "error": true, "iota": true,
}

// goBuiltinFuncs are Go's own functions, which Python programs also use as
// names; they are renamed only when the program defines them itself, since
// len, max and min are Python builtins too.
var goBuiltinFuncs = map[string]bool{
	"append": true, "cap": true, "clear": true, "close": true, "complex": true, "copy": true,
	"delete": true, "imag": true, "len": true, "make": true, "max": true, "min": true, "new": true,
	"real": true, "recover": true, "any": true,
	// Go keywords that Python has as builtins.
	"map": true, "range": true, "type": true,
}

const renameSuffix = "_py"

// renameReserved renames the identifiers of the module that Go would not
// accept, and reports whether it renamed any.
func renameReserved(mod *pyast.Module) bool {
	// Methods and class attributes are reached by attribute access, which
	// keeps the name the program wrote, so those definitions keep it too.
	protected := map[uintptr]bool{}
	walk(mod, func(n pyast.Node) {
		cd, ok := n.(*pyast.ClassDef)
		if !ok {
			return
		}
		for _, st := range cd.Body {
			switch s := st.(type) {
			case *pyast.FunctionDef:
				protected[reflect.ValueOf(s).Pointer()] = true
			case *pyast.Assign:
				for _, tg := range s.Targets {
					if nm, ok := tg.(*pyast.Name); ok {
						protected[reflect.ValueOf(nm).Pointer()] = true
					}
				}
			case *pyast.AnnAssign:
				if nm, ok := s.Target.(*pyast.Name); ok {
					protected[reflect.ValueOf(nm).Pointer()] = true
				}
			}
		}
	})
	// A Go builtin such as min or len is renamed where the program binds it:
	// at module level everywhere, in a function inside that function. A use
	// of the builtin elsewhere keeps its name.
	moduleBound := scopeBound(nil, mod.Body, protected)
	var scopes []map[string]bool
	active := func(name string) bool {
		if moduleBound[name] {
			return true
		}
		for _, sc := range scopes {
			if sc[name] {
				return true
			}
		}
		return false
	}
	// A keyword of a call is renamed only when some def takes a parameter of
	// that name; otherwise it belongs to a builtin or to a mapping.
	params := map[string]bool{}
	walk(mod, func(n pyast.Node) {
		if a, ok := n.(*pyast.Arg); ok {
			params[a.Name] = true
		}
	})
	rename := func(name string) string {
		if goKeywords[name] || goTypeNames[name] || (goBuiltinFuncs[name] && active(name)) {
			return name + renameSuffix
		}
		return name
	}
	renameKeyword := func(name string) string {
		if params[name] {
			return rename(name)
		}
		return name
	}
	changed := false
	nodeType := reflect.TypeOf((*pyast.Node)(nil)).Elem()
	var visit func(v reflect.Value)
	visit = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Interface:
			if !v.IsNil() {
				visit(v.Elem())
			}
		case reflect.Pointer:
			if v.IsNil() {
				return
			}
			if (v.Type().Implements(nodeType) || v.Type().Elem().Kind() == reflect.Struct) && !protected[v.Pointer()] {
				renameFields(v.Elem(), rename, renameKeyword, &changed)
			}
			switch x := v.Interface().(type) {
			case *pyast.FunctionDef:
				scopes = append(scopes, scopeBound(x.Args, x.Body, protected))
				defer func() { scopes = scopes[:len(scopes)-1] }()
			case *pyast.Lambda:
				scopes = append(scopes, scopeBound(x.Args, nil, protected))
				defer func() { scopes = scopes[:len(scopes)-1] }()
			}
			visit(v.Elem())
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Type().Field(i).IsExported() {
					visit(v.Field(i))
				}
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				visit(v.Index(i))
			}
		}
	}
	visit(reflect.ValueOf(mod))
	return changed
}

// renameFields renames the identifier fields of one node.
func renameFields(v reflect.Value, rename, renameKeyword func(string) string, changed *bool) {
	if v.Kind() != reflect.Struct {
		return
	}
	setWith := func(f reflect.Value, rename func(string) string) {
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			if n := rename(f.String()); n != f.String() {
				f.SetString(n)
				*changed = true
			}
		}
	}
	set := func(f reflect.Value) { setWith(f, rename) }
	switch v.Type().Name() {
	case "Name":
		set(v.FieldByName("ID"))
	case "FunctionDef", "ClassDef", "ExceptHandler", "Arg", "MatchStar", "MatchAs":
		set(v.FieldByName("Name"))
	case "Keyword":
		setWith(v.FieldByName("Arg"), renameKeyword)
	case "MatchMapping":
		set(v.FieldByName("Rest"))
	case "Global", "Nonlocal":
		names := v.FieldByName("Names")
		for i := 0; i < names.Len(); i++ {
			set(names.Index(i))
		}
	}
}

// boundNames finds the names the program defines or assigns.
func boundNames(mod *pyast.Module, protected map[uintptr]bool) map[string]bool {
	bound := map[string]bool{}
	var targets func(n pyast.Node)
	targets = func(n pyast.Node) {
		switch x := n.(type) {
		case *pyast.Name:
			if !protected[reflect.ValueOf(x).Pointer()] {
				bound[x.ID] = true
			}
		case *pyast.Tuple:
			for _, e := range x.Elts {
				targets(e)
			}
		case *pyast.List:
			for _, e := range x.Elts {
				targets(e)
			}
		case *pyast.Starred:
			targets(x.Value)
		}
	}
	walk(mod, func(n pyast.Node) {
		switch n := n.(type) {
		case *pyast.FunctionDef:
			if !protected[reflect.ValueOf(n).Pointer()] {
				bound[n.Name] = true
			}
		case *pyast.ClassDef:
			bound[n.Name] = true
		case *pyast.Arg:
			bound[n.Name] = true
		case *pyast.Assign:
			for _, t := range n.Targets {
				targets(t)
			}
		case *pyast.AugAssign:
			targets(n.Target)
		case *pyast.AnnAssign:
			targets(n.Target)
		case *pyast.For:
			targets(n.Target)
		}
	})
	return bound
}

// pyName is the name the program wrote for a name that was renamed.
func pyName(name string) string {
	// A name of another module of the project carries its module as a prefix.
	if i := strings.LastIndex(name, "_pm_"); i >= 0 {
		name = name[i+len("_pm_"):]
	}
	if base, ok := strings.CutSuffix(name, renameSuffix); ok && (goKeywords[base] || goTypeNames[base] || goBuiltinFuncs[base]) {
		return base
	}
	return name
}

// scopeBound is the names a scope binds: its parameters, the targets of its
// assignments and loops, and the names of the defs and classes in it, not
// those of the functions and classes below.
func scopeBound(args *pyast.Arguments, body []pyast.Node, protected map[uintptr]bool) map[string]bool {
	bound := map[string]bool{}
	if args != nil {
		for _, a := range args.Positional() {
			bound[a.Name] = true
		}
		for _, a := range args.Kwonly {
			bound[a.Name] = true
		}
		if args.Vararg != nil {
			bound[args.Vararg.Name] = true
		}
		if args.Kwarg != nil {
			bound[args.Kwarg.Name] = true
		}
	}
	var targets func(n pyast.Node)
	targets = func(n pyast.Node) {
		switch x := n.(type) {
		case *pyast.Name:
			if !protected[reflect.ValueOf(x).Pointer()] {
				bound[x.ID] = true
			}
		case *pyast.Tuple:
			for _, e := range x.Elts {
				targets(e)
			}
		case *pyast.List:
			for _, e := range x.Elts {
				targets(e)
			}
		case *pyast.Starred:
			targets(x.Value)
		}
	}
	for _, st := range body {
		walkPrune(st, func(n pyast.Node) bool {
			switch x := n.(type) {
			case *pyast.FunctionDef:
				if !protected[reflect.ValueOf(x).Pointer()] {
					bound[x.Name] = true
				}
				return false
			case *pyast.ClassDef:
				bound[x.Name] = true
				return false
			case *pyast.Lambda:
				return false
			case *pyast.Assign:
				for _, t := range x.Targets {
					targets(t)
				}
			case *pyast.AugAssign:
				targets(x.Target)
			case *pyast.AnnAssign:
				targets(x.Target)
			case *pyast.For:
				targets(x.Target)
			case *pyast.With:
				for _, it := range x.Items {
					if it.Vars != nil {
						targets(it.Vars)
					}
				}
			case *pyast.ExceptHandler:
				if x.Name != "" {
					bound[x.Name] = true
				}
			case *pyast.NamedExpr:
				targets(x.Target)
			}
			return true
		})
	}
	return bound
}

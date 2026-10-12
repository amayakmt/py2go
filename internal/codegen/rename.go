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
	bound := boundNames(mod, protected)
	// A keyword of a call is renamed only when some def takes a parameter of
	// that name; otherwise it belongs to a builtin or to a mapping.
	params := map[string]bool{}
	walk(mod, func(n pyast.Node) {
		if a, ok := n.(*pyast.Arg); ok {
			params[a.Name] = true
		}
	})
	rename := func(name string) string {
		if goKeywords[name] || goTypeNames[name] || (goBuiltinFuncs[name] && bound[name]) {
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
	if base, ok := strings.CutSuffix(name, renameSuffix); ok && (goKeywords[base] || goTypeNames[base] || goBuiltinFuncs[base]) {
		return base
	}
	return name
}

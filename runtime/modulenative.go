package runtime

import (
	"sort"
	"strings"
)

// A module used as a value is a ModuleObj: its attributes are read through
// the getters the program registers for the top-level names of the module.

var moduleAttrs = map[string]map[string]func() any{}
var moduleObjs = map[string]*ModuleObj{}

// ModuleObj is a module taken as a value.
type ModuleObj struct {
	Name  string
	extra map[string]any
}

func init() {
	RegisterNative("module.object", func(args []any, kw []KW) any {
		name := Unbox(args[0]).(string)
		if m, ok := moduleObjs[name]; ok {
			return m
		}
		m := &ModuleObj{Name: name}
		moduleObjs[name] = m
		return m
	})
	// module.declare(name): the program has the module (it registers the
	// modules it links when it uses importlib).
	RegisterNative("module.declare", func(args []any, kw []KW) any {
		declaredModules[Unbox(args[0]).(string)] = true
		return nil
	})
	// module.import(name): importlib.import_module of a module the program
	// links.
	RegisterNative("module.import", func(args []any, kw []KW) any {
		name := Str(args[0])
		if !declaredModules[name] {
			e := newError("ModuleNotFoundError", "No module named '"+name+"'")
			e.Attrs = map[string]any{"name": name, "path": nil}
			panic(e)
		}
		return NativeCall([]any{"module.object", name}, nil)
	})
	RegisterNative("module.linked", func(args []any, kw []KW) any {
		return declaredModules[Str(args[0])]
	})
	RegisterNative("module.register", func(args []any, kw []KW) any {
		mod, name := Unbox(args[0]).(string), Unbox(args[1]).(string)
		get := args[2]
		m := moduleAttrs[mod]
		if m == nil {
			m = map[string]func() any{}
			moduleAttrs[mod] = m
		}
		m[name] = func() any { return DynCall(get) }
		return nil
	})
}

// declaredModules are the modules importlib.import_module finds.
var declaredModules = map[string]bool{}

func (m *ModuleObj) PyRepr() string      { return "<module '" + m.Name + "'>" }
func (m *ModuleObj) PyClassName() string { return "module" }

func (m *ModuleObj) PyGetAttr(name string) (any, bool) {
	if v, ok := m.extra[name]; ok {
		return v, true
	}
	switch name {
	case "__name__":
		return m.Name, true
	}
	if get, ok := moduleAttrs[m.Name][name]; ok {
		return get(), true
	}
	return nil, false
}

func (m *ModuleObj) PySetAttr(name string, v any) bool {
	if m.extra == nil {
		m.extra = map[string]any{}
	}
	m.extra[name] = v
	return true
}

func (m *ModuleObj) PyAttrNames() []string {
	seen := map[string]bool{"__name__": true}
	out := []string{"__name__"}
	for n := range moduleAttrs[m.Name] {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	for n := range m.extra {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// Virtual subclasses of abstract base classes: ABC.register(sub).
var virtualSubclasses = map[string][]*ClassRef{}

func init() {
	RegisterNative("abc.register", func(args []any, kw []KW) any {
		abc, ok1 := Unbox(args[0]).(*ClassRef)
		sub, ok2 := Unbox(args[1]).(*ClassRef)
		if !ok1 || !ok2 {
			panic(newError("TypeError", "Can only register classes"))
		}
		virtualSubclasses[abc.Name] = append(virtualSubclasses[abc.Name], sub)
		return nil
	})
}

// abcsUnder lists the ABC name and the classes of the program that derive from it.
func abcsUnder(name string) []string {
	out := []string{name}
	for k, chain := range classBases {
		if k == name {
			continue
		}
		for _, b := range chain[1:] {
			if b == name {
				out = append(out, k)
				break
			}
		}
	}
	return out
}

// IsVirtualInstance is the part of isinstance(x, C) that ABC.register adds.
func IsVirtualInstance(x any, name string) bool {
	if structuralMatch(x, name) {
		return true
	}
	for _, abc := range abcsUnder(name) {
		for _, sub := range virtualSubclasses[abc] {
			if isInstanceName(x, sub.Name, sub.Builtin) {
				return true
			}
		}
	}
	return false
}

// IsVirtualSubclass is the part of issubclass(C, A) that ABC.register adds.
func IsVirtualSubclass(c *ClassRef, name string) bool {
	for _, abc := range abcsUnder(name) {
		for _, sub := range virtualSubclasses[abc] {
			if sub.Name == c.Name || DynIsSubclass(c, sub) {
				return true
			}
		}
	}
	return false
}

func init() {
	RegisterNative("module.globals", func(args []any, kw []KW) any {
		name := Unbox(args[0]).(string)
		out := NewDict[string, any]()
		out.Set("__name__", name)
		var names []string
		for n := range moduleAttrs[name] {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			func() {
				defer func() {
					// A name not bound yet (later in the module) is left out.
					if r := recover(); r != nil {
						if _, isPy := r.(*PyError); !isPy {
							panic(r)
						}
					}
				}()
				out.Set(n, moduleAttrs[name][n]())
			}()
		}
		return out
	})
}

// structuralABCs are the ABCs whose __subclasshook__ accepts any object with
// the given methods (collections.abc.Iterable and the like).
var structuralABCs = map[string][]string{}

func init() {
	RegisterNative("abc.structural", func(args []any, kw []KW) any {
		c := Unbox(args[0]).(*ClassRef)
		var names []string
		for it := DynIter(args[1]).Iter(); it.Next(); {
			names = append(names, Unbox(it.Value()).(string))
		}
		structuralABCs[c.Name] = names
		return nil
	})
}

// hasProtocolMethod tells whether x answers the special method: user objects
// by their class, built-in values by what the runtime does with them.
func hasProtocolMethod(x any, name string) bool {
	x = Unbox(x)
	if _, user := x.(Named); user {
		if _, isErr := x.(*PyError); !isErr {
			if DynHasAttr(x, name) {
				return true
			}
			if m, ok := x.(interface{ PyMethodNames() []string }); ok {
				for _, n := range m.PyMethodNames() {
					if n == name {
						return true
					}
				}
			}
		}
	}
	switch name {
	case "__iter__":
		return Catch(func() { DynIter(x) }) == nil
	case "__next__":
		tn := TypeName(x)
		return tn == "generator" || strings.HasSuffix(tn, "iterator") || tn == "enumerate" || tn == "zip" || tn == "map" || tn == "filter" || tn == "reversed"
	case "__len__":
		return Catch(func() { DynLen(x) }) == nil
	case "__contains__":
		if _, isStr := x.(string); isStr {
			return true
		}
		return Catch(func() { DynLen(x) }) == nil
	case "__call__":
		return DynCallable(x)
	case "__hash__":
		return Catch(func() { Hash(x) }) == nil
	case "__reversed__":
		return Catch(func() { DynLen(x) }) == nil
	}
	return false
}

func structuralMatch(x any, abc string) bool {
	names, ok := structuralABCs[abc]
	if !ok {
		return false
	}
	for _, n := range names {
		if !hasProtocolMethod(x, n) {
			return false
		}
	}
	return true
}

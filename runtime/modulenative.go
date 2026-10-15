package runtime

import "sort"

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

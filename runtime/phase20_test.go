package runtime

import "testing"

func TestDictGetOptTreatsNoneAsMissing(t *testing.T) {
	d := NewDict[string, any]()
	d.Set("none", NoneV)
	d.Set("nil", nil)
	d.Set("zero", int64(0))
	if d.GetOpt("none").IsSome() || d.GetOpt("nil").IsSome() || d.GetOpt("absent").IsSome() {
		t.Errorf("None and missing keys must give None")
	}
	if !d.GetOpt("zero").IsSome() {
		t.Errorf("0 is not None")
	}
}

func TestPercentFormatWithAMappingAndNoFields(t *testing.T) {
	m := NewDict[string, any]()
	m.Set("k", int64(1))
	if got := PercentFormat("plain", m); got != "plain" {
		t.Errorf("got %q", got)
	}
	if got := catch(func() { PercentFormat("plain", NewVarTuple[any](int64(1), int64(2))) }); got != "TypeError: not all arguments converted during string formatting" {
		t.Errorf("tuple: %s", got)
	}
}

func TestDictWithKeywords(t *testing.T) {
	base := NewDict[string, any]()
	base.Set("a", int64(1))
	got := StrDictWith(base, "b", "two", "a", int64(3))
	if Repr(got) != "{'a': 3, 'b': 'two'}" {
		t.Errorf("StrDictWith = %s", Repr(got))
	}
	if Repr(DynDictWith(nil, "x", int64(1))) != "{'x': 1}" {
		t.Errorf("DynDictWith without a base")
	}
}

func TestAttrNamesKeepInsertionOrder(t *testing.T) {
	extra := map[string]any{"zeta": 1, "alpha": 2, "mid": 3}
	got := AttrNames([]string{"f"}, extra, []string{"zeta", "mid", "alpha"})
	if len(got) != 4 || got[1] != "zeta" || got[2] != "mid" || got[3] != "alpha" {
		t.Errorf("AttrNames = %v", got)
	}
	if got := DropName([]string{"a", "b", "c"}, "b"); len(got) != 2 || got[1] != "c" {
		t.Errorf("DropName = %v", got)
	}
}

func TestModuleObject(t *testing.T) {
	callNative("module.register", "p20mod", "VALUE", func() any { return int64(7) })
	m := callNative("module.object", "p20mod")
	if v := DynGetAttr(m, "VALUE"); v != int64(7) {
		t.Errorf("VALUE = %v", v)
	}
	if v := DynGetAttr(m, "__name__"); v != "p20mod" {
		t.Errorf("__name__ = %v", v)
	}
	if DynHasAttr(m, "absent") || !DynHasAttr(m, "VALUE") {
		t.Errorf("hasattr on a module")
	}
	if callNative("module.object", "p20mod") != m {
		t.Errorf("a module is one object")
	}
}

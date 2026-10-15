package runtime

import "testing"

func TestBuiltinDunderMethodsAreBound(t *testing.T) {
	s := NewSet[int64]()
	s.Add(3)
	f := DynGetAttr(s, "__contains__")
	if r := DynCall(f, int64(3)); r != true {
		t.Errorf("set.__contains__(3) = %v", r)
	}
	if r := DynCall(DynGetAttr(NewList[int64](1, 2), "__len__")); r != int64(2) {
		t.Errorf("list.__len__() = %v", r)
	}
}

func TestIsInstanceOfType(t *testing.T) {
	if !IsInstanceName(ClassOf("A"), "type", true) {
		t.Errorf("a class is an instance of type")
	}
	if IsInstanceName(int64(3), "type", true) {
		t.Errorf("3 is not an instance of type")
	}
}

func TestClassMembersAndCallables(t *testing.T) {
	RegisterClassMembers("P18", "helper", "test_one")
	RegisterClassCallable("P18", "make", func() any { return func() any { return int64(7) } })
	c := ClassOf("P18")
	names := DynDir(c)
	if names.Len() != 2 || names.Get(0) != "helper" || names.Get(1) != "test_one" {
		t.Errorf("dir = %v", names)
	}
	if r := DynCall(DynGetAttr(c, "make")); r != int64(7) {
		t.Errorf("classmethod = %v", r)
	}
	if !DynHasAttr(c, "helper") || DynHasAttr(c, "absent") {
		t.Errorf("hasattr on the class members")
	}
}

func TestSigData(t *testing.T) {
	f := WithInfo(func() {}, &FuncInfo{Name: "f", Params: []string{"a", "b"}, KwOnly: map[string]bool{"b": true}})
	data := callNative("inspect.sigdata", f).(Seq)
	if data.At(0) != "f" || data.At(1).(*List[any]).Len() != 2 {
		t.Errorf("sigdata = %v", Repr(data))
	}
	if callNative("inspect.sigdata", int64(3)) != nil {
		t.Errorf("sigdata of an int should be None")
	}
}

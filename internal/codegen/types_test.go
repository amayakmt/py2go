package codegen

import (
	"testing"

	"github.com/amayakmt/py2go/internal/pyast"
)

func name(id string) *pyast.Name { return &pyast.Name{ID: id, Lineno: 6} }

func sub(head string, index pyast.Node) *pyast.Subscript {
	return &pyast.Subscript{Value: name(head), Index: index, Lineno: 6}
}

func TestAnnotationType(t *testing.T) {
	tests := []struct {
		name      string
		node      pyast.Node
		allowNone bool
		want      pyType
		wantErr   string
	}{
		{"int", &pyast.Name{ID: "int", Lineno: 1}, false, tInt, ""},
		{"float", &pyast.Name{ID: "float", Lineno: 1}, false, tFloat, ""},
		{"bool", &pyast.Name{ID: "bool", Lineno: 1}, false, tBool, ""},
		{"str", &pyast.Name{ID: "str", Lineno: 1}, false, tStr, ""},
		{"None as return", &pyast.Constant{PyType: "none", Lineno: 1}, true, tNone, ""},
		{"None as parameter", &pyast.Constant{PyType: "none", Lineno: 3}, false, "",
			"unsupported: type annotation None at line 3"},
		{"list", &pyast.Name{ID: "list", Lineno: 4}, false, "",
			"unsupported: type annotation 'list' at line 4"},
		{"unknown kind", &pyast.Unsupported{Type: "BinOp", Lineno: 5}, false, "",
			"unsupported: type annotation of kind BinOp at line 5"},
		{"list[int]", sub("list", name("int")), false, listOf(tInt), ""},
		{"list[list[str]]", sub("list", sub("list", name("str"))), false, listOf(listOf(tStr)), ""},
		{"tuple[int, str]", sub("tuple", &pyast.Tuple{Elts: []pyast.Node{name("int"), name("str")}}), false,
			tupleOf([]pyType{tInt, tStr}), ""},
		{"tuple[int]", sub("tuple", name("int")), false, tupleOf([]pyType{tInt}), ""},
		{"tuple[int, ...]", sub("tuple", &pyast.Tuple{Elts: []pyast.Node{name("int"), &pyast.Unsupported{Type: "ellipsis constant", Lineno: 6}}}), false,
			varTupleOf(tInt), ""},
		{"tuple[()]", sub("tuple", &pyast.Tuple{Lineno: 6}), false, "",
			"unsupported: empty tuple annotation at line 6"},
		{"tuple of seven", sub("tuple", &pyast.Tuple{Elts: []pyast.Node{name("int"), name("int"), name("int"), name("int"), name("int"), name("int"), name("int")}}), false, "",
			"unsupported: tuple with more than 6 elements at line 6"},
		{"frozenset[int]", sub("frozenset", name("int")), false, "",
			"unsupported: type annotation 'frozenset[...]' at line 6"},
		{"dict[str, list[int]]", sub("dict", &pyast.Tuple{Elts: []pyast.Node{name("str"), sub("list", name("int"))}}), false,
			dictOf(tStr, listOf(tInt)), ""},
		{"dict[list[int], int]", sub("dict", &pyast.Tuple{Elts: []pyast.Node{sub("list", name("int")), name("int")}}), false, "",
			"unsupported: dict key type list[int] (not hashable) at line 6"},
		{"dict[int]", sub("dict", name("int")), false, "",
			"unsupported: dict[...] with 1 type arguments at line 6"},
		{"set[int]", sub("set", name("int")), false, setOf(tInt), ""},
		{"set[list[int]]", sub("set", sub("list", name("int"))), false, "",
			"unsupported: set element type list[int] (not hashable) at line 6"},
		{"set[tuple[int, str]]", sub("set", sub("tuple", &pyast.Tuple{Elts: []pyast.Node{name("int"), name("str")}})), false,
			setOf(tupleOf([]pyType{tInt, tStr})), ""},
		{"range", name("range"), false, tRange, ""},
		{"list[None]", sub("list", &pyast.Constant{PyType: "none", Lineno: 6}), false, "",
			"unsupported: type annotation None at line 6"},
		{"list[float]", sub("list", name("float")), true, listOf(tFloat), ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := annotationType(tc.node, tc.allowNone)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("want error %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGoType(t *testing.T) {
	if d := dictOf(tStr, listOf(tInt)); d.keyType() != tStr || d.valType() != listOf(tInt) {
		t.Errorf("dict parameters: %s", d)
	}
	if k, ok := dictOf(tInt, tStr).iterElem(); !ok || k != tInt {
		t.Error("a dict iterates over its keys")
	}
	if it, ok := pyType("dict_items[int, str]").iterElem(); !ok || it != tupleOf([]pyType{tInt, tStr}) {
		t.Errorf("dict items iterate as tuples, got %s", it)
	}
	if _, ok := tupleOf([]pyType{tInt, tInt}).iterElem(); ok {
		t.Error("a tuple cannot be iterated: its items have different types")
	}
	for _, c := range []struct {
		t  pyType
		ok bool
	}{{tInt, true}, {tStr, true}, {tFloat, true}, {listOf(tInt), false}, {dictOf(tInt, tInt), false},
		{tupleOf([]pyType{tInt, tStr}), true}, {tupleOf([]pyType{tInt, listOf(tInt)}), false}} {
		if c.t.hashable() != c.ok {
			t.Errorf("hashable(%s) = %v", c.t, !c.ok)
		}
	}
	want := map[pyType]string{
		dictOf(tStr, listOf(tInt)):      "*runtime.Dict[string, *runtime.List[int64]]",
		setOf(tInt):                     "*runtime.Set[int64]",
		tRange:                          "*runtime.Range",
		iterOf(tStr):                    "*runtime.Iterator[string]",
		pyType("dict_keys[int, str]"):   "*runtime.KeysView[int64, string]",
		pyType("dict_values[int, str]"): "*runtime.ValuesView[int64, string]",
		pyType("dict_items[int, str]"):  "*runtime.ItemsView[int64, string]", tInt: "int64", tFloat: "float64", tBool: "bool", tStr: "string"}
	for pt, goType := range want {
		if got := pt.goType(); got != goType {
			t.Errorf("%v: want %s, got %s", pt, goType, got)
		}
	}
}

func TestContainerTypes(t *testing.T) {
	nested := tupleOf([]pyType{tInt, listOf(tStr), tupleOf([]pyType{tFloat, tBool})})
	if nested != "tuple[int, list[str], tuple[float, bool]]" {
		t.Fatalf("spelling: %s", nested)
	}
	elems := nested.elems()
	if len(elems) != 3 || elems[1] != listOf(tStr) || elems[2] != tupleOf([]pyType{tFloat, tBool}) {
		t.Errorf("elems must split only at the top level, got %v", elems)
	}
	if !nested.isTuple() || nested.isList() || !listOf(tInt).isList() {
		t.Error("isList/isTuple")
	}
	if listOf(listOf(tInt)).elem() != listOf(tInt) {
		t.Error("elem of a nested list")
	}
	want := map[pyType]string{
		listOf(tInt):                             "*runtime.List[int64]",
		listOf(listOf(tStr)):                     "*runtime.List[*runtime.List[string]]",
		tupleOf([]pyType{tInt, tStr}):            "runtime.Tuple2[int64, string]",
		tupleOf([]pyType{tFloat, listOf(tBool)}): "runtime.Tuple2[float64, *runtime.List[bool]]",
	}
	for typ, goType := range want {
		if got := typ.goType(); got != goType {
			t.Errorf("%s: want %s, got %s", typ, goType, got)
		}
	}
}

func TestScope(t *testing.T) {
	outer := newScope(nil)
	outer.declare("x", tInt)
	inner := newScope(outer)
	inner.declare("y", tStr)

	if typ, ok := inner.lookup("x"); !ok || typ != tInt {
		t.Errorf("inner must see outer x, got %v %v", typ, ok)
	}
	if _, ok := outer.lookup("y"); ok {
		t.Error("outer must not see inner y")
	}
	if inner.declare("y", tStr) {
		t.Error("declaring y twice in one block must report false")
	}

	// A function body starts a new chain: no access to the caller's names.
	fn := newScope(nil)
	if _, ok := fn.lookup("x"); ok {
		t.Error("function scope must not see main's variables")
	}
}

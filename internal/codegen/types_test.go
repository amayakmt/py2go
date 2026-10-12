package codegen

import (
	"testing"

	"github.com/amayakmt/py2go/internal/pyast"
)

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
		{"None as parameter", &pyast.Constant{PyType: "none", Lineno: 3}, false, 0,
			"unsupported: type annotation None at line 3"},
		{"list", &pyast.Name{ID: "list", Lineno: 4}, false, 0,
			"unsupported: type annotation 'list' at line 4"},
		{"list[int]", &pyast.Unsupported{Type: "Subscript", Lineno: 5}, false, 0,
			"unsupported: type annotation of kind Subscript at line 5"},
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
	want := map[pyType]string{tInt: "int64", tFloat: "float64", tBool: "bool", tStr: "string"}
	for pt, goType := range want {
		if got := pt.goType(); got != goType {
			t.Errorf("%v: want %s, got %s", pt, goType, got)
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

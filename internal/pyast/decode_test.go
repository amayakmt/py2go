package pyast

import (
	"os"
	"testing"
)

// as asserts that n has concrete type T and returns it.
// Fails the test with a clear message instead of panicking on a bad assertion.
func as[T any](t *testing.T, n Node) T {
	t.Helper()
	v, ok := n.(T)
	if !ok {
		var zero T
		t.Fatalf("expected %T, got %T", zero, n)
	}
	return v
}

func TestDecodeHello(t *testing.T) {
	data, err := os.ReadFile("../../testdata/phase1/hello/ast.json")
	if err != nil {
		t.Fatal(err)
	}

	root, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}

	mod := as[*Module](t, root)
	if len(mod.Body) != 2 {
		t.Fatalf("module body: want 2 statements, got %d", len(mod.Body))
	}

	// def greet() -> None: print("hello")
	fn := as[*FunctionDef](t, mod.Body[0])
	if fn.Name != "greet" {
		t.Errorf("function name: want greet, got %q", fn.Name)
	}
	if fn.Line() != 1 {
		t.Errorf("function line: want 1, got %d", fn.Line())
	}
	// "-> None" is a Constant node holding nil, not an absent return type.
	ret := as[*Constant](t, fn.Returns)
	if ret.Value != nil {
		t.Errorf("return annotation: want nil constant, got %v", ret.Value)
	}

	if len(fn.Body) != 1 {
		t.Fatalf("function body: want 1 statement, got %d", len(fn.Body))
	}
	stmt := as[*Expr](t, fn.Body[0])
	if stmt.Line() != 2 {
		t.Errorf("print statement line: want 2, got %d", stmt.Line())
	}
	call := as[*Call](t, stmt.Value)
	callee := as[*Name](t, call.Func)
	if callee.ID != "print" {
		t.Errorf("callee: want print, got %q", callee.ID)
	}
	if len(call.Args) != 1 {
		t.Fatalf("print args: want 1, got %d", len(call.Args))
	}
	arg := as[*Constant](t, call.Args[0])
	if arg.Value != "hello" {
		t.Errorf("print arg: want \"hello\", got %v", arg.Value)
	}

	// greet()
	top := as[*Expr](t, mod.Body[1])
	if top.Line() != 5 {
		t.Errorf("top-level call line: want 5, got %d", top.Line())
	}
	topCall := as[*Call](t, top.Value)
	if as[*Name](t, topCall.Func).ID != "greet" {
		t.Errorf("top-level call should target greet")
	}
	if len(topCall.Args) != 0 {
		t.Errorf("greet() args: want 0, got %d", len(topCall.Args))
	}
}

func TestDecodeUnsupported(t *testing.T) {
	// Hand-written JSON: a module containing a node type we don't implement.
	data := []byte(`{"_type": "Module", "body": [{"_type": "Match", "lineno": 12}]}`)

	root, err := Decode(data)
	if err != nil {
		t.Fatalf("unknown node types must not be a decode error: %v", err)
	}

	mod := as[*Module](t, root)
	u := as[*Unsupported](t, mod.Body[0])
	if u.Type != "Match" {
		t.Errorf("type: want Match, got %q", u.Type)
	}
	if u.Line() != 12 {
		t.Errorf("line: want 12, got %d", u.Line())
	}
}

func TestDecodeMalformed(t *testing.T) {
	if _, err := Decode([]byte(`{not json`)); err == nil {
		t.Fatal("expected an error on malformed JSON")
	}
}

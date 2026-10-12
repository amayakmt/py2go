package pyast

import (
	"os/exec"
	"testing"

	"github.com/amayakmt/py2go/internal/pyrun"
)

// decodeFile runs the real dumper on a Python file, so these tests
// check the decoder against the JSON that CPython 3.13 actually produces.
func decodeFile(t *testing.T, path string) *Module {
	t.Helper()
	if _, err := exec.LookPath("uv"); err != nil {
		t.Skip("uv not installed")
	}
	js, err := pyrun.DumpAST(path)
	if err != nil {
		t.Fatal(err)
	}
	root, err := Decode(js)
	if err != nil {
		t.Fatal(err)
	}
	return as[*Module](t, root)
}

func TestDecodeSignature(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/typed_core.py")

	fn := as[*FunctionDef](t, mod.Body[0])
	a := fn.Args
	if len(a.PosOnly) != 2 || a.PosOnly[0].Name != "a" || a.PosOnly[1].Name != "b" {
		t.Errorf("posonly: got %+v", a.PosOnly)
	}
	if len(a.Args) != 1 || a.Args[0].Name != "c" {
		t.Errorf("args: got %+v", a.Args)
	}
	if pos := a.Positional(); len(pos) != 3 || pos[2].Name != "c" {
		t.Errorf("Positional(): want a, b, c in order, got %+v", pos)
	}
	if a.Vararg == nil || a.Vararg.Name != "args" || a.Vararg.Line() != 1 {
		t.Errorf("vararg: got %+v", a.Vararg)
	}
	if len(a.Kwonly) != 1 || a.Kwonly[0].Name != "d" {
		t.Errorf("kwonly: got %+v", a.Kwonly)
	}
	if a.Kwarg == nil || a.Kwarg.Name != "kw" {
		t.Errorf("kwarg: got %+v", a.Kwarg)
	}
	ann := as[*Name](t, a.Args[0].Annotation)
	if ann.ID != "str" {
		t.Errorf("annotation: want str, got %q", ann.ID)
	}
	if !fn.HasParams || fn.HasDecorators {
		t.Errorf("flags: HasParams=%v HasDecorators=%v", fn.HasParams, fn.HasDecorators)
	}

	fn2 := as[*FunctionDef](t, mod.Body[1])
	if !fn2.HasDecorators {
		t.Error("decorated function: HasDecorators must be set")
	}
	if len(fn2.Args.Defaults) != 1 || len(fn2.Args.KwDefaults) != 1 {
		t.Errorf("defaults: got %d positional, %d keyword-only",
			len(fn2.Args.Defaults), len(fn2.Args.KwDefaults))
	}
}

func TestDecodeStatementsAndExpressions(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/typed_core.py")
	body := as[*FunctionDef](t, mod.Body[0]).Body

	// x: int = a + 1
	ann := as[*AnnAssign](t, body[0])
	if !ann.Simple || as[*Name](t, ann.Target).ID != "x" || ann.Line() != 2 {
		t.Errorf("AnnAssign: got %+v", ann)
	}
	add := as[*BinOp](t, ann.Value)
	if add.Op != "Add" {
		t.Errorf("BinOp op: want Add, got %q", add.Op)
	}
	if one := as[*Constant](t, add.Right); one.PyType != "int" || one.Value != int64(1) {
		t.Errorf("constant 1: got %+v", one)
	}

	// x += 2
	aug := as[*AugAssign](t, body[1])
	if aug.Op != "Add" || as[*Name](t, aug.Target).ID != "x" {
		t.Errorf("AugAssign: got %+v", aug)
	}

	// y = -x
	asg := as[*Assign](t, body[2])
	neg := as[*UnaryOp](t, asg.Value)
	if len(asg.Targets) != 1 || neg.Op != "USub" {
		t.Errorf("Assign: got %+v, unary op %q", asg, neg.Op)
	}

	// if x < 3 < 4 and not y: ... elif ...: ... else: ...
	ifs := as[*If](t, body[3])
	and := as[*BoolOp](t, ifs.Test)
	if and.Op != "And" || len(and.Values) != 2 {
		t.Fatalf("BoolOp: got %+v", and)
	}
	cmp := as[*Compare](t, and.Values[0])
	if len(cmp.Ops) != 2 || cmp.Ops[0] != "Lt" || cmp.Ops[1] != "Lt" || len(cmp.Comparators) != 2 {
		t.Errorf("Compare: got ops %v", cmp.Ops)
	}
	if not := as[*UnaryOp](t, and.Values[1]); not.Op != "Not" {
		t.Errorf("not: got %q", not.Op)
	}
	ret := as[*Return](t, ifs.Body[0])
	if fd := as[*BinOp](t, ret.Value); fd.Op != "FloorDiv" {
		t.Errorf("want FloorDiv, got %q", fd.Op)
	}
	elif := as[*If](t, ifs.Orelse[0]) // elif is an If nested in Orelse
	if or := as[*BoolOp](t, elif.Test); or.Op != "Or" {
		t.Errorf("elif test: want Or, got %q", or.Op)
	}
	as[*Pass](t, elif.Body[0])
	if bare := as[*Return](t, elif.Orelse[0]); bare.Value != nil {
		t.Errorf("bare return must decode to nil Value, got %+v", bare.Value)
	}

	wh := as[*While](t, body[4])
	as[*Break](t, wh.Body[0])

	fr := as[*For](t, body[5])
	if as[*Name](t, fr.Target).ID != "i" {
		t.Errorf("for target: got %+v", fr.Target)
	}
	rng := as[*Call](t, fr.Iter)
	if len(rng.Args) != 3 {
		t.Fatalf("range args: want 3, got %d", len(rng.Args))
	}
	if step := as[*UnaryOp](t, rng.Args[2]); step.Op != "USub" {
		t.Errorf("range step: got %q", step.Op)
	}
	as[*Continue](t, fr.Body[0])

	if f := as[*Constant](t, as[*Return](t, body[6]).Value); f.PyType != "float" || f.Value != 1.5 {
		t.Errorf("return 1.5: got %+v", f)
	}
}

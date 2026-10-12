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

func TestDecodeSequences(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/sequences.py")

	// xs: list[int] = [1, 2, 3]
	ann := as[*AnnAssign](t, mod.Body[0])
	sub := as[*Subscript](t, ann.Annotation)
	if as[*Name](t, sub.Value).ID != "list" || as[*Name](t, sub.Index).ID != "int" {
		t.Errorf("annotation list[int]: got %+v", sub)
	}
	list := as[*List](t, ann.Value)
	if len(list.Elts) != 3 {
		t.Errorf("list literal: want 3 elements, got %d", len(list.Elts))
	}

	// t: tuple[int, str] = (1, "a")
	ann = as[*AnnAssign](t, mod.Body[1])
	tupAnn := as[*Tuple](t, as[*Subscript](t, ann.Annotation).Index)
	if len(tupAnn.Elts) != 2 {
		t.Errorf("tuple[int, str] index: want 2 elements, got %d", len(tupAnn.Elts))
	}
	as[*Tuple](t, ann.Value)

	// a, b = t
	asg := as[*Assign](t, mod.Body[2])
	if tgt := as[*Tuple](t, asg.Targets[0]); len(tgt.Elts) != 2 {
		t.Errorf("unpacking target: got %+v", tgt)
	}

	// y = xs[1:], xs[::-1], xs[0], "abc".upper()
	elts := as[*Tuple](t, as[*Assign](t, mod.Body[3]).Value).Elts
	s1 := as[*Slice](t, as[*Subscript](t, elts[0]).Index)
	if s1.Lower == nil || s1.Upper != nil || s1.Step != nil {
		t.Errorf("xs[1:]: got %+v", s1)
	}
	s2 := as[*Slice](t, as[*Subscript](t, elts[1]).Index)
	if s2.Lower != nil || s2.Upper != nil || s2.Step == nil {
		t.Errorf("xs[::-1]: got %+v", s2)
	}
	as[*Constant](t, as[*Subscript](t, elts[2]).Index)
	m := as[*Attribute](t, as[*Call](t, elts[3]).Func)
	if m.Attr != "upper" {
		t.Errorf("method name: want upper, got %q", m.Attr)
	}

	// print(f"{a!r:>5} and {{b}}", sep="-", end="")
	call := as[*Call](t, as[*Expr](t, mod.Body[4]).Value)
	if len(call.Keywords) != 2 || call.Keywords[0].Arg != "sep" || call.Keywords[1].Arg != "end" || !call.HasKeywords {
		t.Errorf("keywords: got %+v", call.Keywords)
	}
	if kw := as[*Constant](t, call.Keywords[0].Value); kw.Value != "-" {
		t.Errorf("sep value: got %+v", kw)
	}
	fstr := as[*JoinedStr](t, call.Args[0])
	fv := as[*FormattedValue](t, fstr.Values[0])
	if fv.Conversion != 'r' {
		t.Errorf("conversion: want 'r', got %d", fv.Conversion)
	}
	spec := as[*JoinedStr](t, fv.FormatSpec)
	if c := as[*Constant](t, spec.Values[0]); c.Value != ">5" {
		t.Errorf("format spec: got %+v", c)
	}
	if text := as[*Constant](t, fstr.Values[1]); text.Value != " and {b}" {
		t.Errorf("literal text: got %+v", text)
	}
}

func TestDecodeExceptions(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/exceptions.py")

	try := as[*Try](t, mod.Body[0])
	if try.Line() != 1 || len(try.Handlers) != 2 || len(try.Orelse) != 1 || len(try.Finalbody) != 1 {
		t.Fatalf("Try: got %d handlers, %d else, %d finally", len(try.Handlers), len(try.Orelse), len(try.Finalbody))
	}

	raise := as[*Raise](t, try.Body[0])
	call := as[*Call](t, raise.Exc)
	if as[*Name](t, call.Func).ID != "ValueError" {
		t.Errorf("raise target: got %+v", call.Func)
	}
	if c := as[*Constant](t, raise.Cause); c.PyType != "none" {
		t.Errorf("raise ... from None: got cause %+v", c)
	}

	h := try.Handlers[0]
	if h.Name != "e" || h.Line() != 3 {
		t.Errorf("handler: name %q line %d", h.Name, h.Line())
	}
	if types := as[*Tuple](t, h.Type); len(types.Elts) != 2 {
		t.Errorf("except (A, B): got %+v", types)
	}
	if bare := as[*Raise](t, h.Body[0]); bare.Exc != nil || bare.Cause != nil {
		t.Errorf("bare raise must have no exc or cause, got %+v", bare)
	}

	bare := try.Handlers[1]
	if bare.Type != nil || bare.Name != "" {
		t.Errorf("bare except: got %+v", bare)
	}

	as_ := as[*Assert](t, try.Orelse[0])
	if as_.Msg == nil {
		t.Error("assert message must be kept")
	}
	as[*Compare](t, as_.Test)

	// try/finally without any handler
	tf := as[*Try](t, mod.Body[1])
	if len(tf.Handlers) != 0 || len(tf.Finalbody) != 1 {
		t.Errorf("try/finally: got %+v", tf)
	}
}

func TestDecodeContainers(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/containers.py")
	value := func(i int) Node { return as[*Assign](t, mod.Body[i]).Value }

	d := as[*Dict](t, value(0))
	if len(d.Keys) != 2 || d.Keys[0] == nil || d.Keys[1] != nil {
		t.Errorf("dict keys: a nil key must mark **mapping, got %+v", d.Keys)
	}
	if as[*Name](t, d.Values[1]).ID != "e" {
		t.Errorf("dict values: got %+v", d.Values)
	}

	if s := as[*Set](t, value(1)); len(s.Elts) != 2 {
		t.Errorf("set: got %+v", s)
	}

	lc := as[*ListComp](t, value(2))
	if len(lc.Generators) != 1 || len(lc.Generators[0].Ifs) != 2 {
		t.Fatalf("list comprehension: got %+v", lc.Generators)
	}
	if as[*Name](t, lc.Generators[0].Target).ID != "x" || lc.Generators[0].IsAsync {
		t.Errorf("generator: got %+v", lc.Generators[0])
	}
	as[*BinOp](t, lc.Elt)

	as[*SetComp](t, value(3))
	dc := as[*DictComp](t, value(4))
	if as[*Name](t, dc.Key).ID != "k" || as[*Name](t, dc.Value).ID != "v" {
		t.Errorf("dict comprehension: got %+v %+v", dc.Key, dc.Value)
	}
	if as[*Tuple](t, dc.Generators[0].Target); dc.Line() != 5 {
		t.Errorf("dict comprehension line: got %d", dc.Line())
	}
	ge := as[*GeneratorExp](t, value(5))
	if len(ge.Generators) != 2 {
		t.Errorf("generator expression: want 2 clauses, got %d", len(ge.Generators))
	}

	del := as[*Delete](t, mod.Body[6])
	if len(del.Targets) != 2 {
		t.Errorf("del: got %+v", del.Targets)
	}
	as[*Subscript](t, del.Targets[0])

	fn := as[*FunctionDef](t, mod.Body[7])
	g := as[*Global](t, fn.Body[0])
	if len(g.Names) != 2 || g.Names[0] != "a" || g.Names[1] != "b" {
		t.Errorf("global: got %+v", g.Names)
	}
}

func TestDecodeFunctions(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/functions.py")

	fn := as[*FunctionDef](t, mod.Body[0])
	if len(fn.Decorators) != 2 || !fn.HasDecorators {
		t.Fatalf("decorators: got %d", len(fn.Decorators))
	}
	if as[*Name](t, fn.Decorators[0]).ID != "dec" {
		t.Errorf("first decorator: got %+v", fn.Decorators[0])
	}
	if call := as[*Call](t, fn.Decorators[1]); as[*Name](t, call.Func).ID != "dec2" || len(call.Args) != 1 {
		t.Errorf("decorator with arguments: got %+v", call)
	}
	a := fn.Args
	if len(a.PosOnly) != 1 || a.PosOnly[0].Name != "a" || len(a.Args) != 1 || a.Args[0].Name != "b" {
		t.Errorf("positional parameters: %+v %+v", a.PosOnly, a.Args)
	}
	if len(a.Defaults) != 1 {
		t.Errorf("defaults: want 1, got %d", len(a.Defaults))
	}
	if len(a.Kwonly) != 2 || len(a.KwDefaults) != 2 || a.KwDefaults[0] != nil || a.KwDefaults[1] == nil {
		t.Errorf("keyword-only defaults must stay aligned with the parameters: %+v", a.KwDefaults)
	}
	if a.Vararg == nil || a.Vararg.Name != "args" || a.Kwarg == nil || a.Kwarg.Name != "kw" {
		t.Errorf("*args and **kw: %+v %+v", a.Vararg, a.Kwarg)
	}

	outer := as[*FunctionDef](t, mod.Body[1])
	inner := as[*FunctionDef](t, outer.Body[1])
	nl := as[*Nonlocal](t, inner.Body[0])
	if len(nl.Names) != 1 || nl.Names[0] != "n" {
		t.Errorf("nonlocal: %+v", nl)
	}
	lam := as[*Lambda](t, as[*Return](t, outer.Body[2]).Value)
	if len(lam.Args.Args) != 2 || len(lam.Args.Defaults) != 1 {
		t.Errorf("lambda parameters: %+v", lam.Args)
	}
	as[*BinOp](t, lam.Body)

	call := as[*Call](t, as[*Expr](t, mod.Body[2]).Value)
	if len(call.Args) != 3 {
		t.Fatalf("call args: got %d", len(call.Args))
	}
	as[*Starred](t, call.Args[0])
	as[*Constant](t, call.Args[1])
	as[*Starred](t, call.Args[2])
	if len(call.Keywords) != 2 || call.Keywords[0].Arg != "key" || call.Keywords[1].Arg != "" {
		t.Errorf("keywords: a '**mapping' has an empty name, got %+v", call.Keywords)
	}
}

func TestDecodeClasses(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/classes.py")

	p := as[*ClassDef](t, mod.Body[1])
	if p.Name != "P" || len(p.Bases) != 1 || len(p.Decorators) != 1 || len(p.Body) != 2 {
		t.Fatalf("class P: %+v", p)
	}
	if as[*Name](t, p.Decorators[0]).ID != "dataclass" {
		t.Errorf("decorator: %+v", p.Decorators[0])
	}
	as[*AnnAssign](t, p.Body[0])
	as[*FunctionDef](t, p.Body[1])

	q := as[*ClassDef](t, mod.Body[2])
	if len(q.Bases) != 1 || len(q.Keywords) != 1 || q.Keywords[0].Arg != "metaclass" {
		t.Errorf("class Q: bases %d keywords %+v", len(q.Bases), q.Keywords)
	}
}

func TestDecodePhase8(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/phase8.py")
	as[*IfExp](t, as[*Assign](t, mod.Body[0]).Value)
	as[*NamedExpr](t, as[*Expr](t, mod.Body[1]).Value)
	w := as[*With](t, mod.Body[2])
	if len(w.Items) != 2 || w.Items[0].Vars == nil || w.Items[1].Vars != nil {
		t.Errorf("with items: %+v", w.Items)
	}
	fn := as[*FunctionDef](t, mod.Body[3])
	as[*Yield](t, as[*Expr](t, fn.Body[0]).Value)
	as[*Yield](t, as[*Assign](t, fn.Body[1]).Value)
	as[*YieldFrom](t, as[*Expr](t, fn.Body[2]).Value)
	m := as[*Match](t, mod.Body[4])
	if len(m.Cases) != 6 {
		t.Fatalf("cases: %d", len(m.Cases))
	}
	as[*MatchOr](t, m.Cases[0].Pattern)
	seq := as[*MatchSequence](t, m.Cases[1].Pattern)
	as[*MatchStar](t, seq.Patterns[1])
	if m.Cases[1].Guard == nil {
		t.Error("guard lost")
	}
	if mm := as[*MatchMapping](t, m.Cases[2].Pattern); mm.Rest != "kw" || len(mm.Keys) != 1 {
		t.Errorf("mapping: %+v", mm)
	}
	as_ := as[*MatchAs](t, m.Cases[3].Pattern)
	if cl := as[*MatchClass](t, as_.Pattern); len(cl.Patterns) != 1 || cl.KwdAttrs[0] != "y" {
		t.Errorf("class pattern: %+v", cl)
	}
	if s := as[*MatchSingleton](t, m.Cases[4].Pattern); s.Value != "None" {
		t.Errorf("singleton: %+v", s)
	}
	if w := as[*MatchAs](t, m.Cases[5].Pattern); w.Name != "" || w.Pattern != nil {
		t.Errorf("wildcard: %+v", w)
	}
}

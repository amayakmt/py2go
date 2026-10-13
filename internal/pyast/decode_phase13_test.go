package pyast

import "testing"

func TestDecodeAsync(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/async.py")
	fn := as[*FunctionDef](t, mod.Body[0])
	if !fn.IsAsync || fn.Name != "fetch" {
		t.Fatalf("async def: %+v", fn)
	}
	with := as[*With](t, fn.Body[0])
	if !with.IsAsync {
		t.Error("async with is not marked")
	}
	loop := as[*For](t, with.Body[0])
	if !loop.IsAsync {
		t.Error("async for is not marked")
	}
	call := as[*Call](t, as[*Expr](t, loop.Body[0]).Value.(*Await).Value)
	if as[*Name](t, call.Func).ID != "handle" {
		t.Errorf("await of %+v", call)
	}
	ret := as[*Return](t, fn.Body[1])
	comp := as[*ListComp](t, ret.Value)
	if len(comp.Generators) != 1 || !comp.Generators[0].IsAsync {
		t.Error("async comprehension is not marked")
	}
}

func TestDecodeSyncIsNotAsync(t *testing.T) {
	mod := decodeFile(t, "../../testdata/decode/typed_core.py")
	if fn, ok := mod.Body[0].(*FunctionDef); ok && fn.IsAsync {
		t.Error("a def is not async")
	}
}

package runtime

import (
	"reflect"
	"testing"
)

func TestOpt(t *testing.T) {
	some, none := Some(int64(5)), NoneOf[int64]()
	if !some.IsSome() || none.IsSome() || some.Get() != 5 {
		t.Error("Some and None")
	}
	if Repr(some) != "5" || Repr(none) != "None" || Str(none) != "None" || Repr(Some("a")) != "'a'" {
		t.Errorf("repr: %s %s", Repr(some), Repr(none))
	}
	if !Eq(some, Some(int64(5))) || Eq(some, none) || !Eq(none, NoneOf[int64]()) || !Eq(some, int64(5)) || Eq(none, int64(0)) {
		t.Error("equality of optionals")
	}
	if !Eq(none, NoneV) || !Eq(NoneV, none) || Eq(NoneV, some) {
		t.Error("None equals None")
	}
	if OptTruth(none) || OptTruth(Some(int64(0))) || !OptTruth(Some(int64(1))) || !Truth(Some("x")) || Truth(NoneV) {
		t.Error("truth of optionals")
	}
	if got := catch(func() { none.UnwrapAttr("x") }); got != "AttributeError: 'NoneType' object has no attribute 'x'" {
		t.Errorf("got %q", got)
	}
	if got := catch(func() { none.Unwrap("TypeError", "'NoneType' object is not iterable") }); got != "TypeError: 'NoneType' object is not iterable" {
		t.Errorf("got %q", got)
	}
	if Hash(none) != Hash(NoneOf[string]()) || Hash(some) != Hash(int64(5)) {
		t.Error("hash of optionals")
	}
}

func drainGen[Y, S, R any](g *Gen[Y, S, R]) []Y {
	var out []Y
	for c := g.Iter(); c.Next(); {
		out = append(out, c.Value())
	}
	return out
}

func TestGenerator(t *testing.T) {
	g := NewGen(func(y *Yielder[int64, NoneT]) NoneT {
		for i := int64(0); i < 3; i++ {
			y.Yield(i)
		}
		return NoneV
	})
	if got := drainGen(g); !reflect.DeepEqual(got, []int64{0, 1, 2}) {
		t.Errorf("got %v", got)
	}
	if got := drainGen(g); len(got) != 0 {
		t.Errorf("an exhausted generator is empty, got %v", got)
	}
	// send, and the return value in StopIteration.
	acc := NewGen(func(y *Yielder[int64, int64]) string {
		var total int64
		for {
			got := y.Yield(total)
			if got < 0 {
				return "done"
			}
			total += got
		}
	})
	if acc.NextValue() != 0 || acc.Send(5) != 5 || acc.Send(10) != 15 {
		t.Error("send")
	}
	e := Catch(func() { acc.Send(-1) })
	if e == nil || e.Type != "StopIteration" || StopValue(e) != "done" {
		t.Errorf("StopIteration carries the return value, got %v", e)
	}
	// close runs the finally block; throw can be caught inside.
	var log []string
	c := NewGen(func(y *Yielder[int64, NoneT]) NoneT {
		defer func() { log = append(log, "cleanup") }()
		y.Yield(1)
		y.Yield(2)
		return NoneV
	})
	c.NextValue()
	c.Close()
	if !reflect.DeepEqual(log, []string{"cleanup"}) || Catch(func() { c.NextValue() }) == nil {
		t.Errorf("close: %v", log)
	}
	th := NewGen(func(y *Yielder[int64, NoneT]) NoneT {
		for {
			if e := Catch(func() { y.Yield(1) }); e != nil {
				y.Yield(99)
			}
		}
	})
	th.NextValue()
	if got := th.Throw(newError("ValueError", "x")); got != 99 {
		t.Errorf("throw: %d", got)
	}
	// PEP 479
	bad := NewGen(func(y *Yielder[int64, NoneT]) NoneT {
		y.Yield(1)
		panic(newError("StopIteration", ""))
	})
	bad.NextValue()
	if e := Catch(func() { bad.NextValue() }); e == nil || e.Type != "RuntimeError" || e.str() != "generator raised StopIteration" {
		t.Errorf("a StopIteration inside a generator becomes RuntimeError, got %v", e)
	}
	// a generator that is running cannot be resumed
	var self *Gen[int64, NoneT, NoneT]
	self = NewGen(func(y *Yielder[int64, NoneT]) NoneT {
		self.NextValue()
		return NoneV
	})
	if e := Catch(func() { self.NextValue() }); e == nil || e.str() != "generator already executing" {
		t.Errorf("reentrancy, got %v", e)
	}
}

func TestYieldFrom(t *testing.T) {
	inner := func() *Gen[int64, NoneT, string] {
		return NewGen(func(y *Yielder[int64, NoneT]) string {
			y.Yield(1)
			y.Yield(2)
			return "inner"
		})
	}
	var result string
	outer := NewGen(func(y *Yielder[int64, NoneT]) NoneT {
		result = YieldFromGen(y, inner())
		YieldFromIter(y, NewList[int64](10, 20))
		return NoneV
	})
	if got := drainGen(outer); !reflect.DeepEqual(got, []int64{1, 2, 10, 20}) || result != "inner" {
		t.Errorf("got %v %q", got, result)
	}
}

func TestSplitStar(t *testing.T) {
	v1, v2 := newError("ValueError", "a"), newError("ValueError", "b")
	k := newError("KeyError", "k")
	eg := NewGroup("ExceptionGroup", "g", []*PyError{v1, k, v2})
	m, rest := SplitClasses(eg, "ValueError")
	if m == nil || len(m.Excs) != 2 || rest == nil || len(rest.Excs) != 1 || rest.Excs[0] != k || m.Msg != "g" {
		t.Errorf("split: %v %v", m, rest)
	}
	if got := eg.str(); got != "g (3 sub-exceptions)" {
		t.Errorf("str: %q", got)
	}
	if got := Repr(eg); got != "ExceptionGroup('g', [ValueError('a'), KeyError('k'), ValueError('b')])" {
		t.Errorf("repr: %s", got)
	}
	// A plain exception is wrapped when it matches, and passes unchanged when not.
	w, r := SplitStar(v1, "ValueError")
	if w == nil || !w.IsGroup || r != nil || w.Excs[0] != v1 {
		t.Errorf("wrap: %v %v", w, r)
	}
	if w, r := SplitStar(k, "ValueError"); w != nil || r != k {
		t.Errorf("no match: %v %v", w, r)
	}
	nested := NewGroup("ExceptionGroup", "outer", []*PyError{eg, newError("OSError", "o")})
	m, rest = SplitClasses(nested, "ValueError")
	if m == nil || !m.IsGroup || len(m.Excs) != 1 || len(m.Excs[0].Excs) != 2 || rest == nil || len(rest.Excs) != 2 {
		t.Errorf("nested split: %v %v", m, rest)
	}
	if !IsSubclass("ExceptionGroup", "BaseExceptionGroup") || !IsSubclass("ExceptionGroup", "Exception") || IsSubclass("BaseExceptionGroup", "Exception") {
		t.Error("group hierarchy")
	}
	if e := Catch(func() { NewGroup("ExceptionGroup", "x", nil) }); e == nil || e.Type != "ValueError" {
		t.Errorf("empty group, got %v", e)
	}
}

func TestIterHelpers(t *testing.T) {
	l := NewList[int64](1, 2, 3)
	it := IterOf[int64](l)
	if it.NextValue() != 1 || NextOr[int64](it, -1) != 2 || NextOr[int64](it, -1) != 3 || NextOr[int64](it, -1) != -1 {
		t.Error("next with a default")
	}
	if got := NextOpt[int64](IterOf[int64](NewList[int64]())); got.IsSome() {
		t.Error("next of nothing is None")
	}
	if e := Catch(func() { IterOf[int64](NewList[int64]()).NextValue() }); e == nil || e.Type != "StopIteration" {
		t.Errorf("got %v", e)
	}
	d := NewDict[string, int64](KV[string, int64]{"a", 1}, KV[string, int64]{"b", 2}, KV[string, int64]{"c", 3})
	rest := DictWithout(d, "a", "c")
	if rest.Len() != 1 || !rest.Has("b") || d.Len() != 3 {
		t.Error("DictWithout")
	}
	if e := Catch(func() { CheckUnpackStar(1, 2) }); e == nil || e.str() != "not enough values to unpack (expected at least 2, got 1)" {
		t.Errorf("got %v", e)
	}
	if FormatIntDyn(5, ">4") != "   5" || FormatIntDyn(5, ".2f") != "5.00" {
		t.Error("FormatIntDyn")
	}
	if e := Catch(func() { FormatOther(l, ">3") }); e == nil || e.str() != "unsupported format string passed to list.__format__" {
		t.Errorf("got %v", e)
	}
}

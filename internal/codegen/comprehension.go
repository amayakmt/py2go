package codegen

import (
	"fmt"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// A comprehension is an expression, but its loops are statements, so it
// compiles to an immediately called function literal:
//
//	func() *runtime.List[int64] {
//		var out = runtime.NewList[int64]()
//		<for and if statements that append to out>
//		return out
//	}()
//
// The loop variables are declared inside that function, which is why they do
// not leak, and the closure sees every variable of the enclosing function.

// enterComprehension opens the scope of a comprehension.
func (g *generator) enterComprehension() (leave func()) {
	outerEnv, outerRoot := g.env, g.compRoot
	g.env = newScope(outerEnv)
	g.compRoot = g.env
	return func() { g.env, g.compRoot = outerEnv, outerRoot }
}

// loopNest wraps an innermost action in the for and if statements of the
// comprehension's clauses.
func loopNest(gens []*pyast.Comprehension, inner pyast.Node) (pyast.Node, error) {
	node := inner
	for i := len(gens) - 1; i >= 0; i-- {
		gen := gens[i]
		for j := len(gen.Ifs) - 1; j >= 0; j-- {
			node = &pyast.If{Test: gen.Ifs[j], Body: []pyast.Node{node}, Lineno: gen.Ifs[j].Line()}
		}
		node = &pyast.For{Target: gen.Target, Iter: gen.Iter, Body: []pyast.Node{node}, IsAsync: gen.IsAsync, Lineno: gen.Target.Line()}
	}
	return node, nil
}

// eagerComp generates the loops of a list, set or dict comprehension.
// add runs inside the innermost loop and returns the Go statement that
// stores the current item.
func (g *generator) eagerComp(gens []*pyast.Comprehension, add func(b *strings.Builder) error) (string, error) {
	leave := g.enterComprehension()
	defer leave()
	nest, err := loopNest(gens, &compAction{run: add})
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := g.stmt(&b, nest); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (g *generator) listComp(n *pyast.ListComp, want pyType) (value, error) {
	var elemWant pyType
	if want.isList() {
		elemWant = want.elem()
	}
	out := g.tmp("out")
	var elem pyType
	code, err := g.eagerComp(n.Generators, func(b *strings.Builder) error {
		v, err := g.exprAs(n.Elt, elemWant)
		if err != nil {
			return err
		}
		elem = v.typ
		fmt.Fprintf(b, "%s.Append(%s)\n", out, v.code)
		return nil
	})
	if err != nil {
		return value{}, err
	}
	if elem == "" {
		return value{}, unsupported("comprehension whose element could not be compiled", n.Line())
	}
	t := listOf(elem)
	return g.literalFunc(t, out, fmt.Sprintf("%s[%s]()", g.rt("NewList"), g.goType(elem)), code), nil
}

func (g *generator) setComp(n *pyast.SetComp, want pyType) (value, error) {
	var elemWant pyType
	if want.isSet() {
		elemWant = want.elem()
	}
	out := g.tmp("out")
	var elem pyType
	code, err := g.eagerComp(n.Generators, func(b *strings.Builder) error {
		v, err := g.exprAs(n.Elt, elemWant)
		if err != nil {
			return err
		}
		if err := g.hashableOrError(v.typ, "set element", n.Elt.Line()); err != nil {
			return err
		}
		elem = v.typ
		fmt.Fprintf(b, "%s.Add(%s)\n", out, v.code)
		return nil
	})
	if err != nil {
		return value{}, err
	}
	if elem == "" {
		return value{}, unsupported("comprehension whose element could not be compiled", n.Line())
	}
	elem = g.keyType(elem)
	t := setOf(elem)
	return g.literalFunc(t, out, fmt.Sprintf("%s[%s]()", g.rt("NewSet"), g.goType(elem)), code), nil
}

func (g *generator) dictComp(n *pyast.DictComp, want pyType) (value, error) {
	var keyWant, valWant pyType
	if want.isDict() {
		keyWant, valWant = want.keyType(), want.valType()
	}
	out := g.tmp("out")
	var key, val pyType
	code, err := g.eagerComp(n.Generators, func(b *strings.Builder) error {
		k, err := g.exprAs(n.Key, keyWant)
		if err != nil {
			return err
		}
		v, err := g.exprAs(n.Value, valWant)
		if err != nil {
			return err
		}
		if err := g.hashableOrError(k.typ, "dict key", n.Key.Line()); err != nil {
			return err
		}
		key, val = k.typ, v.typ
		fmt.Fprintf(b, "%s.Set(%s, %s)\n", out, k.code, v.code)
		return nil
	})
	if err != nil {
		return value{}, err
	}
	if key == "" {
		return value{}, unsupported("comprehension whose element could not be compiled", n.Line())
	}
	key = g.keyType(key)
	t := dictOf(key, val)
	return g.literalFunc(t, out, fmt.Sprintf("%s[%s]()", g.rt("NewDict"), g.dictGoArgs(t)), code), nil
}

// literalFunc assembles the function literal of an eager comprehension.
func (g *generator) literalFunc(t pyType, out, ctor, loops string) value {
	return value{
		code: fmt.Sprintf("func() %s {\nvar %s %s = %s\n%sreturn %s\n}()", g.goType(t), out, g.goType(t), ctor, loops, out),
		typ:  t, prec: precAtom,
	}
}

// genExp is a generator expression. It is lazy: it produces one item per
// call, so any(f(x) for x in xs) stops calling f at the first true result.
// The clauses become a small state machine. level is the clause being
// advanced, and each clause has a cursor over its own iterable:
//
//	func() *runtime.Iterator[T] {
//		var x int64
//		var _c0 runtime.Cursor[int64] = xs.Iter()
//		var _c1 runtime.Cursor[int64]
//		_level := 0
//		return runtime.NewIterator(func() (T, bool) {
//			for {
//				switch _level {
//				case 0: ...advance the first clause, enter the second...
//				case 1: ...advance the second clause, yield...
//				}
//			}
//		})
//	}()
func (g *generator) genExp(n *pyast.GeneratorExp) (value, error) {
	leave := g.enterComprehension()
	defer leave()

	type clause struct {
		cursorType string // Go type of the cursor variable
		cursor     string // expression that creates it
		bind       string // statements that assign the target
		conds      []string
	}
	var decls, inner strings.Builder
	clauses := make([]clause, len(n.Generators))

	for i, gen := range n.Generators {
		if gen.IsAsync {
			return value{}, unsupported("async generator expression", gen.Target.Line())
		}
		src, err := g.iterOf(gen.Iter)
		if err != nil {
			return value{}, err
		}
		bound, err := leaves(gen.Target, src.typ, gen.Target.Line())
		if err != nil {
			return value{}, err
		}
		for _, l := range bound {
			if err := checkName(l.name, gen.Target.Line()); err != nil {
				return value{}, err
			}
			g.declareLocal(&decls, l.name, l.typ, "")
		}
		c := clause{cursorType: fmt.Sprintf("%s[%s]", g.rt("Cursor"), g.goType(src.typ)), cursor: src.cursor}
		var bind strings.Builder
		item := fmt.Sprintf("_c%d.Value()", i)
		switch tg := gen.Target.(type) {
		case *pyast.Name:
			err = g.assignTarget(&bind, tg, item, src.typ, tg.Line())
		default:
			tmp := g.tmp("e")
			g.temp(&bind, tmp, src.typ, item)
			err = g.assignTarget(&bind, gen.Target, tmp, src.typ, gen.Target.Line())
		}
		if err != nil {
			return value{}, err
		}
		c.bind = bind.String()
		for _, cond := range gen.Ifs {
			v, err := g.condition(cond)
			if err != nil {
				return value{}, err
			}
			c.conds = append(c.conds, v.code)
		}
		clauses[i] = c
	}

	elt, err := g.typedExpr(n.Elt)
	if err != nil {
		return value{}, err
	}
	if elt.typ == tExc {
		return value{}, unsupported("exception stored in a generator", n.Elt.Line())
	}
	T := g.goType(elt.typ)

	// The first iterable is evaluated when the generator is created; the
	// others each time their clause is entered.
	for i, c := range clauses {
		if i == 0 {
			fmt.Fprintf(&inner, "var _c0 %s = %s\n", c.cursorType, c.cursor)
		} else {
			fmt.Fprintf(&inner, "var _c%d %s\n", i, c.cursorType)
		}
	}
	inner.WriteString("_level := 0\n")
	fmt.Fprintf(&inner, "return %s(func() (%s, bool) {\nfor {\nswitch _level {\n", g.rt("NewIterator"), T)
	for i, c := range clauses {
		fmt.Fprintf(&inner, "case %d:\n", i)
		if i == 0 {
			fmt.Fprintf(&inner, "if !_c0.Next() {\nvar zero %s\nreturn zero, false\n}\n", T)
		} else {
			fmt.Fprintf(&inner, "if !_c%d.Next() {\n_level = %d\ncontinue\n}\n", i, i-1)
		}
		inner.WriteString(c.bind)
		for _, cond := range c.conds {
			fmt.Fprintf(&inner, "if !(%s) {\ncontinue\n}\n", cond)
		}
		if i == len(clauses)-1 {
			fmt.Fprintf(&inner, "return %s, true\n", elt.code)
		} else {
			fmt.Fprintf(&inner, "_c%d = %s\n_level = %d\n", i+1, clauses[i+1].cursor, i+1)
		}
	}
	inner.WriteString("}\n}\n})\n")

	t := iterOf(elt.typ)
	return value{
		code: fmt.Sprintf("func() %s {\n%s%s}()", g.goType(t), decls.String(), inner.String()),
		typ:  t, prec: precAtom,
	}, nil
}

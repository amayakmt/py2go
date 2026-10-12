package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// listLit generates [a, b, c]. The type comes from the elements, or from
// want for an empty list, since [] says nothing about what it will hold.
func (g *generator) listLit(n *pyast.List, want pyType) (value, error) {
	var elemWant pyType
	if want.isList() {
		elemWant = want.elem()
	}
	if len(n.Elts) == 0 {
		if !want.isList() {
			return value{}, unsupported("empty list without a type annotation", n.Line())
		}
		return value{code: g.rt("NewList") + "[" + g.goType(elemWant) + "]()", typ: want, prec: precAtom}, nil
	}

	codes := make([]string, len(n.Elts))
	var elem pyType
	for i, e := range n.Elts {
		v, err := g.exprAs(e, elemWant)
		if err != nil {
			return value{}, err
		}
		if v.typ == tExc {
			return value{}, unsupported("exception stored in a list", e.Line())
		}
		if i == 0 {
			elem = v.typ
		} else if v.typ != elem {
			return value{}, unsupported(fmt.Sprintf("list with mixed element types (%s and %s)", elem, v.typ), e.Line())
		}
		codes[i] = v.code
	}
	return value{
		code: g.rt("NewList") + "[" + g.goType(elem) + "](" + strings.Join(codes, ", ") + ")",
		typ:  listOf(elem), prec: precAtom,
	}, nil
}

// tupleLit generates (a, b). Tuples keep the type of each position.
func (g *generator) tupleLit(n *pyast.Tuple, want pyType) (value, error) {
	if want.isVarTuple() {
		// A tuple display where tuple[T, ...] is expected.
		codes := make([]string, len(n.Elts))
		for i, e := range n.Elts {
			v, err := g.exprAs(e, want.elem())
			if err != nil {
				return value{}, err
			}
			v, err = g.coerce(v, want.elem(), "tuple item", e.Line())
			if err != nil {
				return value{}, err
			}
			codes[i] = v.code
		}
		return value{code: g.rt("NewVarTuple") + "[" + g.goType(want.elem()) + "](" + strings.Join(codes, ", ") + ")", typ: want, prec: precAtom}, nil
	}
	if len(n.Elts) == 0 {
		return value{}, unsupported("empty tuple", n.Line())
	}
	if len(n.Elts) > maxTupleArity {
		return value{}, unsupported(fmt.Sprintf("tuple with more than %d elements", maxTupleArity), n.Line())
	}
	var wants []pyType
	if want.isTuple() && len(want.elems()) == len(n.Elts) {
		wants = want.elems()
	}

	types := make([]pyType, len(n.Elts))
	codes := make([]string, len(n.Elts))
	for i, e := range n.Elts {
		var w pyType
		if wants != nil {
			w = wants[i]
		}
		v, err := g.exprAs(e, w)
		if err != nil {
			return value{}, err
		}
		if v.typ == tExc {
			return value{}, unsupported("exception stored in a tuple", e.Line())
		}
		types[i], codes[i] = v.typ, v.code
	}
	t := tupleOf(types)
	return value{code: g.tupleCtor(t, codes), typ: t, prec: precAtom}, nil
}

// tupleCtor builds a tuple of type t from element expressions. The type
// arguments are spelled out so that a constant such as 1 becomes an int64
// field, not an int.
func (g *generator) tupleCtor(t pyType, codes []string) string {
	n := len(codes)
	typeArgs := strings.TrimSuffix(strings.TrimPrefix(g.goType(t), fmt.Sprintf("runtime.Tuple%d[", n)), "]")
	return fmt.Sprintf("%s%d[%s](%s)", g.rt("Tup"), n, typeArgs, strings.Join(codes, ", "))
}

// subscript generates x[i] and x[a:b:c] on str, list and tuple.
func (g *generator) subscript(n *pyast.Subscript) (value, error) {
	base, err := g.typedExpr(n.Value)
	if err != nil {
		return value{}, err
	}

	if sl, ok := n.Index.(*pyast.Slice); ok {
		if base.typ != tStr && !base.typ.isList() && !base.typ.isVarTuple() {
			return value{}, unsupported("slice of "+string(base.typ), n.Line())
		}
		bounds, err := g.bounds(sl)
		if err != nil {
			return value{}, err
		}
		switch {
		case base.typ == tStr:
			return value{code: call(g.rt("StrSlice"), base.code, bounds), typ: tStr, prec: precAtom}, nil
		case base.typ.isList() || base.typ.isVarTuple():
			return value{code: base.at(precAtom) + ".Slice(" + bounds + ")", typ: base.typ, prec: precAtom}, nil
		}
		return value{}, unsupported("slice of "+string(base.typ), n.Line())
	}

	if base.typ.isDict() {
		key, err := g.exprAs(n.Index, base.typ.keyType())
		if err != nil {
			return value{}, err
		}
		if key.typ != base.typ.keyType() {
			return value{}, mismatch("dict key", base.typ.keyType(), key.typ, n.Index.Line())
		}
		return value{code: base.at(precAtom) + ".Get(" + key.code + ")", typ: base.typ.valType(), prec: precAtom}, nil
	}
	idx, err := g.typedExpr(n.Index)
	if err != nil {
		return value{}, err
	}
	if idx.typ != tInt {
		return value{}, unsupported(fmt.Sprintf("index of type %s", idx.typ), n.Line())
	}
	switch {
	case base.typ == tStr:
		return value{code: call(g.rt("StrIndex"), base.code, idx.code), typ: tStr, prec: precAtom}, nil
	case base.typ.isList() || base.typ.isVarTuple():
		return value{code: base.at(precAtom) + ".Get(" + idx.code + ")", typ: base.typ.elem(), prec: precAtom}, nil
	case base.typ.isTuple():
		return g.tupleIndex(n, base, idx)
	case base.typ == tRange:
		return value{code: base.at(precAtom) + ".Get(" + idx.code + ")", typ: tInt, prec: precAtom}, nil
	}
	return value{}, unsupported("subscript of "+string(base.typ), n.Line())
}

// tupleIndex reads one field. The index has to be a literal because every
// position can have a different type.
func (g *generator) tupleIndex(n *pyast.Subscript, base, idx value) (value, error) {
	elems := base.typ.elems()
	if !idx.intLit {
		return value{}, unsupported("tuple index that is not an integer literal", n.Line())
	}
	i := idx.intVal
	if i < 0 {
		i += int64(len(elems))
	}
	if i < 0 || i >= int64(len(elems)) {
		return value{}, unsupported("tuple index out of range", n.Line())
	}
	return value{code: fmt.Sprintf("%s.F%d", base.at(precAtom), i), typ: elems[i], prec: precAtom}, nil
}

// bounds generates the runtime.Bounds value for lo:hi:step.
func (g *generator) bounds(sl *pyast.Slice) (string, error) {
	var fields []string
	for _, part := range []struct {
		name string
		node pyast.Node
	}{{"Lo", sl.Lower}, {"Hi", sl.Upper}, {"Step", sl.Step}} {
		if part.node == nil {
			continue
		}
		v, err := g.typedExpr(part.node)
		if err != nil {
			return "", err
		}
		if v.typ != tInt {
			return "", unsupported(fmt.Sprintf("slice bound of type %s", v.typ), part.node.Line())
		}
		fields = append(fields, fmt.Sprintf("%s: %s, Has%s: true", part.name, v.code, part.name))
	}
	return g.rt("Bounds") + "{" + strings.Join(fields, ", ") + "}", nil
}

// fstring generates an f-string as a Go string concatenation. Replacement
// fields turn into runtime calls chosen by the static type of the value.
func (g *generator) fstring(n *pyast.JoinedStr) (value, error) {
	var parts []string
	for _, piece := range n.Values {
		switch p := piece.(type) {
		case *pyast.Constant:
			text, _ := p.Value.(string)
			parts = append(parts, strconv.Quote(text))
		case *pyast.FormattedValue:
			code, err := g.formatField(p)
			if err != nil {
				return value{}, err
			}
			parts = append(parts, code)
		default:
			return value{}, unsupported(nodeName(piece), piece.Line())
		}
	}
	switch len(parts) {
	case 0:
		return value{code: `""`, typ: tStr, prec: precAtom, konst: true}, nil
	case 1:
		return value{code: parts[0], typ: tStr, prec: precAtom}, nil
	}
	return value{code: strings.Join(parts, " + "), typ: tStr, prec: precAdd}, nil
}

func (g *generator) formatField(f *pyast.FormattedValue) (string, error) {
	spec := ""
	if f.FormatSpec != nil {
		js, ok := f.FormatSpec.(*pyast.JoinedStr)
		if !ok || len(js.Values) > 1 {
			return "", unsupported("replacement field inside a format spec", f.Line())
		}
		if len(js.Values) == 1 {
			c, ok := js.Values[0].(*pyast.Constant)
			if !ok {
				return "", unsupported("replacement field inside a format spec", f.Line())
			}
			spec, _ = c.Value.(string)
		}
	}

	v, err := g.typedExpr(f.Value)
	if err != nil {
		return "", err
	}

	// !s and !r convert first, then format the resulting string.
	switch f.Conversion {
	case -1:
	case 's':
		v = value{code: call(g.rt("Str"), v.code), typ: tStr, prec: precAtom}
	case 'r':
		v = value{code: call(g.rt("Repr"), v.code), typ: tStr, prec: precAtom}
	default:
		return "", unsupported("!a conversion in an f-string", f.Line())
	}

	reject := func(err error) (string, error) {
		return "", unsupported(fmt.Sprintf("format spec '%s': %v", spec, err), f.Line())
	}
	quoted := strconv.Quote(spec)
	switch {
	case v.typ == tStr:
		if spec == "" {
			return v.code, nil
		}
		if err := runtime.ValidateSpec("str", spec); err != nil {
			return reject(err)
		}
		return call(g.rt("FormatStr"), v.code, quoted), nil

	case v.typ == tInt:
		if spec == "" {
			return call(g.rt("Str"), v.code), nil
		}
		if err := runtime.ValidateSpec("int", spec); err != nil {
			return reject(err)
		}
		if runtime.SpecIsFloat(spec) {
			return call(g.rt("FormatFloat"), toFloat(v).code, quoted), nil
		}
		return call(g.rt("FormatInt"), v.code, quoted), nil

	case v.typ == tFloat:
		if spec == "" {
			return call(g.rt("Str"), v.code), nil
		}
		if err := runtime.ValidateSpec("float", spec); err != nil {
			return reject(err)
		}
		return call(g.rt("FormatFloat"), v.code, quoted), nil
	}

	// bool, list and tuple: Python only accepts an empty spec for these
	// here (a bool with a spec formats as an int, which is not supported).
	if spec != "" {
		return "", unsupported(fmt.Sprintf("format spec '%s' for %s", spec, v.typ), f.Line())
	}
	return call(g.rt("Str"), v.code), nil
}

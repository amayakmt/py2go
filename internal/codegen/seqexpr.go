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
	if starIndex(n.Elts) >= 0 {
		return g.starredDisplay(n.Elts, elemWant, "list", n.Line())
	}
	if len(n.Elts) == 0 {
		if !want.isList() {
			want, elemWant = listOf(tObject), tObject
		}
		return value{code: g.rt("NewList") + "[" + g.goType(elemWant) + "]()", typ: want, prec: precAtom}, nil
	}

	codes := make([]string, len(n.Elts))
	vals := make([]value, len(n.Elts))
	var elem pyType
	for i, e := range n.Elts {
		v, err := g.exprAs(e, elemWant)
		if err != nil {
			return value{}, err
		}
		vals[i] = v
		if i == 0 {
			elem = v.typ
		} else if v.typ != elem {
			common, ok := g.joinTypes(elem, v.typ)
			if !ok {
				common = tObject
			}
			elem = common
		}
	}
	if elem == tNoneType {
		elem = tObject
	}
	for i, v := range vals {
		codes[i] = g.relabel(v, elem).code
	}
	return value{
		code: g.rt("NewList") + "[" + g.goType(elem) + "](" + strings.Join(codes, ", ") + ")",
		typ:  listOf(elem), prec: precAtom,
	}, nil
}

// tupleLit generates (a, b). Tuples keep the type of each position.
func (g *generator) tupleLit(n *pyast.Tuple, want pyType) (value, error) {
	if starIndex(n.Elts) >= 0 {
		var elemWant pyType
		if want.isVarTuple() {
			elemWant = want.elem()
		}
		return g.starredDisplay(n.Elts, elemWant, "tuple", n.Line())
	}
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
		return value{code: g.rt("NewVarTuple") + "[any]()", typ: varTupleOf(tObject), prec: precAtom}, nil
	}
	if len(n.Elts) > maxTupleArity {
		return g.longTuple(n)
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
	// Box[int] names a class: Box[int].create().
	if head, ok := n.Value.(*pyast.Name); ok && !g.isVariable(head.ID) && g.genericClasses[head.ID] != nil {
		t, err := g.annotationType(n, false)
		if err != nil {
			return value{}, err
		}
		c := g.classes[string(t)]
		return value{typ: classRefOf(c.name), prec: precAtom, classRef: c}, nil
	}
	base, err := g.typedExpr(n.Value)
	if err != nil {
		return value{}, err
	}
	if base.typ.isOptional() {
		base = value{code: base.at(precAtom) + ".Unwrap(\"TypeError\", \"'NoneType' object is not subscriptable\")", typ: base.typ.optElem(), prec: precAtom}
	}
	if base.typ.isClass() {
		return g.classGetItem(n, base)
	}
	if base.typ == tObject || g.indexIsObject(n.Index) {
		return g.dynSubscript(base, n.Index)
	}

	if sl, ok := n.Index.(*pyast.Slice); ok {
		if base.typ != tStr && !base.typ.isList() && !base.typ.isVarTuple() {
			if g.boxable(base) {
				return g.dynSubscript(base, n.Index)
			}
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
			if mentionsObject(key.typ) || (isNumberType(key.typ) && isNumberType(base.typ.keyType())) {
				return g.dynSubscript(base, n.Index)
			}
			return value{}, mismatch("dict key", base.typ.keyType(), key.typ, n.Index.Line())
		}
		return value{code: base.at(precAtom) + ".Get(" + key.code + ")", typ: base.typ.valType(), prec: precAtom}, nil
	}
	idx, err := g.typedExpr(n.Index)
	if err != nil {
		return value{}, err
	}
	if idx.typ == tBool {
		return g.dynSubscript(base, n.Index) // True is the index 1
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
	if g.boxable(base) {
		return g.dynSubscript(base, n.Index)
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
		if v.typ == tObject {
			if v, err = g.coerce(v, tInt, "slice bound", part.node.Line()); err != nil {
				return "", err
			}
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
	dynamic := ""
	if f.FormatSpec != nil {
		js, ok := f.FormatSpec.(*pyast.JoinedStr)
		if !ok {
			return "", unsupported("format spec of kind "+nodeName(f.FormatSpec), f.Line())
		}
		static := true
		for _, part := range js.Values {
			if _, isConst := part.(*pyast.Constant); !isConst {
				static = false
			}
		}
		if static {
			for _, part := range js.Values {
				text, _ := part.(*pyast.Constant).Value.(string)
				spec += text
			}
		} else {
			// Replacement fields inside the spec: it is built when the
			// f-string runs.
			sv, err := g.fstring(js)
			if err != nil {
				return "", err
			}
			dynamic = sv.code
		}
	}

	v, err := g.typedExpr(f.Value)
	if err != nil {
		return "", err
	}
	if dynamic != "" {
		return g.formatDynamic(f, v, dynamic)
	}

	// !s and !r convert first, then format the resulting string.
	switch f.Conversion {
	case -1:
	case 's':
		v = value{code: call(g.rt("Str"), v.code), typ: tStr, prec: precAtom}
	case 'r':
		v = value{code: call(g.rt("Repr"), v.code), typ: tStr, prec: precAtom}
	case 'a':
		v = value{code: call(g.rt("Ascii"), g.boxed(v)), typ: tStr, prec: precAtom}
	default:
		return "", unsupported("unknown conversion in an f-string", f.Line())
	}

	reject := func(err error) (string, error) {
		return "", unsupported(fmt.Sprintf("format spec '%s': %v", spec, err), f.Line())
	}
	quoted := strconv.Quote(spec)
	if c := g.classOf(v.typ); c != nil {
		if m := c.lookup("__format__"); m != nil && m.kind == mMethod && len(m.sig.pos) == 1 && (m.sig.pos[0].typ == tStr || m.sig.pos[0].typ == tObject) && m.sig.ret == tStr {
			return v.at(precAtom) + ".M___format__(" + quoted + ")", nil
		}
		if spec == "" {
			return call(g.rt("Str"), v.code), nil
		}
		return "", unsupported(fmt.Sprintf("format spec '%s' for an instance of %s (TypeError: unsupported format string passed to %s.__format__)", spec, c.name, c.name), f.Line())
	}
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

	if v.typ == tObject {
		return call(g.rt("FormatDyn"), v.code, quoted), nil
	}
	if v.typ == tBool && spec != "" {
		if err := runtime.ValidateSpec("int", spec); err != nil {
			return reject(err)
		}
		return call(g.rt("FormatInt"), call(g.rt("BoolToInt"), v.code), quoted), nil
	}

	// bool, list and tuple: Python only accepts an empty spec for these
	// here (a bool with a spec formats as an int, which is not supported).
	if spec != "" {
		return "", unsupported(fmt.Sprintf("format spec '%s' for %s", spec, v.typ), f.Line())
	}
	return call(g.rt("Str"), v.code), nil
}

// formatDynamic formats a value with a spec that is only known when the
// f-string runs, so the checks the compiler makes for a literal spec happen
// at run time.
func (g *generator) formatDynamic(f *pyast.FormattedValue, v value, spec string) (string, error) {
	switch f.Conversion {
	case 's':
		v = value{code: call(g.rt("Str"), v.code), typ: tStr, prec: precAtom}
	case 'r':
		v = value{code: call(g.rt("Repr"), v.code), typ: tStr, prec: precAtom}
	case 'a':
		v = value{code: call(g.rt("Ascii"), g.boxed(v)), typ: tStr, prec: precAtom}
	case -1:
	default:
		return "", unsupported("unknown conversion in an f-string", f.Line())
	}
	if c := g.classOf(v.typ); c != nil {
		if m := c.lookup("__format__"); m != nil && m.kind == mMethod && len(m.sig.pos) == 1 && (m.sig.pos[0].typ == tStr || m.sig.pos[0].typ == tObject) && m.sig.ret == tStr {
			return v.at(precAtom) + ".M___format__(" + spec + ")", nil
		}
	}
	switch v.typ {
	case tObject:
		return call(g.rt("FormatDyn"), v.code, spec), nil
	case tStr:
		return call(g.rt("FormatStr"), v.code, spec), nil
	case tInt:
		return call(g.rt("FormatIntDyn"), v.code, spec), nil
	case tFloat:
		return call(g.rt("FormatFloat"), v.code, spec), nil
	}
	return call(g.rt("FormatOther"), g.relabel(v, tObject).code, spec), nil
}

// longTuple builds a tuple with more items than the fixed-size tuple types
// have: it holds them as a sequence of one type, or of objects.
func (g *generator) longTuple(n *pyast.Tuple) (value, error) {
	vals := make([]value, len(n.Elts))
	var elem pyType
	for i, e := range n.Elts {
		v, err := g.typedExpr(e)
		if err != nil {
			return value{}, err
		}
		if !g.boxable(v) {
			return value{}, unsupported("tuple item of type "+string(v.typ), e.Line())
		}
		vals[i] = v
		switch {
		case i == 0:
			elem = v.typ
		case elem != v.typ:
			elem = tObject
		}
	}
	codes := make([]string, len(vals))
	for i, v := range vals {
		codes[i] = g.relabel(v, elem).typed().code
	}
	return value{code: g.rt("NewVarTuple") + "[" + g.goType(elem) + "](" + strings.Join(codes, ", ") + ")", typ: varTupleOf(elem), prec: precAtom}, nil
}

func isNumberType(t pyType) bool { return t == tInt || t == tFloat || t == tBool }

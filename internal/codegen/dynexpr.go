package codegen

import (
	"fmt"
	"github.com/amayakmt/py2go/runtime"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// Operations on values whose type is known only at run time (object) go
// through the runtime's Dyn functions, which apply Python's rules to whatever
// the values turn out to be.

// boxed is the Go code of v as an object.
func (g *generator) boxed(v value) string {
	return g.relabel(g.escape(v), tObject).typed().code
}

func (g *generator) dynValue(code string, t pyType) value {
	return value{code: code, typ: t, prec: precAtom}
}

// dynBinary is l op r where an operand is an object.
func (g *generator) dynBinary(op string, l, r value) value {
	return g.dynValue(g.rt("DynBinary")+`("`+op+`", `+g.boxed(l)+", "+g.boxed(r)+")", tObject)
}

// dynCompare is l op r for a comparison with an object operand.
func (g *generator) dynCompare(op string, l, r value) value {
	return g.dynValue(g.rt("DynCompare")+`("`+op+`", `+g.boxed(l)+", "+g.boxed(r)+")", tBool)
}

// dynUnary is op v for an object.
func (g *generator) dynUnary(op string, v value) value {
	return g.dynValue(g.rt("DynUnary")+`("`+op+`", `+g.boxed(v)+")", tObject)
}

// dynArgs boxes values for a variadic Dyn call.
func (g *generator) dynArgs(vs []value) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = g.boxed(v)
	}
	return strings.Join(parts, ", ")
}

func involvesObject(l, r value) bool { return l.typ == tObject || r.typ == tObject }

// dynBuiltin handles a call of a builtin function that has an object among its
// arguments. It reports false when no argument is an object.
func (g *generator) dynBuiltin(name string, n *pyast.Call, force bool) (value, bool, error) {
	if (name == "getattr" || name == "setattr" || name == "hasattr" || name == "delattr") && !g.isVariable(name) && g.sigs[name] == nil && !n.HasKeywords {
		args := make([]value, len(n.Args))
		for i, a := range n.Args {
			v, err := g.typedExpr(a)
			if err != nil {
				return value{}, true, err
			}
			args[i] = v
		}
		want := map[string][2]int{"getattr": {2, 3}, "setattr": {3, 3}, "hasattr": {2, 2}, "delattr": {2, 2}}[name]
		if len(args) < want[0] || len(args) > want[1] {
			return value{}, true, unsupported(fmt.Sprintf("%s() with %d arguments", name, len(args)), n.Line())
		}
		nameArg, err := g.coerce(args[1], tStr, name+"() attribute name", n.Line())
		if err != nil {
			return value{}, true, err
		}
		parts := []string{g.boxed(args[0]), nameArg.code}
		for _, extra := range args[2:] {
			parts = append(parts, g.boxed(extra))
		}
		a := strings.Join(parts, ", ")
		switch name {
		case "getattr":
			if len(args) == 3 {
				return g.dynValue(g.rt("DynGetAttrDefault")+"("+a+")", tObject), true, nil
			}
			return g.dynValue(g.rt("DynGetAttr")+"("+a+")", tObject), true, nil
		case "setattr":
			return g.dynValue(g.rt("DynSetAttr")+"("+a+")", tNone), true, nil
		case "hasattr":
			return g.dynValue(g.rt("DynHasAttr")+"("+a+")", tBool), true, nil
		}
		return g.dynValue(g.rt("DynDelAttr")+"("+a+")", tNone), true, nil
	}
	switch name {
	case "len", "int", "float", "abs", "round", "sum", "list", "tuple", "set", "dict", "sorted", "reversed",
		"enumerate", "any", "all", "min", "max", "hash", "map", "filter", "zip",
		"ord", "chr", "bin", "hex", "oct", "pow", "divmod", "format":
	default:
		return value{}, false, nil
	}
	if len(n.Args) == 0 && len(n.Keywords) == 0 && name != "zip" {
		return value{}, false, nil
	}
	if starIndex(n.Args) >= 0 || hasDoubleStar(n) {
		// f(*xs): the arguments are collected when the call runs.
		results := map[string]pyType{"zip": tObject, "map": tObject, "filter": tObject,
			"sorted": tObject, "list": tObject, "tuple": tObject, "set": tObject,
			"max": tObject, "min": tObject, "sum": tObject, "any": tObject, "all": tObject, "len": tObject,
			"enumerate": tObject, "dict": tObject}
		rt, known := results[name]
		if !known || g.isVariable(name) || g.sigs[name] != nil {
			return value{}, false, nil
		}
		args, kw, err := g.dynCallArgs(n)
		if err != nil {
			return value{}, true, err
		}
		if kw != "" {
			return value{}, false, nil
		}
		return g.dynValue(g.rt("DynBuiltinCall")+"("+strconv.Quote(name)+", "+args+")", rt), true, nil
	}
	args := make([]value, len(n.Args))
	any := false
	for i, a := range n.Args {
		v, err := g.typedExpr(a)
		if err != nil {
			return value{}, false, nil
		}
		args[i] = v
		if nm, isName := a.(*pyast.Name); isName && (dynBuiltinValues[nm.ID] || dynBuiltinTypes[nm.ID]) && !g.isVariable(nm.ID) {
			continue
		}
		if v.typ == tObject || g.hasObjectElem(v.typ) {
			any = true
		}
	}
	if !any && !force {
		return value{}, false, nil
	}
	kw := map[string]pyast.Node{}
	for _, k := range n.Keywords {
		if k.Arg == "" {
			return value{}, false, nil
		}
		kw[pyName(k.Arg)] = k.Value
	}
	rt := g.rt
	one := func(fn string, t pyType) (value, bool, error) {
		if len(args) != 1 || len(kw) != 0 {
			return value{}, false, nil
		}
		return g.dynValue(rt(fn)+"("+g.boxed(args[0])+")", t), true, nil
	}
	switch name {
	case "len":
		return one("DynLen", tInt)
	case "int":
		if len(args) == 2 && len(kw) == 0 {
			return g.dynValue(rt("DynIntBase")+"("+g.boxed(args[0])+", "+g.boxed(args[1])+")", tInt), true, nil
		}
		return one("DynInt", tInt)
	case "float":
		return one("DynFloat", tFloat)
	case "list":
		return one("DynList", listOf(tObject))
	case "tuple":
		return one("DynTuple", varTupleOf(tObject))
	case "set":
		return one("DynSet", setOf(tObject))
	case "dict":
		return one("DynDict", dictOf(tObject, tObject))
	case "reversed":
		return one("DynReversed", iterOf(tObject))
	case "any":
		return one("DynAny", tBool)
	case "all":
		return one("DynAll", tBool)
	case "hash":
		return one("Hash", tInt)
	case "abs":
		if len(args) == 1 {
			return g.dynUnary("Abs", args[0]), true, nil
		}
	case "round":
		digits := "nil"
		if len(args) == 2 {
			digits = g.boxed(args[1])
		}
		return g.dynValue(rt("DynRound")+"("+g.boxed(args[0])+", "+digits+")", tObject), true, nil
	case "sum":
		start := "int64(0)"
		if len(args) == 2 {
			start = g.boxed(args[1])
		}
		return g.dynValue(rt("DynSum")+"("+g.boxed(args[0])+", "+start+")", tObject), true, nil
	case "ord":
		return one("DynOrd", tInt)
	case "chr":
		return one("DynChr", tStr)
	case "bin":
		return one("DynBin", tStr)
	case "hex":
		return one("DynHex", tStr)
	case "oct":
		return one("DynOct", tStr)
	case "divmod":
		if len(args) != 2 || len(kw) != 0 {
			return value{}, false, nil
		}
		return g.dynValue(rt("DynDivmod")+"("+g.dynArgs(args)+")", varTupleOf(tObject)), true, nil
	case "pow":
		if len(args) < 2 || len(args) > 3 || len(kw) != 0 {
			return value{}, false, nil
		}
		mod := "nil"
		if len(args) == 3 {
			mod = g.boxed(args[2])
		}
		return g.dynValue(rt("DynPow")+"("+g.boxed(args[0])+", "+g.boxed(args[1])+", "+mod+")", tObject), true, nil
	case "format":
		if len(args) < 1 || len(args) > 2 || len(kw) != 0 {
			return value{}, false, nil
		}
		spec := `""`
		if len(args) == 2 {
			s, err := g.coerce(args[1], tStr, "format() spec", n.Line())
			if err != nil {
				return value{}, true, err
			}
			spec = s.code
		}
		return g.dynValue(rt("DynFormat")+"("+g.boxed(args[0])+", "+spec+")", tStr), true, nil
	case "map":
		if len(args) < 2 || len(kw) != 0 {
			return value{}, false, nil
		}
		return g.dynValue(rt("DynMap")+"("+g.dynArgs(args)+")", iterOf(tObject)), true, nil
	case "filter":
		if len(args) != 2 || len(kw) != 0 {
			return value{}, false, nil
		}
		return g.dynValue(rt("DynFilter")+"("+g.dynArgs(args)+")", iterOf(tObject)), true, nil
	case "zip":
		if len(kw) != 0 {
			return value{}, false, nil
		}
		return g.dynValue(rt("DynZip")+"("+g.dynArgs(args)+")", iterOf(varTupleOf(tObject))), true, nil
	case "sorted":
		if len(args) != 1 {
			return value{}, false, nil
		}
		key, reverse := "nil", "false"
		for k, node := range kw {
			v, err := g.typedExpr(node)
			if err != nil {
				return value{}, true, err
			}
			switch k {
			case "key":
				key = g.boxed(v)
			case "reverse":
				reverse = g.runtimeTruth(v)
			default:
				return value{}, false, nil
			}
		}
		return g.dynValue(rt("DynSorted")+"("+g.boxed(args[0])+", "+key+", "+reverse+")", listOf(tObject)), true, nil
	case "min", "max":
		items := ""
		if len(args) == 1 {
			items = g.boxed(args[0])
		} else {
			items = rt("NewVarTuple") + "[any](" + g.dynArgs(args) + ")"
		}
		key, def, hasDef := "nil", "nil", "false"
		for k, node := range kw {
			v, err := g.typedExpr(node)
			if err != nil {
				return value{}, true, err
			}
			switch k {
			case "key":
				key = g.boxed(v)
			case "default":
				def, hasDef = g.boxed(v), "true"
			default:
				return value{}, false, nil
			}
		}
		return g.dynValue(rt("DynMinMax")+"("+strconv.FormatBool(name == "max")+", "+items+", "+key+", "+def+", "+hasDef+")", tObject), true, nil
	case "enumerate":
		if len(kw) != 0 {
			return value{}, false, nil
		}
		start := "0"
		if len(args) == 2 {
			s, err := g.coerce(args[1], tInt, "enumerate() start", n.Line())
			if err != nil {
				return value{}, true, err
			}
			start = s.code
		}
		return g.dynValue(rt("DynEnumerate")+"("+g.boxed(args[0])+", "+start+")", iterOf(tupleOf([]pyType{tInt, tObject}))), true, nil
	}
	return value{}, false, nil
}

// dynIndexCode is the object a[...] indexes with: the key, or a Slice.
func (g *generator) dynIndexCode(index pyast.Node) (string, error) {
	sl, ok := index.(*pyast.Slice)
	if !ok {
		v, err := g.typedExpr(index)
		if err != nil {
			return "", err
		}
		return g.boxed(v), nil
	}
	parts := make([]string, 0, 3)
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
		parts = append(parts, part.name+": "+g.boxed(v))
	}
	return g.rt("Slice") + "{" + strings.Join(parts, ", ") + "}", nil
}

// dynSubscript is base[index] where the base or the index is an object.
func (g *generator) dynSubscript(base value, index pyast.Node) (value, error) {
	idx, err := g.dynIndexCode(index)
	if err != nil {
		return value{}, err
	}
	return g.dynValue(g.rt("DynGetItem")+"("+g.boxed(base)+", "+idx+")", tObject), nil
}

// indexIsObject reports whether the index expression is an object.
func (g *generator) indexIsObject(index pyast.Node) bool {
	if _, isSlice := index.(*pyast.Slice); isSlice {
		return false
	}
	v, err := g.typedExpr(index)
	return err == nil && v.typ == tObject
}

// dynCallArgs generates the arguments of a call as the runtime takes them: a
// []any and a []KW.
func (g *generator) dynCallArgs(n *pyast.Call) (args, kw string, err error) {
	simple := true
	for _, a := range n.Args {
		if _, star := a.(*pyast.Starred); star {
			simple = false
		}
	}
	for _, k := range n.Keywords {
		if k.Arg == "" {
			simple = false
		}
	}
	if simple {
		var pos, kws []string
		for _, a := range n.Args {
			v, err := g.typedExpr(a)
			if err != nil {
				return "", "", err
			}
			pos = append(pos, g.boxed(v))
		}
		for _, k := range n.Keywords {
			v, err := g.typedExpr(k.Value)
			if err != nil {
				return "", "", err
			}
			kws = append(kws, g.rt("KW")+"{Name: "+strconv.Quote(pyName(k.Arg))+", Val: "+g.boxed(v)+"}")
		}
		args = "[]any{" + strings.Join(pos, ", ") + "}"
		kw = "nil"
		if len(kws) > 0 {
			kw = "[]" + g.rt("KW") + "{" + strings.Join(kws, ", ") + "}"
		}
		return args, kw, nil
	}

	// With "*" and "**" the arguments are collected in order.
	var pre strings.Builder
	a, k := g.tmp("args"), g.tmp("kw")
	fmt.Fprintf(&pre, "%s := []any{}\nvar %s []%s\n", a, k, g.rt("KW"))
	for _, arg := range n.Args {
		if st, star := arg.(*pyast.Starred); star {
			v, err := g.typedExpr(st.Value)
			if err != nil {
				return "", "", err
			}
			fmt.Fprintf(&pre, "%s = %s(%s, %s)\n", a, g.rt("AppendStar"), a, g.boxed(v))
			continue
		}
		v, err := g.typedExpr(arg)
		if err != nil {
			return "", "", err
		}
		fmt.Fprintf(&pre, "%s = append(%s, %s)\n", a, a, g.boxed(v))
	}
	for _, kwd := range n.Keywords {
		v, err := g.typedExpr(kwd.Value)
		if err != nil {
			return "", "", err
		}
		if kwd.Arg == "" {
			fmt.Fprintf(&pre, "%s = %s(%s, %s)\n", k, g.rt("AppendDoubleStar"), k, g.boxed(v))
			continue
		}
		fmt.Fprintf(&pre, "%s = append(%s, %s{Name: %s, Val: %s})\n", k, k, g.rt("KW"), strconv.Quote(pyName(kwd.Arg)), g.boxed(v))
	}
	// The statements run in a function that returns both slices.
	fn := fmt.Sprintf("func() %s {\n%sreturn %s{Args: %s, KW: %s}\n}", g.rt("CallArgs"), pre.String(), g.rt("CallArgs"), a, k)
	return fn + "()", "", nil
}

// dynCall is f(args) for an object f.
func (g *generator) dynCall(n *pyast.Call, f value) (value, error) {
	g.callsUnknown = true
	args, kw, err := g.dynCallArgs(n)
	if err != nil {
		return value{}, err
	}
	if kw == "" {
		return g.dynValue(g.rt("DynCallPair")+"("+g.boxed(f)+", "+args+")", tObject), nil
	}
	return g.dynValue(g.rt("DynCallKw")+"("+g.boxed(f)+", "+args+", "+kw+")", tObject), nil
}

// dynMethodCall is recv.name(args) for an object recv.
func (g *generator) dynMethodCall(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	g.callsUnknown = true
	args, kw, err := g.dynCallArgs(n)
	if err != nil {
		return value{}, err
	}
	if kw == "" {
		return g.dynValue(g.rt("DynMethodPair")+"("+g.boxed(recv)+", "+strconv.Quote(m.Attr)+", "+args+")", tObject), nil
	}
	return g.dynValue(g.rt("DynCallMethod")+"("+g.boxed(recv)+", "+strconv.Quote(m.Attr)+", "+args+", "+kw+")", tObject), nil
}

// augDynItem is "base[index] op= value" where the base or the index is an
// object.
func (g *generator) augDynItem(b *strings.Builder, n *pyast.AugAssign, target *pyast.Subscript, base value) error {
	idx, err := g.dynIndexCode(target.Index)
	if err != nil {
		return err
	}
	right, err := g.typedExpr(n.Value)
	if err != nil {
		return err
	}
	obj, key := g.tmp("o"), g.tmp("k")
	fmt.Fprintf(b, "%s := %s\n%s := %s\n", obj, g.boxed(base), key, idx)
	fmt.Fprintf(b, "%s(%s, %s, %s(%q, %s(%s, %s), %s))\n", g.rt("DynSetItem"), obj, key,
		g.rt("DynInplace"), n.Op, g.rt("DynGetItem"), obj, key, g.boxed(right))
	return nil
}

// dynLambda is a lambda that nothing gives a Callable type: its parameters are
// objects.
func (g *generator) dynLambda(n *pyast.Lambda) (value, error) {
	a := n.Args
	if a.Vararg != nil || a.Kwarg != nil || len(a.Kwonly) > 0 || len(a.Defaults) > 0 {
		return g.lambdaAsDef(n)
	}
	params := make([]pyType, len(n.Args.Positional()))
	for i := range params {
		params[i] = tObject
	}
	v, err := g.lambda(n, params, tObject)
	if err != nil {
		if v, err = g.lambda(n, params, ""); err != nil {
			return v, err
		}
	}
	// The lambda keeps its parameters for calls through an object.
	names := make([]string, len(params))
	for i, p := range n.Args.Positional() {
		names[i] = strconv.Quote(p.Name)
	}
	v.code = g.rt("WithInfo") + "(" + v.code + ", &" + g.rt("FuncInfo") + `{Name: "<lambda>", Params: []string{` + strings.Join(names, ", ") + "}})"
	return v, nil
}

// hasObjectElem reports whether t is a container of objects.
func (g *generator) hasObjectElem(t pyType) bool {
	switch {
	case t.isList(), t.isSet(), t.isVarTuple(), t.isIter(), t.isGen(), t.isView():
		e, ok := t.iterElem()
		return ok && e == tObject
	case t.isDict():
		return t.keyType() == tObject || t.valType() == tObject
	}
	return false
}

// runtimeTruth is the Go bool of a value's truth.
func (g *generator) runtimeTruth(v value) string { return g.truth(v).code }

// dynMethodOK reports whether a method call on recv can go to the run time.
func (g *generator) dynMethodOK(recv value) bool {
	t := recv.typ
	return t == tObject || t == tStr || t.isList() || t.isDict() || t.isSet() || t.isVarTuple() || t == tInt || t == tFloat
}

// The built-ins that can be used as values.
var dynBuiltinValues = map[string]bool{
	"len": true, "abs": true, "repr": true, "any": true, "all": true, "hash": true, "reversed": true,
	"print": true, "sorted": true, "sum": true, "min": true, "max": true, "round": true, "enumerate": true,
	"pow": true, "divmod": true, "ord": true, "chr": true, "bin": true, "hex": true, "oct": true, "format": true,
	"isinstance": true, "callable": true, "iter": true, "next": true, "getattr": true, "hasattr": true,
	"zip": true, "map": true, "filter": true,
}

var dynBuiltinTypes = map[string]bool{
	"int": true, "float": true, "str": true, "bool": true, "list": true, "tuple": true, "set": true, "dict": true,
}

// boxable reports whether v can be handed to the dynamic operations.
func (g *generator) boxable(v value) bool {
	return v.typ != tNone && v.typ != "" && !v.typ.isClassRef()
}

// valueBoolOp is "a and b" and "a or b" where an operand is not a bool: the
// result is one of the operands, so it takes their common type.
func (g *generator) valueBoolOp(op string, values []pyast.Node) (value, error) {
	vals := make([]value, len(values))
	var t pyType
	for i, e := range values {
		v, err := g.typedExpr(e)
		if err != nil {
			return value{}, err
		}
		if !g.boxable(v) {
			return value{}, unsupported(fmt.Sprintf("'%s' on %s operand", strings.ToLower(op), v.typ), e.Line())
		}
		v = v.typed()
		vals[i] = v
		switch {
		case t == "":
			t = v.typ
		case t != v.typ:
			j, ok := g.joinTypes(t, v.typ)
			if !ok {
				j = tObject
			}
			t = j
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "func() %s {\n", g.goType(t))
	for i, v := range vals[:len(vals)-1] {
		tmp := g.tmp("b")
		fmt.Fprintf(&b, "%s := %s\n", tmp, g.relabel(v, t).code)
		cond := g.truth(value{code: tmp, typ: t, prec: precAtom}).code
		if op == "And" {
			cond = "!(" + cond + ")"
		}
		fmt.Fprintf(&b, "if %s {\nreturn %s\n}\n", cond, tmp)
		_ = i
	}
	fmt.Fprintf(&b, "return %s\n}()", g.relabel(vals[len(vals)-1], t).code)
	return value{code: b.String(), typ: t, prec: precAtom}, nil
}

// fallbackOK reports whether an error from the static rules may be left to the
// run time. A value of a known type that breaks the declared type of a
// parameter is a mistake the compiler keeps reporting; an object, or something
// the static rules do not cover, is decided when the program runs.
func fallbackOK(err error) bool {
	msg := err.Error()
	return !strings.Contains(msg, " expects ") || strings.Contains(msg, "object")
}

// hasDoubleStar reports whether a call passes **mapping.
func hasDoubleStar(n *pyast.Call) bool {
	for _, k := range n.Keywords {
		if k.Arg == "" {
			return true
		}
	}
	return false
}

// lambdaAsDef compiles a lambda that has defaults, keyword-only parameters or
// *args as the def it stands for, defined where the lambda is written:
//
//	lambda x, n=2: x * n   is   def _lambda(x, n=2): return x * n
func (g *generator) lambdaAsDef(n *pyast.Lambda) (value, error) {
	g.nextID++
	name := fmt.Sprintf("lambda_%d", g.nextID)
	fn := &pyast.FunctionDef{
		Name: name, Args: n.Args, Lineno: n.Line(),
		Body: []pyast.Node{&pyast.Return{Value: n.Body, Lineno: n.Line()}},
	}
	outer := g.env
	g.env = newScope(outer)
	defer func() { g.env = outer }()
	var b strings.Builder
	if err := g.predeclareDefs(&b, []pyast.Node{fn}); err != nil {
		return value{}, err
	}
	if err := g.nestedDef(&b, fn); err != nil {
		return value{}, err
	}
	sig, _ := g.env.lookupSig(name)
	f := value{code: name, typ: sig.goFuncType(), prec: precAtom, fsig: sig}
	if sig.plain() {
		f.typ = funcOf(sig.positionalTypes(), sig.ret)
	}
	sig.display = "<lambda>"
	sigName := sig.name
	sig.name = "<lambda>"
	f = g.relabel(f, tObject)
	sig.name = sigName
	return g.dynValue("func() any {\n"+b.String()+"return "+f.code+"\n}()", tObject), nil
}

// classesObject is the second argument of issubclass as an object: a class or
// a tuple of classes.
func (g *generator) classesObject(n pyast.Node) (string, error) {
	if tup, ok := n.(*pyast.Tuple); ok {
		parts := make([]string, len(tup.Elts))
		for i, e := range tup.Elts {
			p, err := g.classesObject(e)
			if err != nil {
				return "", err
			}
			parts[i] = p
		}
		return g.rt("NewVarTuple") + "[any](" + strings.Join(parts, ", ") + ")", nil
	}
	if name, ok := n.(*pyast.Name); ok && !g.isVariable(name.ID) {
		if c := g.classes[name.ID]; c != nil {
			return g.boxed(g.classObject(value{typ: classRefOf(c.name), prec: precAtom, classRef: c})), nil
		}
		if runtime.IsExceptionClass(name.ID) || dynBuiltinTypes[name.ID] || name.ID == "object" {
			return g.rt("BuiltinClassOf") + "(" + strconv.Quote(name.ID) + ")", nil
		}
	}
	v, err := g.typedExpr(n)
	if err != nil {
		return "", err
	}
	return g.boxed(v), nil
}

// userCallOrDynamic is a call of a def by name. The few call shapes the static
// binding of arguments does not cover (keywords together with **mapping) are
// made when the program runs.
func (g *generator) userCallOrDynamic(n *pyast.Call, sig *funcSig) (value, error) {
	v, err := g.userCall(n, sig)
	if err == nil || !callShapeError(err) {
		return v, err
	}
	fv := value{code: sig.goName, typ: sig.goFuncType(), prec: precAtom, fsig: sig}
	if sig.plain() {
		fv.typ = funcOf(sig.positionalTypes(), sig.ret)
	}
	return g.dynCall(n, g.relabel(fv, tObject))
}

// callShapeError reports whether an error says that the arguments do not fit
// the parameters. Python raises that TypeError when the call runs.
func callShapeError(err error) bool {
	msg := err.Error()
	for _, part := range []string{"together with a **mapping", "() missing ", "positional argument", "unexpected keyword argument",
		"multiple values for argument", "keyword-only argument"} {
		if strings.Contains(msg, part) {
			return true
		}
	}
	return false
}

// pythonBuiltins are the names of Python's builtins module. One that is not
// implemented stays a compile error: reading it would be a NameError in this
// program and not in Python.
var pythonBuiltins = map[string]bool{}

func init() {
	for _, n := range strings.Fields(`ArithmeticError AssertionError AttributeError BaseException BaseExceptionGroup BlockingIOError BrokenPipeError BufferError BytesWarning ChildProcessError ConnectionAbortedError ConnectionError ConnectionRefusedError ConnectionResetError DeprecationWarning EOFError Ellipsis EncodingWarning EnvironmentError Exception ExceptionGroup False FileExistsError FileNotFoundError FloatingPointError FutureWarning GeneratorExit IOError ImportError ImportWarning IndentationError IndexError InterruptedError IsADirectoryError KeyError KeyboardInterrupt LookupError MemoryError ModuleNotFoundError NameError None NotADirectoryError NotImplemented NotImplementedError OSError OverflowError PendingDeprecationWarning PermissionError ProcessLookupError PythonFinalizationError RecursionError ReferenceError ResourceWarning RuntimeError RuntimeWarning StopAsyncIteration StopIteration SyntaxError SyntaxWarning SystemError SystemExit TabError TimeoutError True TypeError UnboundLocalError UnicodeDecodeError UnicodeEncodeError UnicodeError UnicodeTranslateError UnicodeWarning UserWarning ValueError Warning ZeroDivisionError __debug__ abs aiter all anext any ascii bin bool breakpoint bytearray bytes callable chr classmethod compile complex copyright credits delattr dict dir divmod enumerate eval exec exit filter float format frozenset getattr globals hasattr hash help hex id input int isinstance issubclass iter len license list locals map max memoryview min next object oct open ord pow print property quit range repr reversed round set setattr slice sorted staticmethod str sum super tuple type vars zip`) {
		pythonBuiltins[n] = true
	}
}

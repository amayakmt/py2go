package codegen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// userCall generates a call of a def with known parameters: top-level
// functions and nested ones.
func (g *generator) userCall(n *pyast.Call, sig *funcSig) (value, error) {
	// At module level, Python runs top to bottom, so the def must come first.
	if !sig.nested && g.fnName == "" && !g.defined[sig.name] {
		return value{}, unsupported(fmt.Sprintf("call to '%s' before its definition", sig.name), n.Line())
	}
	if sig.nested {
		g.callsUnknown = true
	} else if g.topFn != "" {
		if g.calls[g.topFn] == nil {
			g.calls[g.topFn] = map[string]bool{}
		}
		g.calls[g.topFn][sig.name] = true
	}
	fname := sig.name
	if sig.display != "" {
		fname = sig.display
	}
	return g.bindCall(n, sig, sig.goName, fname)
}

// callValue generates a call of an expression whose type is a Callable.
func (g *generator) callValue(n *pyast.Call, fn value) (value, error) {
	g.callsUnknown = true
	if hasSpread(n) {
		return value{}, unsupported("*iterable or **mapping argument in a call of a Callable value (the TypeError text needs the function's own name)", n.Line())
	}
	if n.HasKeywords && !onlyDoubleStar(n) {
		return value{}, unsupported("keyword arguments in a call of a Callable value", n.Line())
	}
	sig := &funcSig{name: "function", ret: fn.typ.funcRet()}
	for i, t := range fn.typ.funcParams() {
		sig.pos = append(sig.pos, param{name: "arg" + strconv.Itoa(i+1), typ: t})
	}
	return g.bindCall(n, sig, fn.at(precAtom), "function")
}

func hasSpread(n *pyast.Call) bool {
	for _, a := range n.Args {
		if _, ok := a.(*pyast.Starred); ok {
			return true
		}
	}
	for _, kw := range n.Keywords {
		if kw.Arg == "" {
			return true
		}
	}
	return false
}

func onlyDoubleStar(n *pyast.Call) bool {
	for _, kw := range n.Keywords {
		if kw.Arg != "" {
			return false
		}
	}
	return true
}

// slot is an argument matched to a parameter.
type slot struct {
	node  pyast.Node
	order int // position in the call as written, for evaluation order
	label string
}

// bindCall matches the arguments of a call to the parameters of sig, as
// Python does, and generates the Go call. Everything that can be checked
// statically is reported with CPython's TypeError message.
func (g *generator) bindCall(n *pyast.Call, sig *funcSig, callee, fname string) (value, error) {
	for _, a := range n.Args {
		if _, ok := a.(*pyast.Starred); ok {
			return g.bindDynamic(n, sig, callee, fname)
		}
	}
	for _, kw := range n.Keywords {
		if kw.Arg == "" {
			return g.bindDynamic(n, sig, callee, fname)
		}
	}

	nPos := len(sig.pos)
	given := len(n.Args)
	kwonlyGiven := int64(0)
	for _, kw := range n.Keywords {
		for _, p := range sig.kwonly {
			if p.name == kw.Arg {
				kwonlyGiven++
			}
		}
	}
	if given > nPos && sig.vararg == nil {
		return value{}, unsupported(runtime.MsgTooManyPositional(fname, int64(given+sig.implicit), int64(sig.minPositional()+sig.implicit), int64(nPos+sig.implicit), kwonlyGiven), n.Line())
	}

	slots := make([]*slot, nPos+len(sig.kwonly))
	for i := 0; i < given && i < nPos; i++ {
		slots[i] = &slot{n.Args[i], i, fmt.Sprintf("argument %d of '%s'", i+1, fname)}
	}

	type extra struct {
		node  pyast.Node
		order int
	}
	var extras []extra
	for i := nPos; i < given; i++ {
		extras = append(extras, extra{n.Args[i], i})
	}

	var kwNames []string
	for _, p := range sig.pos[sig.nPosOnly:] {
		kwNames = append(kwNames, p.name)
	}
	for _, p := range sig.kwonly {
		kwNames = append(kwNames, p.name)
	}
	type kwEntry struct {
		name  string
		node  pyast.Node
		order int
	}
	var kwEntries []kwEntry
	var posOnlyAsKw []string
	for ki, kw := range n.Keywords {
		idx := -1
		for i := sig.nPosOnly; i < nPos; i++ {
			if sig.pos[i].name == kw.Arg {
				idx = i
			}
		}
		for j, p := range sig.kwonly {
			if p.name == kw.Arg {
				idx = nPos + j
			}
		}
		switch {
		case idx >= 0:
			if slots[idx] != nil {
				return value{}, unsupported(runtime.MsgMultipleValues(fname, kw.Arg), kw.Line())
			}
			slots[idx] = &slot{kw.Value, given + ki, fmt.Sprintf("argument '%s' of '%s'", kw.Arg, fname)}
		case sig.kwarg != nil:
			kwEntries = append(kwEntries, kwEntry{kw.Arg, kw.Value, given + ki})
		default:
			isPosOnly := false
			for _, p := range sig.pos[:sig.nPosOnly] {
				isPosOnly = isPosOnly || p.name == kw.Arg
			}
			if isPosOnly {
				posOnlyAsKw = append(posOnlyAsKw, kw.Arg)
				continue
			}
			return value{}, unsupported(runtime.MsgUnexpectedKeyword(fname, kw.Arg, runtime.Suggest(kw.Arg, kwNames)), kw.Line())
		}
	}
	if len(posOnlyAsKw) > 0 {
		return value{}, unsupported(runtime.MsgPositionalOnlyAsKeyword(fname, posOnlyAsKw...), n.Line())
	}

	var missing, missingKw []string
	for i, p := range sig.pos {
		if slots[i] == nil && !p.hasDefault {
			missing = append(missing, p.name)
		}
	}
	if len(missing) > 0 {
		return value{}, unsupported(runtime.MsgMissing(fname, "positional", missing...), n.Line())
	}
	for j, p := range sig.kwonly {
		if slots[nPos+j] == nil && !p.hasDefault {
			missingKw = append(missingKw, p.name)
		}
	}
	if len(missingKw) > 0 {
		return value{}, unsupported(runtime.MsgMissing(fname, "keyword-only", missingKw...), n.Line())
	}

	// Generate the supplied arguments first, each once, remembering the order
	// they were written in.
	type supplied struct {
		code  string
		goT   string
		order int
	}
	var vals []supplied
	add := func(node pyast.Node, t pyType, label string, order int) (int, error) {
		v, err := g.exprAs(node, t)
		if err != nil {
			return 0, err
		}
		v, err = g.coerce(v, t, label, node.Line())
		if err != nil {
			return 0, err
		}
		vals = append(vals, supplied{v.code, g.goType(t), order})
		return len(vals) - 1, nil
	}

	params := append(append([]param{}, sig.pos...), sig.kwonly...)
	slotVal := make([]int, len(params))
	for i, p := range params {
		slotVal[i] = -1
		if slots[i] != nil {
			k, err := add(slots[i].node, p.typ, slots[i].label, slots[i].order)
			if err != nil {
				return value{}, err
			}
			slotVal[i] = k
		}
	}
	var varVals, kwVals []int
	if sig.vararg != nil {
		elem := sig.vararg.typ.elem()
		for _, e := range extras {
			k, err := add(e.node, elem, fmt.Sprintf("argument %d of '%s'", e.order+1, fname), e.order)
			if err != nil {
				return value{}, err
			}
			varVals = append(varVals, k)
		}
	}
	if sig.kwarg != nil {
		val := sig.kwarg.typ.valType()
		for _, e := range kwEntries {
			k, err := add(e.node, val, fmt.Sprintf("argument '%s' of '%s'", e.name, fname), e.order)
			if err != nil {
				return value{}, err
			}
			kwVals = append(kwVals, k)
		}
	}

	// Python evaluates the arguments in the order written. Go evaluates the
	// ones in the call in the order of the parameters, which differs when
	// keywords are given out of order, so such a call evaluates into
	// temporaries first.
	useTemps := len(vals) > 1 && !isMonotone(vals, func(s supplied) int { return s.order })
	ref := make([]string, len(vals))
	for k, v := range vals {
		ref[k] = v.code
	}
	var prelude strings.Builder
	if useTemps {
		byOrder := make([]int, len(vals))
		for k := range byOrder {
			byOrder[k] = k
		}
		sort.SliceStable(byOrder, func(a, b int) bool { return vals[byOrder[a]].order < vals[byOrder[b]].order })
		for _, k := range byOrder {
			tmp := g.tmp("a")
			fmt.Fprintf(&prelude, "var %s %s = %s\n", tmp, vals[k].goT, vals[k].code)
			ref[k] = tmp
		}
	}

	codes := make([]string, 0, len(params)+3)
	if sig.leadArg != "" {
		codes = append(codes, sig.leadArg)
	}
	for i, p := range params {
		if slotVal[i] < 0 {
			d, err := g.defaultArg(p)
			if err != nil {
				return value{}, err
			}
			codes = append(codes, d)
		} else {
			codes = append(codes, ref[slotVal[i]])
		}
	}
	if sig.vararg != nil {
		items := make([]string, len(varVals))
		for i, k := range varVals {
			items[i] = ref[k]
		}
		codes = append(codes, g.rt("NewVarTuple")+"["+g.goType(sig.vararg.typ.elem())+"]("+strings.Join(items, ", ")+")")
	}
	if sig.kwarg != nil {
		val := sig.kwarg.typ.valType()
		pairs := make([]string, len(kwVals))
		for i, k := range kwVals {
			pairs[i] = fmt.Sprintf("%s[string, %s]{Key: %s, Val: %s}", g.rt("KV"), g.goType(val), strconv.Quote(kwEntries[i].name), ref[k])
		}
		codes = append(codes, g.rt("NewDict")+"[string, "+g.goType(val)+"]("+strings.Join(pairs, ", ")+")")
	}

	call := callee + "(" + strings.Join(codes, ", ") + ")"
	if !useTemps {
		return value{code: call, typ: sig.ret, prec: precAtom}, nil
	}
	if sig.ret == tNone {
		return value{code: "func() {\n" + prelude.String() + call + "\n}()", typ: sig.ret, prec: precAtom}, nil
	}
	return value{code: "func() " + g.goType(sig.ret) + " {\n" + prelude.String() + "return " + call + "\n}()", typ: sig.ret, prec: precAtom}, nil
}

func isMonotone[T any](xs []T, key func(T) int) bool {
	for i := 1; i < len(xs); i++ {
		if key(xs[i]) < key(xs[i-1]) {
			return false
		}
	}
	return true
}

// bindDynamic generates a call with *iterable or **mapping arguments, whose
// length and keys are only known at run time. The call becomes a function
// literal that binds the parameters one by one, raising CPython's TypeErrors
// for a count or a key that does not fit. The unpacked items must share one
// type, which every parameter they can reach must have too.
func (g *generator) bindDynamic(n *pyast.Call, sig *funcSig, callee, fname string) (value, error) {
	// f(*a, 1, *b): everything from the first star on is one list.
	if first := starIndex(n.Args); first >= 0 {
		merge := false
		for _, a := range n.Args[first:] {
			if _, ok := a.(*pyast.Starred); !ok {
				merge = true
			}
		}
		if merge {
			cp := *n
			cp.Args = append(append([]pyast.Node{}, n.Args[:first]...),
				&pyast.Starred{Value: &pyast.List{Elts: n.Args[first:], Lineno: n.Line()}, Lineno: n.Line()})
			n = &cp
		}
	}
	var static []pyast.Node
	var stars []*pyast.Starred
	for _, a := range n.Args {
		if st, ok := a.(*pyast.Starred); ok {
			stars = append(stars, st)
			continue
		}
		if len(stars) > 0 {
			return value{}, unsupported("a positional argument after a *iterable argument", a.Line())
		}
		static = append(static, a)
	}
	var dstar pyast.Node
	for _, kw := range n.Keywords {
		if kw.Arg != "" {
			return value{}, unsupported("keyword arguments together with a **mapping argument", kw.Line())
		}
		if dstar != nil {
			return value{}, unsupported("more than one **mapping argument", kw.Line())
		}
		dstar = kw.Value
	}

	nPos := len(sig.pos)
	s := len(static)
	var b strings.Builder
	retGo := ""
	if sig.ret != tNone {
		retGo = " " + g.goType(sig.ret)
	}
	fmt.Fprintf(&b, "func()%s {\n", retGo)

	// The unpacked iterables become one list of items of one type.
	list, count := "", "int64("+strconv.Itoa(s)+")"
	var itemT pyType
	if len(stars) > 0 {
		parts := make([]string, len(stars))
		for i, st := range stars {
			v, err := g.typedExpr(st.Value)
			if err != nil {
				return value{}, err
			}
			var code string
			var elem pyType
			switch {
			case v.typ.isTuple():
				elems := v.typ.elems()
				if !allEqual(elems, elems[0]) {
					return value{}, unsupported(fmt.Sprintf("*%s unpacks items of different types into arguments", v.typ), st.Line())
				}
				elem = elems[0]
				code = call(g.rt("SeqList")+"["+g.goType(elem)+"]", v.code)
			default:
				c, e, ok := g.iterableOf(v)
				if !ok {
					return value{}, unsupported(fmt.Sprintf("*argument of type %s is not iterable", v.typ), st.Line())
				}
				elem = e
				if v.typ.isList() {
					code = v.code
				} else {
					code = call(g.rt("ListFrom")+"["+g.goType(elem)+"]", c)
				}
			}
			if i == 0 {
				itemT = elem
			} else if elem != itemT {
				return value{}, unsupported(fmt.Sprintf("*arguments of different item types (%s and %s)", itemT, elem), st.Line())
			}
			parts[i] = code
		}
		list = g.tmp("l")
		if len(parts) == 1 {
			fmt.Fprintf(&b, "var %s %s = %s\n", list, g.goType(listOf(itemT)), parts[0])
		} else {
			fmt.Fprintf(&b, "var %s %s = %s\n", list, g.goType(listOf(itemT)), call(g.rt("JoinLists")+"["+g.goType(itemT)+"]", parts...))
		}
		fmt.Fprintf(&b, "_ = %s\n", list)
		count += " + " + list + ".Len()"
	}
	n_ := g.tmp("n")
	fmt.Fprintf(&b, "%s := %s\n_ = %s\n", n_, count, n_)

	// The mapping of keyword arguments.
	dict := ""
	var valT pyType
	if dstar != nil {
		v, err := g.typedExpr(dstar)
		if err != nil {
			return value{}, err
		}
		if !v.typ.isDict() || v.typ.keyType() != tStr {
			return value{}, unsupported(fmt.Sprintf("**%s, which is not a dict with str keys", v.typ), dstar.Line())
		}
		valT = v.typ.valType()
		dict = g.tmp("d")
		fmt.Fprintf(&b, "var %s %s = %s\n_ = %s\n", dict, g.goType(v.typ), v.code, dict)
	}

	// Too many positional arguments.
	maxPos := strconv.Itoa(nPos)
	if sig.vararg == nil {
		fmt.Fprintf(&b, "if %s > %s {\n%s(%q, %s+%d, %d, %d, 0)\n}\n", n_, maxPos, g.rt("TooManyPositional"), fname, n_, sig.implicit, sig.minPositional()+sig.implicit, nPos+sig.implicit)
	}

	// Each parameter, filled from the static arguments, the list, the
	// mapping or its default.
	var missing, missingKw string
	var args []string
	need := func(p param, from string, t pyType, line int) error {
		if t != p.typ {
			return unsupported(fmt.Sprintf("%s unpacked into parameter '%s' of type %s", from, p.name, p.typ), line)
		}
		return nil
	}
	if len(stars) > 0 || dstar != nil {
		missing, missingKw = g.tmp("missing"), g.tmp("missing")
		fmt.Fprintf(&b, "var %s, %s []string\n_, _ = %s, %s\n", missing, missingKw, missing, missingKw)
	}
	for i, p := range sig.pos {
		a := g.tmp("a")
		args = append(args, a)
		fmt.Fprintf(&b, "var %s %s\n", a, g.goType(p.typ))
		switch {
		case i < s:
			v, err := g.exprAs(static[i], p.typ)
			if err != nil {
				return value{}, err
			}
			v, err = g.coerce(v, p.typ, fmt.Sprintf("argument %d of '%s'", i+1, fname), static[i].Line())
			if err != nil {
				return value{}, err
			}
			fmt.Fprintf(&b, "%s = %s\n", a, v.code)
			if dict != "" && i >= sig.nPosOnly {
				fmt.Fprintf(&b, "if %s.Has(%q) {\n%s(%q, %q)\n}\n", dict, p.name, g.rt("MultipleValues"), fname, p.name)
			}
		default:
			fmt.Fprintf(&b, "switch {\n")
			if list != "" {
				if err := need(p, "*"+string(listOf(itemT)), itemT, n.Line()); err != nil {
					return value{}, err
				}
				fmt.Fprintf(&b, "case %s < %s:\n%s = %s.Get(%d)\n", strconv.Itoa(i), n_, a, list, i-s)
				if dict != "" && i >= sig.nPosOnly {
					fmt.Fprintf(&b, "if %s.Has(%q) {\n%s(%q, %q)\n}\n", dict, p.name, g.rt("MultipleValues"), fname, p.name)
				}
			}
			if dict != "" && i >= sig.nPosOnly {
				if err := need(p, "**dict[str, "+string(valT)+"]", valT, n.Line()); err != nil {
					return value{}, err
				}
				fmt.Fprintf(&b, "case %s.Has(%q):\n%s = %s.Get(%q)\n", dict, p.name, a, dict, p.name)
			}
			if p.hasDefault {
				d, err := g.defaultArg(p)
				if err != nil {
					return value{}, err
				}
				fmt.Fprintf(&b, "default:\n%s = %s\n", a, d)
			} else {
				fmt.Fprintf(&b, "default:\n%s = append(%s, %q)\n", missing, missing, p.name)
			}
			b.WriteString("}\n")
		}
	}
	if len(stars) > 0 || dict != "" {
		fmt.Fprintf(&b, "if len(%s) > 0 {\n%s(%q, %s...)\n}\n", missing, g.rt("MissingPositional"), fname, missing)
	}
	for _, p := range sig.kwonly {
		a := g.tmp("a")
		args = append(args, a)
		fmt.Fprintf(&b, "var %s %s\n", a, g.goType(p.typ))
		if dict != "" {
			if err := need(p, "**dict[str, "+string(valT)+"]", valT, n.Line()); err != nil {
				return value{}, err
			}
			fmt.Fprintf(&b, "if %s.Has(%q) {\n%s = %s.Get(%q)\n} else ", dict, p.name, a, dict, p.name)
		}
		if p.hasDefault {
			d, err := g.defaultArg(p)
			if err != nil {
				return value{}, err
			}
			fmt.Fprintf(&b, "{\n%s = %s\n}\n", a, d)
		} else {
			fmt.Fprintf(&b, "{\n%s = append(%s, %q)\n}\n", missingKw, missingKw, p.name)
		}
	}
	if len(sig.kwonly) > 0 && missingKw != "" {
		fmt.Fprintf(&b, "if len(%s) > 0 {\n%s(%q, %s...)\n}\n", missingKw, g.rt("MissingKeywordOnly"), fname, missingKw)
	}

	// *args receives what is left over, **kwargs the keys no parameter took.
	if sig.vararg != nil {
		elem := sig.vararg.typ.elem()
		var extra []string
		for i := nPos; i < s; i++ {
			v, err := g.exprAs(static[i], elem)
			if err != nil {
				return value{}, err
			}
			v, err = g.coerce(v, elem, fmt.Sprintf("argument %d of '%s'", i+1, fname), static[i].Line())
			if err != nil {
				return value{}, err
			}
			extra = append(extra, v.code)
		}
		tail := "nil"
		if list != "" {
			if itemT != elem {
				return value{}, unsupported(fmt.Sprintf("*%s unpacked into *%s", listOf(itemT), sig.vararg.name), n.Line())
			}
			tail = list
		}
		args = append(args, call(g.rt("VarTupleOf")+"["+g.goType(elem)+"]",
			g.rt("NewList")+"["+g.goType(elem)+"]("+strings.Join(extra, ", ")+")", tail, strconv.Itoa(max(nPos-s, 0))))
	}
	known := []string{}
	posonly := []string{}
	for _, p := range sig.pos[:sig.nPosOnly] {
		posonly = append(posonly, p.name)
	}
	for _, p := range sig.pos[sig.nPosOnly:] {
		known = append(known, p.name)
	}
	for _, p := range sig.kwonly {
		known = append(known, p.name)
	}
	if sig.kwarg != nil {
		val := sig.kwarg.typ.valType()
		if dict != "" {
			if val != valT {
				return value{}, unsupported(fmt.Sprintf("**dict[str, %s] unpacked into **%s of dict[str, %s]", valT, sig.kwarg.name, val), n.Line())
			}
			args = append(args, call(g.rt("SplitKeys")+"["+g.goType(val)+"]", dict, strconv.Quote(fname), quoteList(known), quoteList(posonly), "true"))
		} else {
			args = append(args, g.rt("NewDict")+"[string, "+g.goType(val)+"]()")
		}
	} else if dict != "" {
		fmt.Fprintf(&b, "%s\n", call(g.rt("SplitKeys")+"["+g.goType(valT)+"]", dict, strconv.Quote(fname), quoteList(known), quoteList(posonly), "false"))
	}

	callCode := callee + "(" + strings.Join(args, ", ") + ")"
	if sig.ret == tNone {
		b.WriteString(callCode + "\n}()")
	} else {
		b.WriteString("return " + callCode + "\n}()")
	}
	return value{code: b.String(), typ: sig.ret, prec: precAtom}, nil
}

// quoteList spells a []string literal.
func quoteList(names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = strconv.Quote(n)
	}
	return "[]string{" + strings.Join(q, ", ") + "}"
}

// defaultArg is the value a call passes for a parameter it gives no argument
// for: the default evaluated when the def ran, or for a dataclass field with
// default_factory a new value made by the factory.
func (g *generator) defaultArg(p param) (string, error) {
	if p.factory != nil {
		return g.factoryValue(&dcField{name: p.name, typ: p.typ, factory: p.factory, line: p.line})
	}
	return p.defaultVar, nil
}

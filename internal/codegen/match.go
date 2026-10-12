package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// A match statement becomes a chain of cases. Each case is a function
// literal that tests the pattern and assigns the names it captures, which are
// declared before it, followed by the guard and the body:
//
//	{ var s T = subject; matched := false
//	  if !matched { var a int64; if func() bool {...; a = ...; return true}() && guard { matched = true; body } }
//	  ... }

// patternCtx is the state while one case is compiled.
type patternCtx struct {
	body     *strings.Builder // statements of the test closure
	captures []capture
	narrowed *value // the subject as a class pattern left it, see patternNarrow
}

type capture struct {
	name   string
	typ    pyType
	assign bool // the name already exists, so the case assigns it
}

func (g *generator) matchStmt(b *strings.Builder, n *pyast.Match) error {
	subj, err := g.typedExpr(n.Subject)
	if err != nil {
		return err
	}
	if subj.typ == tNoneType {
		return unsupported("match on None itself", n.Line())
	}
	b.WriteString("{\n")
	s := g.tmp("subject")
	g.temp(b, s, subj.typ, subj.typed().code)
	done := g.tmp("matched")
	fmt.Fprintf(b, "%s := false\n_ = %s\n", done, done)
	subject := value{code: s, typ: subj.typ, prec: precAtom}

	for _, c := range n.Cases {
		outer := g.env
		g.env = newScope(outer)
		pc := &patternCtx{body: &strings.Builder{}}
		if err := g.pattern(pc, c.Pattern, subject); err != nil {
			g.env = outer
			return err
		}
		var decls strings.Builder
		for _, cap := range pc.captures {
			if !cap.assign {
				g.env.declare(cap.name, cap.typ)
				fmt.Fprintf(&decls, "var %s %s\n_ = %s\n", cap.name, g.goType(cap.typ), cap.name)
			}
		}
		cond := "func() bool {\n" + pc.body.String() + "return true\n}()"
		if c.Guard != nil {
			guard, err := g.condition(c.Guard)
			if err != nil {
				g.env = outer
				return err
			}
			cond += " && " + guard.at(precAnd+1)
		}
		fmt.Fprintf(b, "if !%s {\n%sif %s {\n%s = true\n", done, decls.String(), cond, done)
		if err := g.stmtList(b, c.Body); err != nil {
			g.env = outer
			return err
		}
		b.WriteString("}\n}\n")
		g.env = outer
	}
	b.WriteString("}\n")
	return nil
}

// fail writes "this pattern does not match" when cond is false.
func (pc *patternCtx) require(cond string) {
	fmt.Fprintf(pc.body, "if !(%s) {\nreturn false\n}\n", cond)
}

// bindSub puts a part of the subject in a variable of the closure, which
// compare can then refer to by name.
func (g *generator) bindSub(pc *patternCtx, v value) value {
	if simpleOperand.MatchString(v.code) {
		return v
	}
	id := g.tmp("p")
	g.env.declare(id, v.typ)
	fmt.Fprintf(pc.body, "var %s %s = %s\n_ = %s\n", id, g.goType(v.typ), v.typed().code, id)
	return value{code: id, typ: v.typ, prec: precAtom}
}

// nameNode refers to a bound value from a synthetic expression.
func (g *generator) nameNode(v value, line int) pyast.Node {
	g.env.declare(v.code, v.typ)
	return &pyast.Name{ID: v.code, Lineno: line}
}

func (g *generator) captureName(pc *patternCtx, name string, v value, line int) error {
	if name == "" || name == "_" {
		return nil
	}
	if err := checkName(name, line); err != nil {
		return err
	}
	for _, c := range pc.captures {
		if c.name == name {
			return unsupported(fmt.Sprintf("name '%s' captured twice in a pattern", name), line)
		}
	}
	t, declared, global := g.lookupBinding(name)
	if declared {
		if t != v.typ {
			if conv, err := g.coerce(v, t, fmt.Sprintf("variable '%s'", name), line); err == nil {
				v = conv
			} else {
				return err
			}
		}
		fmt.Fprintf(pc.body, "%s = %s\n", name, v.typed().code)
		g.afterAssign(pc.body, name, global)
		pc.captures = append(pc.captures, capture{name, t, true})
		return nil
	}
	pc.captures = append(pc.captures, capture{name, v.typ, false})
	g.env.declare(name, v.typ)
	fmt.Fprintf(pc.body, "%s = %s\n", name, v.typed().code)
	return nil
}

// pattern compiles one pattern against a value into pc.
func (g *generator) pattern(pc *patternCtx, p pyast.Node, v value) error {
	_, err := g.patternNarrow(pc, p, v)
	return err
}

// patternNarrow is pattern, and also returns the subject as it is known to be
// once the pattern matched: an object tested against a class is that class.
func (g *generator) patternNarrow(pc *patternCtx, p pyast.Node, v value) (value, error) {
	switch p := p.(type) {
	case *pyast.MatchAs:
		if p.Pattern != nil {
			nv, err := g.patternNarrow(pc, p.Pattern, v)
			if err != nil {
				return v, err
			}
			v = nv
		}
		return v, g.captureName(pc, p.Name, v, p.Line())

	case *pyast.MatchValue:
		return v, g.valuePattern(pc, p, v)

	case *pyast.MatchSingleton:
		switch p.Value {
		case "None":
			switch {
			case v.typ.isOptional():
				pc.require("!" + v.at(precAtom) + ".IsSome()")
			case v.typ == tNoneType:
			case v.typ == tObject:
				pc.require(g.rt("IsNone") + "(" + v.code + ")")
			default:
				pc.body.WriteString("return false\n")
			}
		default:
			v = g.stripOptional(pc, v)
			if v.typ == tObject {
				pc.require(g.rt("Unbox") + "(" + v.code + ") == any(" + strings.ToLower(p.Value) + ")")
			} else if v.typ != tBool {
				pc.body.WriteString("return false\n")
			} else if p.Value == "True" {
				pc.require(v.code)
			} else {
				pc.require("!" + v.at(precUnary))
			}
		}
		return v, nil

	case *pyast.MatchOr:
		var alts []string
		for _, alt := range p.Patterns {
			sub := &patternCtx{body: &strings.Builder{}}
			if err := g.pattern(sub, alt, v); err != nil {
				return v, err
			}
			if len(sub.captures) > 0 {
				return v, unsupported("or-pattern that captures names", p.Line())
			}
			alts = append(alts, "func() bool {\n"+sub.body.String()+"return true\n}()")
		}
		pc.require(strings.Join(alts, " || "))
		return v, nil

	case *pyast.MatchSequence:
		return v, g.sequencePattern(pc, p, v)

	case *pyast.MatchMapping:
		return v, g.mappingPattern(pc, p, v)

	case *pyast.MatchClass:
		pc.narrowed = nil
		err := g.classPattern(pc, p, v)
		if pc.narrowed != nil {
			return *pc.narrowed, err
		}
		return v, err

	case *pyast.MatchStar:
		return v, unsupported("star pattern outside a sequence pattern", p.Line())
	}
	return v, unsupported("pattern of kind "+nodeName(p), p.Line())
}

// stripOptional reads the value of an Optional that must be a value to
// match: None never matches the patterns that follow.
func (g *generator) stripOptional(pc *patternCtx, v value) value {
	if !v.typ.isOptional() {
		return v
	}
	pc.require(v.at(precAtom) + ".IsSome()")
	return value{code: v.at(precAtom) + ".Get()", typ: v.typ.optElem(), prec: precAtom}
}

func (g *generator) valuePattern(pc *patternCtx, p *pyast.MatchValue, v value) error {
	lit, err := g.typedExpr(p.Value)
	if err != nil {
		return err
	}
	v = g.bindSub(pc, v)
	numeric := isNumeric(strip(v.typ)) && isNumeric(lit.typ)
	if strip(v.typ) != lit.typ && !numeric && !(v.typ == tObject) && !g.isSubtype(lit.typ, v.typ) && !g.isSubtype(v.typ, lit.typ) {
		// A value of another type is never equal.
		pc.body.WriteString("return false\n")
		return nil
	}
	left := g.nameNode(v, p.Line())
	cmp, err := g.compare(&pyast.Compare{Left: left, Ops: []string{"Eq"}, Comparators: []pyast.Node{p.Value}, Lineno: p.Line()})
	if err != nil {
		return err
	}
	pc.require(cmp.code)
	return nil
}

func (g *generator) sequencePattern(pc *patternCtx, p *pyast.MatchSequence, v value) error {
	v = g.stripOptional(pc, v)
	star := -1
	for i, e := range p.Patterns {
		if _, ok := e.(*pyast.MatchStar); ok {
			if star >= 0 {
				return unsupported("two starred names in a sequence pattern", p.Line())
			}
			star = i
		}
	}
	fixed := len(p.Patterns)
	if star >= 0 {
		fixed--
	}
	if v.typ == tObject {
		pc.require(g.rt("IsSequence") + "(" + v.code + ")")
		v = value{code: g.rt("DynList") + "(" + v.code + ")", typ: listOf(tObject), prec: precAtom}
	}
	v = g.bindSub(pc, v)
	switch {
	case v.typ.isTuple():
		elems := v.typ.elems()
		if (star < 0 && len(elems) != fixed) || (star >= 0 && len(elems) < fixed) {
			pc.body.WriteString("return false\n")
			return nil
		}
		for i, e := range p.Patterns {
			switch {
			case i == star:
				return unsupported("starred name in a pattern for a tuple", p.Line())
			default:
				sub := value{code: fmt.Sprintf("%s.F%d", v.at(precAtom), i), typ: elems[i], prec: precAtom}
				if err := g.pattern(pc, e, sub); err != nil {
					return err
				}
			}
		}
		return nil
	case v.typ.isList() || v.typ.isVarTuple():
		elem := v.typ.elem()
		if star < 0 {
			pc.require(fmt.Sprintf("%s.Len() == %d", v.at(precAtom), fixed))
		} else {
			pc.require(fmt.Sprintf("%s.Len() >= %d", v.at(precAtom), fixed))
		}
		after := 0
		if star >= 0 {
			after = len(p.Patterns) - star - 1
		}
		for i, e := range p.Patterns {
			switch {
			case i == star:
				name := e.(*pyast.MatchStar).Name
				if name == "" {
					continue
				}
				slice := fmt.Sprintf("%s.Slice(%s{Lo: %d, HasLo: true, Hi: %s.Len() - %d, HasHi: true})", v.at(precAtom), g.rt("Bounds"), star, v.at(precAtom), after)
				rest := value{code: slice, typ: v.typ, prec: precAtom}
				if v.typ.isVarTuple() {
					rest = value{code: slice + ".ToList()", typ: listOf(elem), prec: precAtom}
				}
				if err := g.captureName(pc, name, rest, e.Line()); err != nil {
					return err
				}
			case star >= 0 && i > star:
				sub := value{code: fmt.Sprintf("%s.Get(%s.Len() - %d)", v.at(precAtom), v.at(precAtom), len(p.Patterns)-i), typ: elem, prec: precAtom}
				if err := g.pattern(pc, e, sub); err != nil {
					return err
				}
			default:
				sub := value{code: fmt.Sprintf("%s.Get(%d)", v.at(precAtom), i), typ: elem, prec: precAtom}
				if err := g.pattern(pc, e, sub); err != nil {
					return err
				}
			}
		}
		return nil
	case v.typ == tObject, v.typ.isClass():
		return unsupported("sequence pattern on a value of type "+string(v.typ), p.Line())
	}
	// str, dict and numbers never match a sequence pattern.
	pc.body.WriteString("return false\n")
	return nil
}

func (g *generator) mappingPattern(pc *patternCtx, p *pyast.MatchMapping, v value) error {
	v = g.stripOptional(pc, v)
	if v.typ == tObject {
		pc.require(g.rt("IsMapping") + "(" + v.code + ")")
		v = value{code: g.rt("DynDict") + "(" + v.code + ")", typ: dictOf(tObject, tObject), prec: precAtom}
	}
	v = g.bindSub(pc, v)
	if !v.typ.isDict() {
		pc.body.WriteString("return false\n")
		return nil
	}
	var keyCodes []string
	for i, kn := range p.Keys {
		k, err := g.exprAs(kn, v.typ.keyType())
		if err != nil {
			return err
		}
		if v.typ.keyType() == tObject {
			k = g.relabel(k, tObject)
		}
		if k.typ != v.typ.keyType() {
			pc.body.WriteString("return false\n")
			return nil
		}
		pc.require(fmt.Sprintf("%s.Has(%s)", v.at(precAtom), k.typed().code))
		keyCodes = append(keyCodes, k.typed().code)
		sub := value{code: fmt.Sprintf("%s.Get(%s)", v.at(precAtom), k.typed().code), typ: v.typ.valType(), prec: precAtom}
		if err := g.pattern(pc, p.Patterns[i], sub); err != nil {
			return err
		}
	}
	if p.Rest != "" {
		rest := value{code: fmt.Sprintf("%s(%s, %s)", g.rt("DictWithout")+"["+g.dictGoArgs(v.typ)+"]", v.code, strings.Join(keyCodes, ", ")), typ: v.typ, prec: precAtom}
		return g.captureName(pc, p.Rest, rest, p.Line())
	}
	return nil
}

func (g *generator) classPattern(pc *patternCtx, p *pyast.MatchClass, v value) error {
	name, ok := p.Cls.(*pyast.Name)
	if !ok {
		return unsupported("class pattern with a dotted class name", p.Line())
	}
	v = g.stripOptional(pc, v)
	v = g.bindSub(pc, v)

	// Built-in types: int(), str(x), float(), bool(), list(), dict(), ...
	if c := g.classes[name.ID]; c == nil || g.isVariable(name.ID) {
		return g.builtinClassPattern(pc, p, name.ID, v)
	}
	c := g.classes[name.ID]
	cur := v
	switch {
	case v.typ == tObject, v.typ == tExc && c.exc:
		var test, conv string
		if v.typ == tExc {
			test = g.rt("IsSubclass") + "(" + v.at(precAtom) + ".Type, " + strconv.Quote(c.name) + ")"
			conv = v.at(precAtom) + ".Obj.(" + c.name + ")"
		} else {
			test = g.rt("InstanceOf") + "[" + c.name + "](" + v.code + ")"
			conv = v.at(precAtom) + ".(" + c.name + ")"
		}
		pc.require(test)
		cur = g.bindSub(pc, value{code: conv, typ: c.typ(), prec: precAtom})
	case v.typ.isClass():
		vc := g.classOf(v.typ)
		switch {
		case vc.isSubclassOf(c):
		case c.isSubclassOf(vc):
			pc.require(g.rt("InstanceOf") + "[" + c.name + "](" + v.code + ")")
			cur = g.bindSub(pc, value{code: v.at(precAtom) + ".(" + c.name + ")", typ: c.typ(), prec: precAtom})
		default:
			pc.body.WriteString("return false\n")
			return nil
		}
	default:
		pc.body.WriteString("return false\n")
		return nil
	}

	pc.narrowed = &cur
	sub := func(attr string, pat pyast.Node) error {
		m := c.lookup(attr)
		if m == nil || !m.isAttr() {
			// A missing attribute simply does not match.
			if m == nil {
				pc.body.WriteString("return false\n")
				return nil
			}
			return unsupported(fmt.Sprintf("class pattern attribute '%s' of '%s' is a method", attr, c.name), p.Line())
		}
		a := value{code: cur.at(precAtom) + ".Get_" + attr + "()", typ: m.typ, prec: precAtom}
		return g.pattern(pc, pat, a)
	}
	if len(p.Patterns) > 0 {
		if len(p.Patterns) > len(c.matchArgs) {
			return unsupported(fmt.Sprintf("%s() accepts %d positional sub-patterns (%d given); define __match_args__", c.name, len(c.matchArgs), len(p.Patterns)), p.Line())
		}
		for i, pat := range p.Patterns {
			if err := sub(c.matchArgs[i], pat); err != nil {
				return err
			}
		}
	}
	for i, attr := range p.KwdAttrs {
		if err := sub(g.mangledIn(c, attr), p.KwdPatterns[i]); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) mangledIn(c *classInfo, attr string) string { return attr }

func (g *generator) builtinClassPattern(pc *patternCtx, p *pyast.MatchClass, name string, v value) error {
	var want pyType
	switch name {
	case "int":
		want = tInt
	case "str":
		want = tStr
	case "float":
		want = tFloat
	case "bool":
		want = tBool
	case "list", "dict", "set", "tuple":
	default:
		return unsupported(fmt.Sprintf("class pattern with '%s'", name), p.Line())
	}
	if len(p.KwdAttrs) > 0 || len(p.Patterns) > 1 {
		return unsupported("class pattern of a built-in type with attributes", p.Line())
	}
	switch {
	case v.typ == tObject && want != "":
		spec := typeSpec{builtin: name}
		test, err := g.instanceTest(v, spec, p.Line())
		if err != nil {
			return err
		}
		pc.require(test)
		conv := g.narrowObject(v, want)
		v = g.bindSub(pc, conv)
	case want != "":
		match := v.typ == want || (name == "int" && v.typ == tBool)
		if !match {
			pc.body.WriteString("return false\n")
			return nil
		}
	default:
		if pythonTypeName(v.typ) != name {
			pc.body.WriteString("return false\n")
			return nil
		}
	}
	pc.narrowed = &v
	if len(p.Patterns) == 1 {
		return g.pattern(pc, p.Patterns[0], v)
	}
	return nil
}

// narrowObject reads an object known to hold a builtin scalar.
func (g *generator) narrowObject(v value, want pyType) value {
	switch want {
	case tInt:
		return value{code: g.rt("AsInt") + "(" + v.code + ")", typ: tInt, prec: precAtom}
	case tStr:
		return value{code: v.at(precAtom) + ".(string)", typ: tStr, prec: precAtom}
	case tFloat:
		return value{code: v.at(precAtom) + ".(float64)", typ: tFloat, prec: precAtom}
	}
	return value{code: v.at(precAtom) + ".(bool)", typ: tBool, prec: precAtom}
}

package codegen

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// walk calls f on n and on every node below it.
func walk(n pyast.Node, f func(pyast.Node)) {
	var visit func(v reflect.Value)
	nodeType := reflect.TypeOf((*pyast.Node)(nil)).Elem()
	visit = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Interface:
			if !v.IsNil() {
				visit(v.Elem())
			}
		case reflect.Pointer:
			if v.IsNil() {
				return
			}
			if v.Type().Implements(nodeType) {
				f(v.Interface().(pyast.Node))
			}
			visit(v.Elem())
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Type().Field(i).IsExported() {
					visit(v.Field(i))
				}
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				visit(v.Index(i))
			}
		}
	}
	visit(reflect.ValueOf(n))
}

// prescan reads facts about the whole module that classes need before any
// code is generated: whether dataclasses is imported, and which attribute
// names are assigned somewhere.
func (g *generator) prescan(mod *pyast.Module) {
	var store func(t pyast.Node)
	store = func(t pyast.Node) {
		switch t := t.(type) {
		case *pyast.Attribute:
			// C.x = v and cls.x = v assign the class, not an instance.
			if recv, ok := t.Value.(*pyast.Name); ok && (recv.ID == "cls" || g.classes[recv.ID] != nil) {
				return
			}
			g.storedAttrs[t.Attr] = true
		case *pyast.Tuple:
			for _, e := range t.Elts {
				store(e)
			}
		case *pyast.List:
			for _, e := range t.Elts {
				store(e)
			}
		case *pyast.Starred:
			store(t.Value)
		}
	}
	walk(mod, func(n pyast.Node) {
		switch n := n.(type) {
		case *pyast.ImportFrom:
			if n.Module == "dataclasses" {
				g.dataclassImported = true
			}
		case *pyast.Assign:
			for _, t := range n.Targets {
				store(t)
			}
		case *pyast.AugAssign:
			store(n.Target)
		case *pyast.AnnAssign:
			store(n.Target)
		case *pyast.For:
			store(n.Target)
		}
	})
}

// inferClasses finds the types that classes need and cannot be read from
// annotations: the unannotated class attributes, which attributes instances
// assign, and the types of the instance attributes. The methods are compiled
// once with the output thrown away, and the first assignment to self.x
// declares x. Compiling repeats until no new attribute appears, since a
// method can depend on the attributes of another class.
func (g *generator) inferClasses() error {
	for _, c := range g.classOrder {
		for _, name := range c.order {
			m := c.members[name]
			if m.kind != mClassAttr {
				continue
			}
			if m.typ == "" {
				v, err := g.exprAs(m.value, "")
				if err != nil {
					return err
				}
				if v.typ == tNone || v.typ == tExc || v.typ.isClassRef() || v.typ == tOptExc {
					return unsupported(fmt.Sprintf("class attribute '%s.%s' of type %s", c.name, name, v.typ), m.line)
				}
				m.typ = v.typ
			}
			if g.storedAttrs[name] {
				m.shadow = true
			}
		}
	}

	usesRuntime := g.usesRuntime
	g.infer = true
	for round := 0; round < 8; round++ {
		g.inferGrew = false
		for _, c := range g.classOrder {
			g.inferClass(c)
		}
		if !g.inferGrew {
			break
		}
	}
	g.infer = false
	g.usesRuntime = usesRuntime
	return nil
}

// inferClass compiles the methods of a class, discarding the result.
func (g *generator) inferClass(c *classInfo) {
	for _, name := range append([]string{}, c.order...) {
		m := c.members[name]
		if m == nil || m.node == nil {
			continue
		}
		var out strings.Builder
		ctx := func(sig *funcSig) {
			saved := g.saveFunctionState()
			_, _ = g.funcBody(m.node, sig, "func f()", nil, false, &methodCtx{cls: c, definer: c, kind: m.kind, special: m.special, name: m.name})
			g.restoreFunctionState(saved)
		}
		_ = out
		ctx(m.sig)
		if m.setter != nil {
			saved := g.saveFunctionState()
			_, _ = g.funcBody(m.setNode, m.setter, "func f()", nil, false, &methodCtx{cls: c, definer: c, kind: mProperty, name: m.name})
			g.restoreFunctionState(saved)
		}
	}
}

// stmtList generates statements in order. While classes are inferred a
// statement that fails is skipped. After "if not isinstance(x, C): return" the
// rest of the block sees x as a C.
func (g *generator) stmtList(b *strings.Builder, stmts []pyast.Node) error {
	for i, s := range stmts {
		var env *scope
		var loops int
		var handlers []string
		var flow *flowCtx
		if g.infer {
			env, loops, handlers, flow = g.env, g.loopDepth, g.handlers, g.flow
		}
		if err := g.stmt(b, s); err != nil {
			if g.infer {
				g.env, g.loopDepth, g.handlers, g.flow = env, loops, handlers, flow
				continue
			}
			return err
		}
		if nar := g.earlyExitNarrowing(s); nar != nil {
			outer := g.env
			g.env = newScope(outer)
			b.WriteString("{\n")
			g.declareLocal(b, nar.name, nar.to, nar.expr)
			err := g.stmtList(b, stmts[i+1:])
			g.env = outer
			b.WriteString("}\n")
			return err
		}
	}
	return nil
}

// narrowing is a variable whose type is narrowed by an isinstance test.
type narrowing struct {
	name string
	to   pyType
	expr string // Go code of the narrowed value
}

// narrowOf reads "isinstance(x, C)" and says how x is narrowed to C.
func (g *generator) narrowOf(test pyast.Node) *narrowing {
	call, ok := test.(*pyast.Call)
	if !ok || call.HasKeywords || len(call.Args) != 2 {
		return nil
	}
	if name, _ := callName(call); name != "isinstance" || g.isVariable("isinstance") {
		return nil
	}
	arg, ok := call.Args[0].(*pyast.Name)
	if !ok {
		return nil
	}
	spec, ok := call.Args[1].(*pyast.Name)
	if !ok || g.isVariable(spec.ID) {
		return nil
	}
	cur, found := g.env.lookup(arg.ID)
	src := arg.ID
	if !found {
		v, ok := g.readGlobal(arg.ID)
		if !ok {
			return nil
		}
		cur, src = v.typ, v.code
	}
	if g.compRoot != nil || g.fnGlobals[arg.ID] {
		return nil
	}
	if k := g.classes[spec.ID]; k != nil {
		switch {
		case cur == tObject:
			return &narrowing{arg.ID, k.typ(), src + ".(" + k.name + ")"}
		case cur.isClass() && g.classOf(cur) != k && k.isSubclassOf(g.classOf(cur)):
			return &narrowing{arg.ID, k.typ(), src + ".(" + k.name + ")"}
		case cur == tExc && k.exc:
			return &narrowing{arg.ID, k.typ(), src + ".Obj.(" + k.name + ")"}
		}
		return nil
	}
	if cur == tObject {
		switch spec.ID {
		case "int":
			return &narrowing{arg.ID, tInt, g.rt("AsInt") + "(" + src + ")"}
		case "str":
			return &narrowing{arg.ID, tStr, src + ".(string)"}
		case "float":
			return &narrowing{arg.ID, tFloat, src + ".(float64)"}
		case "bool":
			return &narrowing{arg.ID, tBool, src + ".(bool)"}
		}
	}
	return nil
}

// earlyExitNarrowing recognises "if not isinstance(x, C): <leave>" without
// an else, after which x is a C.
func (g *generator) earlyExitNarrowing(s pyast.Node) *narrowing {
	n, ok := s.(*pyast.If)
	if !ok || len(n.Orelse) > 0 || len(n.Body) == 0 {
		return nil
	}
	not, ok := n.Test.(*pyast.UnaryOp)
	if !ok || not.Op != "Not" {
		return nil
	}
	switch n.Body[len(n.Body)-1].(type) {
	case *pyast.Return, *pyast.Raise, *pyast.Continue, *pyast.Break:
	default:
		return nil
	}
	return g.narrowOf(not.Operand)
}

// isExceptionName reports whether name is an exception class of the program
// or a built-in one.
func (g *generator) isUserExceptionClass(name string) *classInfo {
	if g.isVariable(name) {
		return nil
	}
	if c := g.classes[name]; c != nil && c.exc {
		return c
	}
	return nil
}

var _ = runtime.IsExceptionClass

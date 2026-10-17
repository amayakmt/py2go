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
	walkPrune(n, func(x pyast.Node) bool { f(x); return true })
}

// walkPrune is walk, except that it does not go below a node for which f
// returns false.
func walkPrune(n pyast.Node, f func(pyast.Node) bool) {
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
				if !f(v.Interface().(pyast.Node)) {
					return
				}
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
		case *pyast.ClassDef:
			// A method that returns something decides the result type of
			// the methods of that name in the other classes.
			for _, st := range n.Body {
				if fn, ok := st.(*pyast.FunctionDef); ok && fn.Returns == nil && !fn.IsAsync && !isGeneratorBody(fn.Body) && returnsValue(fn.Body) {
					g.valueMethods[fn.Name] = true
				}
			}
		case *pyast.Call:
			if f, ok := n.Func.(*pyast.Name); ok && f.ID == "_native" && len(n.Args) > 0 {
				if c, ok := n.Args[0].(*pyast.Constant); ok {
					if s, ok := c.Value.(string); ok {
						if strings.HasPrefix(s, "thread.") || strings.HasPrefix(s, "signal.") {
							g.threads = true
						}
						if strings.HasPrefix(s, "pickle.") {
							g.usesPickle = true
						}
					}
				}
			}
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
		if err := g.inferClassAttrs(c); err != nil {
			return err
		}
	}
	g.inferRounds(g.classOrder)
	g.phase = 2
	return nil
}

// inferClassAttrs types the class attributes that have no annotation, and
// marks the ones that instances assign.
func (g *generator) inferClassAttrs(c *classInfo) error {
	savedInfer := g.infer
	g.infer = true
	defer func() { g.infer = savedInfer }()
	g.classBodyCur, g.classBodySeen = c, map[string]bool{}
	defer func() { g.classBodyCur, g.classBodySeen = nil, nil }()
	return g.withSubst(c.tsubst, func() error {
		for _, name := range c.order {
			m := c.members[name]
			if m.kind != mClassAttr {
				continue
			}
			if m.typ == "" {
				v, err := g.exprAs(m.value, "")
				if err != nil && strings.Contains(err.Error(), "undefined name") {
					// A module variable that is not typed yet: refreshed later.
					m.typ = tObject
					if g.storedAttrs[name] {
						m.shadow = true
					}
					continue
				}
				if err != nil {
					return err
				}
				if v.typ == tNone || v.typ == tExc || v.typ.isClassRef() || v.typ == tOptExc {
					return unsupported(fmt.Sprintf("class attribute '%s.%s' of type %s", c.name, name, v.typ), m.line)
				}
				m.typ = v.typ
				if v.typ == tNoneType {
					// "attr = None" in a class body is a placeholder the
					// instances or subclasses fill in: an object.
					m.typ = tObject
				}
			}
			if g.storedAttrs[name] {
				m.shadow = true
			}
			g.classBodySeen[name] = true
		}
		return nil
	})
}

// inferRounds compiles the methods of the classes until no attribute is new.
func (g *generator) inferRounds(classes []*classInfo) {
	usesRuntime := g.usesRuntime
	savedInfer, savedGrew := g.infer, g.inferGrew
	g.infer = true
	for round := 0; round < 8; round++ {
		g.inferGrew = false
		for _, c := range classes {
			if c.completed || c.generic == nil {
				g.inferClass(c)
			}
		}
		if g.unifyAttrTypes(classes) {
			g.inferGrew = true
		}
		if !g.inferGrew {
			break
		}
	}
	g.infer, g.inferGrew = savedInfer, savedGrew
	g.usesRuntime = usesRuntime
}

// inferClass compiles the methods of a class, discarding the result.
func (g *generator) inferClass(c *classInfo) {
	saved := g.tsubst
	g.tsubst = c.tsubst
	defer func() { g.tsubst = saved }()
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
		if nars := dropAssigned(g.earlyExitNarrowing(s), stmts[i+1:]); len(nars) > 0 {
			outer := g.env
			g.env = newScope(outer)
			b.WriteString("{\n")
			g.bindNarrow(b, nars)
			err := g.stmtList(b, stmts[i+1:])
			g.env = outer
			b.WriteString("}\n")
			return err
		}
	}
	return nil
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

// attrCount is the number of attributes the classes have found so far.
func (g *generator) attrCount() int {
	n := 0
	for _, c := range g.classOrder {
		n += len(c.members)
	}
	return n
}

// refreshClassAttrs types again the class attributes of c that were typed
// before the module variables were: one that reads a module variable was an
// object. It runs when the class is generated, where the variables exist.
func (g *generator) refreshClassAttrs(c *classInfo) {
	g.classBodyCur, g.classBodySeen = c, map[string]bool{}
	defer func() { g.classBodyCur, g.classBodySeen = nil, nil }()
	_ = g.withSubst(c.tsubst, func() error {
		for _, name := range c.order {
			m := c.members[name]
			if m.kind == mClassAttr {
				g.classBodySeen[name] = true
			}
			if m.kind != mClassAttr || m.typ != tObject || m.value == nil || m.wide {
				continue
			}
			delete(g.classBodySeen, name)
			v, err := g.exprAs(m.value, "")
			if err != nil || v.typ == tObject || v.typ == tNone || v.typ == tNoneType || v.typ == tExc || v.typ.isClassRef() || v.typ == tOptExc {
				continue
			}
			baseIsObject := false
			for _, k := range c.mro[1:] {
				if bm := k.members[name]; bm != nil && (bm.kind == mClassAttr || bm.kind == mField) && bm.typ == tObject {
					baseIsObject = true
				}
			}
			if baseIsObject {
				continue // a base holds any value here, so this class does too
			}
			m.typ = v.typ
		}
		return nil
	})
}

// unifyAttrTypes makes an attribute an object in every class of a hierarchy
// where classes give it different types (Node.childNodes a NodeList,
// Childless.childNodes an EmptyNodeList): Python lets them differ, the Go
// interface of the base cannot. It reports whether a type changed.
func (g *generator) unifyAttrTypes(classes []*classInfo) bool {
	changed := false
	for _, c := range classes {
		if c.generic != nil {
			continue
		}
		for _, name := range c.visibleNames() {
			var first pyType
			mixed := false
			var attrs []*member
			for _, k := range c.mro {
				m := k.members[name]
				if m == nil || (m.kind != mField && m.kind != mClassAttr) || m.typ == "" {
					continue
				}
				attrs = append(attrs, m)
				if first == "" {
					first = m.typ
				} else if m.typ != first {
					mixed = true
				}
			}
			if mixed {
				for _, m := range attrs {
					m.wide = true
					if m.typ != tObject {
						m.typ = tObject
						changed = true
					}
				}
			}
		}
	}
	return changed
}

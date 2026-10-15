package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// Classes. A class the program defines is a Go interface of the same name,
// which is the static type of every variable that holds an instance, and one
// Go struct per class (_pyc_Name) that implements it. The interface lists the
// accessors and methods the class has, so a value of a subclass is accepted
// wherever a base class is expected and calls go through the interface.
//
// Every method is compiled once for each class that inherits it, with self
// bound to that exact class. Inside a method, self.x and super() are then
// resolved at compile time, and the cooperative multiple inheritance of
// super() needs no run-time method resolution order.

type memberKind int

const (
	mField       memberKind = iota // an instance attribute
	mClassAttr                     // an attribute of the class, shared by its instances
	mMethod                        // def m(self, ...)
	mStatic                        // @staticmethod
	mClassMethod                   // @classmethod
	mProperty                      // @property
)

// member is one name defined in a class body, or assigned through self.
type member struct {
	kind  memberKind
	name  string
	owner *classInfo
	typ   pyType // field, class attribute and property type
	line  int

	// Methods: the signature without self, and the def. A property has the
	// getter here and the setter in setter.
	sig     *funcSig
	node    *pyast.FunctionDef
	setter  *funcSig
	setNode *pyast.FunctionDef

	// special marks the methods that may return NotImplemented. They return
	// (result, implemented) in Go.
	special bool

	// value is the initializer of a class attribute.
	value pyast.Node
	// shadow marks a class attribute that some code assigns through an
	// instance, which then gets an attribute of its own.
	shadow bool

	// synth names a method that is generated, not written: a dataclass's
	// __repr__, __eq__ and friends.
	synth string

	// A dataclass field's init parameter, if it has one.
	dcField *dcField
}

func (m *member) isCallable() bool {
	return m.kind == mMethod || m.kind == mStatic || m.kind == mClassMethod
}

// classInfo is everything the transpiler knows about a class.
type classInfo struct {
	name    string
	node    *pyast.ClassDef
	line    int
	builtin bool // object or a built-in exception class: a name and nothing else
	index   int  // position among the module's statements

	bases      []*classInfo
	mro        []*classInfo
	members    map[string]*member
	order      []string // member names in the order they were first seen
	doc        string
	hasDoc     bool
	unhashable bool // __hash__ = None
	dc         *dcInfo

	matchArgs []string // __match_args__, for class patterns
	exc       bool     // derives from a built-in exception class
	excRoot   string   // the nearest built-in exception class in the MRO

	// The constructor: the __init__ in effect and where it is defined.
	ctorSig   *funcSig
	ctorOwner *classInfo

	emitted bool

	// Instantiations of a generic class: the name Python shows, the type
	// arguments, and the template they come from.
	display   string
	tsubst    map[string]pyType
	generic   *genericClass
	targs     []pyType
	completed bool
}

// pyName is the name of the class in Python, which an instantiation of a
// generic class shares with its template.
func (c *classInfo) pyName() string {
	if c.display != "" {
		return c.display
	}
	return pyName(c.name)
}

func (c *classInfo) goStruct() string { return "_pyc_" + c.name }
func (c *classInfo) goNew() string    { return "_pyNew_" + c.name }
func (c *classInfo) typ() pyType      { return pyType(c.name) }

// lookup finds the member a name resolves to on instances of c: the first
// class in the MRO that defines it.
func (c *classInfo) lookup(name string) *member {
	for _, k := range c.mro {
		if m := k.members[name]; m != nil {
			return m
		}
	}
	return nil
}

// isSubclassOf reports whether c is d or derives from it.
func (c *classInfo) isSubclassOf(d *classInfo) bool {
	for _, k := range c.mro {
		if k == d {
			return true
		}
	}
	return false
}

func (c *classInfo) addMember(m *member) {
	if _, dup := c.members[m.name]; !dup {
		c.order = append(c.order, m.name)
	}
	c.members[m.name] = m
}

// specialDunders are the methods whose result may be NotImplemented.
var specialDunders = func() map[string]bool {
	m := map[string]bool{}
	for _, op := range []string{"add", "sub", "mul", "truediv", "floordiv", "mod", "pow", "matmul", "and", "or", "xor", "lshift", "rshift"} {
		m["__"+op+"__"] = true
		m["__r"+op+"__"] = true
		m["__i"+op+"__"] = true
	}
	for _, op := range []string{"eq", "ne", "lt", "le", "gt", "ge"} {
		m["__"+op+"__"] = true
	}
	return m
}()

// builtinClass is the stand-in for object and the built-in exception classes,
// which user classes can derive from.
func (g *generator) builtinClass(name string) *classInfo {
	if c, ok := g.builtinCls[name]; ok {
		return c
	}
	c := &classInfo{name: name, builtin: true, members: map[string]*member{}}
	g.builtinCls[name] = c
	if name == "object" {
		c.mro = []*classInfo{c}
		return c
	}
	chain := runtime.BuiltinMRO(name)
	c.exc = true
	c.excRoot = name
	c.mro = []*classInfo{c}
	for _, n := range chain[1:] {
		c.mro = append(c.mro, g.builtinClass(n))
	}
	c.mro = append(c.mro, g.builtinClass("object"))
	return c
}

// c3 computes the method resolution order of a class from its bases.
func c3(c *classInfo, line int) ([]*classInfo, error) {
	seqs := make([][]*classInfo, 0, len(c.bases)+1)
	for _, b := range c.bases {
		seqs = append(seqs, append([]*classInfo{}, b.mro...))
	}
	seqs = append(seqs, append([]*classInfo{}, c.bases...))
	out := []*classInfo{c}
	for {
		nonEmpty := seqs[:0]
		for _, s := range seqs {
			if len(s) > 0 {
				nonEmpty = append(nonEmpty, s)
			}
		}
		seqs = nonEmpty
		if len(seqs) == 0 {
			return out, nil
		}
		var pick *classInfo
		for _, s := range seqs {
			cand := s[0]
			blocked := false
			for _, t := range seqs {
				for _, x := range t[1:] {
					if x == cand {
						blocked = true
					}
				}
			}
			if !blocked {
				pick = cand
				break
			}
		}
		if pick == nil {
			names := make([]string, len(c.bases))
			for i, b := range c.bases {
				names[i] = b.name
			}
			return nil, unsupported("Cannot create a consistent method resolution order (MRO) for bases "+strings.Join(names, ", "), line)
		}
		out = append(out, pick)
		for i, s := range seqs {
			if s[0] == pick {
				seqs[i] = s[1:]
			}
		}
	}
}

// reservedClassNames cannot be redefined by the program.
func (g *generator) reservedClassName(name string) bool {
	if _, ok := typeTable[name]; ok || name == "object" || runtime.IsExceptionClass(name) {
		return true
	}
	return builtinFunctions[name] || name == "print" || name == "type" || name == "isinstance" || name == "super"
}

// registerClassNames makes every top-level class known by name, so that
// annotations anywhere can refer to it.
func (g *generator) registerClassNames(mod *pyast.Module) error {
	g.modBody = mod.Body
	for i, stmt := range mod.Body {
		cd, ok := stmt.(*pyast.ClassDef)
		if !ok {
			continue
		}
		if err := checkName(cd.Name, cd.Line()); err != nil {
			return err
		}
		if g.reservedClassName(cd.Name) {
			return unsupported(fmt.Sprintf("class named '%s' (a built-in name)", cd.Name), cd.Line())
		}
		if _, dup := g.classes[cd.Name]; dup || g.genericClasses[cd.Name] != nil {
			return unsupported(fmt.Sprintf("redefinition of class '%s'", cd.Name), cd.Line())
		}
		if len(cd.Keywords) > 0 {
			return unsupported("class keyword '"+cd.Keywords[0].Arg+"=' (metaclasses arrive in phase 9)", cd.Line())
		}
		params, err := g.classTypeParams(cd)
		if err != nil {
			return err
		}
		if len(params) > 0 {
			g.genericClasses[cd.Name] = &genericClass{name: cd.Name, node: cd, params: params, index: i}
			continue
		}
		c := &classInfo{name: cd.Name, node: cd, line: cd.Line(), index: i, members: map[string]*member{}}
		g.classes[cd.Name] = c
		g.classOrder = append(g.classOrder, c)
	}
	return nil
}

// collectClasses reads every top-level class statement: bases, MRO, members
// and dataclass options. Nothing is generated yet.
func (g *generator) collectClasses(mod *pyast.Module) error {
	for _, c := range g.classOrder {
		if _, clash := g.sigs[c.name]; clash {
			return unsupported(fmt.Sprintf("'%s' is both a function and a class", c.name), c.line)
		}
		if _, clash := g.sigErrs[c.name]; clash {
			return unsupported(fmt.Sprintf("'%s' is both a function and a class", c.name), c.line)
		}
	}
	var plain []*classInfo
	for _, c := range g.classOrder {
		if c.generic == nil {
			plain = append(plain, c)
		}
	}
	for _, c := range plain {
		if err := g.classBases(c); err != nil {
			return err
		}
	}
	for _, c := range plain {
		if err := g.classBody(c); err != nil {
			return err
		}
	}
	for _, c := range plain {
		if err := g.finishClass(c); err != nil {
			return err
		}
	}
	g.phase = 1
	return g.completeInstances()
}

// classBases resolves the base classes and computes the MRO. A base has to
// be defined above the class that uses it.
func (g *generator) classBases(c *classInfo) error {
	cd := c.node
	for _, b := range cd.Bases {
		if sub, ok := b.(*pyast.Subscript); ok {
			head, isName := sub.Value.(*pyast.Name)
			if isName && head.ID == "Generic" {
				continue
			}
			if isName && g.genericClasses[head.ID] != nil {
				t, err := g.annotationType(sub, false)
				if err != nil {
					return err
				}
				base := g.classes[string(t)]
				if base.mro == nil {
					if err := g.completeInstance(base); err != nil {
						return err
					}
				}
				c.bases = append(c.bases, base)
				continue
			}
		}
		name, ok := b.(*pyast.Name)
		if !ok {
			return unsupported("base class that is not a plain name", b.Line())
		}
		if gc := g.genericClasses[name.ID]; gc != nil {
			return unsupported(fmt.Sprintf("generic class '%s' as a base without type arguments", name.ID), b.Line())
		}
		switch {
		case g.classes[name.ID] != nil:
			base := g.classes[name.ID]
			if base.index >= c.index && c.generic == nil {
				return unsupported(fmt.Sprintf("base class '%s' used before its definition", name.ID), b.Line())
			}
			if base == c {
				return unsupported("class derived from itself", b.Line())
			}
			if base.mro == nil {
				// Defined above, so its bases are resolved already.
				return unsupported(fmt.Sprintf("base class '%s' used before its definition", name.ID), b.Line())
			}
			c.bases = append(c.bases, base)
		case name.ID == "object":
			c.bases = append(c.bases, g.builtinClass("object"))
		case runtime.IsExceptionClass(name.ID):
			c.bases = append(c.bases, g.builtinClass(name.ID))
		default:
			why := "unknown base class '" + name.ID + "'"
			switch name.ID {
			case "Enum", "IntEnum", "Flag", "IntFlag", "StrEnum":
				why = "enumerations arrive in phase 12 with the enum module"
			case "ABC", "Protocol", "Generic", "NamedTuple", "TypedDict":
				why = "base class '" + name.ID + "' arrives with the typing and abc modules in phase 12"
			case "dict", "list", "set", "str", "int", "float", "tuple", "bool":
				why = "deriving from built-in type '" + name.ID + "' arrives in phase 9"
			}
			return unsupported(why, b.Line())
		}
	}
	if len(c.bases) == 0 {
		c.bases = []*classInfo{g.builtinClass("object")}
	}
	seen := map[*classInfo]bool{}
	for _, b := range c.bases {
		if seen[b] {
			return unsupported("duplicate base class '"+b.name+"'", cd.Line())
		}
		seen[b] = true
	}
	mro, err := c3(c, cd.Line())
	if err != nil {
		return err
	}
	c.mro = mro
	for _, k := range mro {
		if k.builtin && k.exc {
			c.exc = true
			if c.excRoot == "" {
				c.excRoot = k.name
			}
		}
	}
	return nil
}

// classBody reads the statements of a class body into members.
func (g *generator) classBody(c *classInfo) error {
	cd := c.node
	dc, err := g.dataclassOptions(c)
	if err != nil {
		return err
	}
	c.dc = dc

	stmts := cd.Body
	if doc, ok := docstring(stmts); ok {
		c.doc, c.hasDoc = doc, true
		stmts = stmts[1:]
	}
	for _, st := range stmts {
		switch n := st.(type) {
		case *pyast.Pass:
		case *pyast.Expr:
			if _, ok := n.Value.(*pyast.Constant); ok {
				continue
			}
			if isEllipsis(n.Value) {
				continue
			}
			return unsupported("expression statement in a class body", n.Line())
		case *pyast.FunctionDef:
			if err := g.classMethod(c, n); err != nil {
				return err
			}
		case *pyast.AnnAssign:
			if err := g.classAnnAssign(c, n); err != nil {
				return err
			}
		case *pyast.Assign:
			if err := g.classAssign(c, n); err != nil {
				return err
			}
		case *pyast.ClassDef:
			return unsupported("class defined inside a class", n.Line())
		case *pyast.Unsupported:
			return unsupported(n.Type+" in a class body", n.Line())
		default:
			return unsupported(nodeName(n)+" in a class body", n.Line())
		}
	}
	return nil
}

// classMethod reads one def in a class body.
func (g *generator) classMethod(c *classInfo, fn *pyast.FunctionDef) error {
	kind := mMethod
	setterOf := ""
	for _, d := range fn.Decorators {
		switch d := d.(type) {
		case *pyast.Name:
			switch d.ID {
			case "staticmethod":
				kind = mStatic
			case "classmethod":
				kind = mClassMethod
			case "property":
				kind = mProperty
			default:
				return unsupported(fmt.Sprintf("decorator '%s' on a method", d.ID), d.Line())
			}
		case *pyast.Attribute:
			base, ok := d.Value.(*pyast.Name)
			if !ok || base.ID != fn.Name || (d.Attr != "setter" && d.Attr != "deleter") {
				return unsupported("decorator on a method", d.Line())
			}
			if d.Attr == "deleter" {
				return unsupported("property deleter", d.Line())
			}
			kind = mProperty
			setterOf = fn.Name
		default:
			return unsupported("decorator on a method", d.Line())
		}
	}
	if len(fn.Decorators) > 1 {
		return unsupported("more than one decorator on a method", fn.Line())
	}
	name := fn.Name
	if strings.HasPrefix(name, "__") && !strings.HasSuffix(name, "__") {
		name = "_" + c.name + name // private names are mangled
	}
	switch name {
	case "__new__", "__init_subclass__", "__class_getitem__", "__set_name__":
		return unsupported("method "+name+" (arrives in phase 9)", fn.Line())
	case "__getattribute__", "__setattr__", "__delattr__", "__get__", "__set__", "__delete__":
		return unsupported("method "+name+" (the attribute hooks and descriptors arrive in phase 9)", fn.Line())
	case "__del__", "__sizeof__", "__copy__", "__deepcopy__":
		return unsupported("method "+name, fn.Line())
	}
	if prev := c.members[name]; prev != nil && setterOf == "" {
		return unsupported(fmt.Sprintf("redefinition of '%s.%s'", c.name, fn.Name), fn.Line())
	}

	sig, err := g.methodSig(c, fn, kind, setterOf != "")
	if err != nil {
		return err
	}
	qual := c.pyName() + "." + fn.Name
	sig.qualname = qual

	if setterOf != "" {
		prop := c.members[setterOf]
		if prop == nil || prop.kind != mProperty || prop.setter != nil {
			return unsupported(fmt.Sprintf("property setter '%s' without a property", fn.Name), fn.Line())
		}
		if len(sig.pos) != 1 || sig.vararg != nil || sig.kwarg != nil || len(sig.kwonly) > 0 {
			return unsupported("property setter that does not take exactly one value", fn.Line())
		}
		if sig.pos[0].typ != prop.typ {
			return unsupported(fmt.Sprintf("property setter of type %s for a property of type %s", sig.pos[0].typ, prop.typ), fn.Line())
		}
		if sig.ret != tNone {
			return unsupported("property setter that returns a value", fn.Line())
		}
		prop.setter, prop.setNode = sig, fn
		return nil
	}

	m := &member{kind: kind, name: name, owner: c, sig: sig, node: fn, line: fn.Line()}
	if kind == mProperty {
		if len(sig.pos) != 0 || sig.vararg != nil || sig.kwarg != nil || len(sig.kwonly) > 0 {
			return unsupported("property getter that takes parameters", fn.Line())
		}
		if sig.ret == tNone && len(fn.Body) > 0 {
			if _, raises := fn.Body[len(fn.Body)-1].(*pyast.Raise); raises {
				sig.ret = tObject // a getter that only raises, as abstract bases write them
			}
		}
		if sig.ret == tNone {
			return unsupported("property getter that returns nothing", fn.Line())
		}
		m.typ = sig.ret
	}
	if kind == mMethod && specialDunders[name] {
		m.special = true
	}
	if name == "__init__" && sig.ret != tNone {
		return unsupported("__init__ that returns a value", fn.Line())
	}
	c.addMember(m)
	return nil
}

// methodSig builds the signature of a method without its self or cls, which
// counts as an implicit argument in error messages.
func (g *generator) methodSig(c *classInfo, fn *pyast.FunctionDef, kind memberKind, isSetter bool) (*funcSig, error) {
	cp := *fn
	if kind != mStatic {
		args := *fn.Args
		switch {
		case len(args.PosOnly) > 0:
			args.PosOnly = args.PosOnly[1:]
		case len(args.Args) > 0:
			args.Args = args.Args[1:]
		default:
			if args.Vararg != nil {
				return nil, unsupported("method whose first parameter is *args", fn.Line())
			}
			return nil, unsupported(fmt.Sprintf("method '%s' without a self parameter", fn.Name), fn.Line())
		}
		cp.Args = &args
	}
	cp.Decorators = nil
	g.buildingMethod = true
	g.buildingProperty = kind == mProperty || len(fn.Decorators) > 0 && hasPropertyDecorator(fn)
	sig, err := g.buildSig(&cp, c.pyName()+"."+fn.Name, true)
	g.buildingMethod, g.buildingProperty = false, false
	if err != nil {
		return nil, err
	}
	if kind != mStatic {
		sig.implicit = 1
		first := fn.Args.Positional()[0]
		if err := checkName(first.Name, first.Line()); err != nil {
			return nil, err
		}
		sig.selfName = first.Name
	}
	ps, _ := sig.defaultNodes()
	for i, p := range ps {
		p.defaultVar = fmt.Sprintf("_pyd_%s_%s_%d", c.name, fn.Name, i)
		if isSetter {
			p.defaultVar += "s"
		}
	}
	return sig, nil
}

// classAnnAssign is "x: int" and "x: int = 0" in a class body.
func (g *generator) classAnnAssign(c *classInfo, n *pyast.AnnAssign) error {
	target, ok := n.Target.(*pyast.Name)
	if !ok || !n.Simple {
		return unsupported("annotated assignment to "+targetKind(n.Target), n.Line())
	}
	name := target.ID
	if strings.HasPrefix(name, "__") && !strings.HasSuffix(name, "__") {
		name = "_" + c.name + name
	}
	ann := n.Annotation
	classVar := false
	if sub, ok := ann.(*pyast.Subscript); ok {
		if head, ok := sub.Value.(*pyast.Name); ok && head.ID == "ClassVar" {
			ann, classVar = sub.Index, true
		}
	}
	if head, ok := ann.(*pyast.Name); ok && head.ID == "ClassVar" {
		return unsupported("ClassVar without a type", n.Line())
	}
	t, err := g.annotationType(ann, false)
	if err != nil {
		return err
	}
	if prev := c.members[name]; prev != nil {
		return unsupported(fmt.Sprintf("redefinition of '%s.%s'", c.name, name), n.Line())
	}
	switch {
	case classVar:
		if n.Value == nil {
			return unsupported("ClassVar without a value", n.Line())
		}
		c.addMember(&member{kind: mClassAttr, name: name, owner: c, typ: t, value: n.Value, line: n.Line()})
	case c.dc != nil:
		f, err := g.dataclassField(c, name, t, n)
		if err != nil {
			return err
		}
		c.dc.own = append(c.dc.own, f)
		c.addMember(&member{kind: mField, name: name, owner: c, typ: t, line: n.Line(), dcField: f})
	case n.Value == nil:
		c.addMember(&member{kind: mField, name: name, owner: c, typ: t, line: n.Line()})
	default:
		c.addMember(&member{kind: mClassAttr, name: name, owner: c, typ: t, value: n.Value, line: n.Line()})
	}
	return nil
}

// classAssign is "x = value" in a class body: a class attribute.
func (g *generator) classAssign(c *classInfo, n *pyast.Assign) error {
	if len(n.Targets) != 1 {
		return unsupported("chained assignment in a class body", n.Line())
	}
	target, ok := n.Targets[0].(*pyast.Name)
	if !ok {
		return unsupported("assignment to "+targetKind(n.Targets[0])+" in a class body", n.Line())
	}
	name := target.ID
	// alias = method: the method again under another name.
	if src, ok := n.Value.(*pyast.Name); ok && c.members[src.ID] != nil && c.members[src.ID].node != nil &&
		(c.members[src.ID].kind == mMethod || c.members[src.ID].kind == mStatic || c.members[src.ID].kind == mClassMethod) {
		fn := *c.members[src.ID].node
		fn.Name = name
		return g.classMethod(c, &fn)
	}
	switch name {
	case "__match_args__":
		tup, ok := n.Value.(*pyast.Tuple)
		if !ok {
			return unsupported("__match_args__ that is not a tuple of names", n.Line())
		}
		c.matchArgs = []string{}
		for _, e := range tup.Elts {
			s, ok := e.(*pyast.Constant)
			if !ok || s.PyType != "str" {
				return unsupported("__match_args__ that is not a tuple of names", n.Line())
			}
			c.matchArgs = append(c.matchArgs, s.Value.(string))
		}
		return nil
	case "__slots__", "__annotations__":
		return nil
	case "__hash__":
		if k, ok := n.Value.(*pyast.Constant); ok && k.PyType == "none" {
			c.unhashable = true
			return nil
		}
		return unsupported("__hash__ assigned something other than None", n.Line())
	}
	if strings.HasPrefix(name, "__") && !strings.HasSuffix(name, "__") {
		name = "_" + c.name + name
	}
	if strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__") {
		return unsupported("class attribute "+name, n.Line())
	}
	if c.members[name] != nil {
		return unsupported(fmt.Sprintf("redefinition of '%s.%s'", c.name, name), n.Line())
	}
	if c.dc != nil {
		return unsupported("unannotated attribute in a dataclass (annotate it, or use ClassVar)", n.Line())
	}
	c.addMember(&member{kind: mClassAttr, name: name, owner: c, value: n.Value, line: n.Line()})
	return nil
}

// finishClass completes a class once all classes have their members: the
// dataclass methods, the constructor, and the check that every override
// has the signature of what it overrides.
func (g *generator) finishClass(c *classInfo) error {
	if err := g.externalClassAttrs(c); err != nil {
		return err
	}
	if c.dc != nil {
		if err := g.dataclassMembers(c); err != nil {
			return err
		}
	}
	// The constructor is the first __init__ in the MRO.
	for _, k := range c.mro {
		if m := k.members["__init__"]; m != nil {
			if m.kind != mMethod {
				return unsupported("__init__ that is not an instance method", m.line)
			}
			c.ctorSig, c.ctorOwner = m.sig, k
			break
		}
	}
	return nil
}

// visibleNames lists the attribute names instances of c have, from the whole
// MRO, in a stable order.
func (c *classInfo) visibleNames() []string {
	names := c.orderedNames()
	sort.Strings(names)
	return names
}

// orderedNames is visibleNames in definition order, base classes first: the
// order of the attributes in an instance's __dict__.
func (c *classInfo) orderedNames() []string {
	seen := map[string]bool{}
	var names []string
	for i := len(c.mro) - 1; i >= 0; i-- {
		for _, n := range c.mro[i].order {
			if !seen[n] {
				seen[n] = true
				names = append(names, n)
			}
		}
	}
	return names
}

// shape is what the Go interface of a class says about a member: kinds and
// types, which every provider of the name in the MRO has to agree on.
func (m *member) shape(g *generator) string {
	switch m.kind {
	case mField, mClassAttr:
		return "attr " + string(m.typ)
	case mProperty:
		return "attr " + string(m.typ)
	}
	var b strings.Builder
	b.WriteString("method(")
	for _, p := range m.sig.goParams() {
		b.WriteString(string(p.typ) + ",")
	}
	b.WriteString(")" + string(m.sig.ret))
	if m.special {
		b.WriteString("!")
	}
	return b.String()
}

// checkOverrides makes sure the Go interfaces fit together: a name defined
// in several classes of the MRO has the same type in all of them. Python
// does not require it, Go does.
func (g *generator) checkOverrides(c *classInfo) error {
	for _, name := range c.visibleNames() {
		if name == "__init__" || name == "__post_init__" {
			continue // constructors are not part of the interface
		}
		var winner *member
		for _, k := range c.mro {
			m := k.members[name]
			if m == nil {
				continue
			}
			if winner == nil {
				winner = m
				continue
			}
			if (winner.kind == mClassMethod || winner.kind == mStatic) && (m.kind == mClassMethod || m.kind == mStatic) {
				continue // called on the class, so each class resolves its own
			}
			if winner.shape(g) != m.shape(g) {
				return unsupported(fmt.Sprintf("'%s.%s' overrides '%s.%s' with a different signature or type (%s vs %s)",
					winner.owner.name, shortName(name), m.owner.name, shortName(name), winner.shape(g), m.shape(g)), winner.line)
			}
			// A property that is read-only cannot stand for an attribute that
			// can be assigned, and the reverse.
			if (winner.kind == mProperty && winner.setter == nil) != (m.kind == mProperty && m.setter == nil) &&
				(winner.kind == mProperty || m.kind == mProperty) {
				return unsupported(fmt.Sprintf("'%s.%s' and '%s.%s' differ in whether they can be assigned",
					winner.owner.name, shortName(name), m.owner.name, shortName(name)), winner.line)
			}
		}
	}
	return nil
}

func shortName(n string) string { return n }

// externalClassAttrs: "C.x = v" at the top of the module declares a class
// attribute x whose value is computed where the class statement stands,
// unless the class itself already has x (a subclass shadows the attribute it inherits).
func (g *generator) externalClassAttrs(c *classInfo) error {
	cd := c.node
	// "C.x = v" at the top of the module, after the class, declares a class
	// attribute whose value is computed where the class statement stands.
	for _, st := range g.modBody {
		a, ok := st.(*pyast.Assign)
		if !ok || len(a.Targets) != 1 {
			continue
		}
		at, ok := a.Targets[0].(*pyast.Attribute)
		if !ok {
			continue
		}
		if recv, ok := at.Value.(*pyast.Name); !ok || recv.ID != cd.Name || c.members[g.mangled(at.Attr)] != nil {
			continue
		}
		synthetic := &pyast.Assign{Targets: []pyast.Node{&pyast.Name{ID: at.Attr, Lineno: a.Lineno}}, Value: a.Value, Lineno: a.Lineno}
		if err := g.classAssign(c, synthetic); err != nil {
			return err
		}
		g.externalAttr[a] = true
	}
	return nil
}

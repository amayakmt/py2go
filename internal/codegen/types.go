package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// pyType is the static type of an expression or variable. It is a string
// so types compare with == and work as map keys: the scalars are "int",
// "float", "bool" and "str", containers are spelled like their annotation,
// "list[int]" and "tuple[int, str]", and a type is a container of the types
// inside the brackets.
type pyType string

const (
	tInt   pyType = "int"
	tFloat pyType = "float"
	tBool  pyType = "bool"
	tStr   pyType = "str"
	tRange pyType = "range"
	tNone  pyType = "None" // only the result of a call to a function that returns None

	// tObject is the annotation "object": anything, held as a Go any.
	tObject pyType = "object"

	// tNoneType is the type of the value None itself, before any Optional
	// gives it a home.
	tNoneType pyType = "NoneType"

	// tOptExc is the type of e.__cause__ and e.__context__: an exception, or
	// None, which is a nil pointer. It can be tested with "is None", printed,
	// and its class name read.
	tOptExc pyType = "exception|None"

	// tExc is the type of the name bound by "except ... as e". It cannot be
	// written as an annotation, so an exception never outlives its handler.
	tExc pyType = "exception"
)

// maxTupleArity is the largest tuple the runtime has a struct for.
const maxTupleArity = 6

func (t pyType) String() string { return string(t) }

func listOf(elem pyType) pyType { return "list[" + elem + "]" }
func setOf(elem pyType) pyType  { return "set[" + elem + "]" }
func iterOf(elem pyType) pyType { return "iter[" + elem + "]" }

// iterableOfType is the type of a parameter annotated Iterable[T]: any
// list, set, dict, range, view or iterator of T can be passed for it.
func iterableOfType(elem pyType) pyType { return "iterable[" + elem + "]" }

func varTupleOf(elem pyType) pyType { return pyType("tuple[" + string(elem) + ", ...]") }

// funcOf is the type of a function taking params and returning ret. The
// return type is the last type argument, tNone for a function that
// returns nothing.
func funcOf(params []pyType, ret pyType) pyType {
	parts := make([]string, 0, len(params)+1)
	for _, p := range params {
		parts = append(parts, string(p))
	}
	parts = append(parts, string(ret))
	return pyType("func[" + strings.Join(parts, ", ") + "]")
}

func (t pyType) funcParams() []pyType {
	args := t.typeArgs()
	return args[:len(args)-1]
}

func (t pyType) funcRet() pyType {
	args := t.typeArgs()
	return args[len(args)-1]
}

func dictOf(key, val pyType) pyType { return pyType("dict[" + string(key) + ", " + string(val) + "]") }

func tupleOf(elems []pyType) pyType {
	parts := make([]string, len(elems))
	for i, e := range elems {
		parts[i] = string(e)
	}
	return pyType("tuple[" + strings.Join(parts, ", ") + "]")
}

// Containers are written name[args]. head is the name and typeArgs the
// arguments, split at the top level only.
func (t pyType) head() string {
	if i := strings.IndexByte(string(t), '['); i >= 0 {
		return string(t)[:i]
	}
	return string(t)
}

func (t pyType) typeArgs() []pyType {
	i := strings.IndexByte(string(t), '[')
	if i < 0 {
		return nil
	}
	inner := strings.TrimSuffix(string(t)[i+1:], "]")
	var out []pyType
	depth, start := 0, 0
	for i := 0; i < len(inner); i++ {
		switch inner[i] {
		case '[':
			depth++
		case ']':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, pyType(inner[start:i]))
				start = i + 2 // skip ", "
			}
		}
	}
	return append(out, pyType(inner[start:]))
}

func (t pyType) isList() bool { return t.head() == "list" }

// isTuple is a tuple of fixed length, tuple[int, str]. A tuple[int, ...]
// is a different type, with a different representation.
func (t pyType) isTuple() bool { return t.head() == "tuple" && !t.isVarTuple() }

func (t pyType) isVarTuple() bool {
	if t.head() != "tuple" {
		return false
	}
	args := t.typeArgs()
	return len(args) == 2 && args[1] == "..."
}

// isClass is the type of an instance of a class the program defines. Its
// pyType is the class name, and the Go type an interface of the same name.
func (t pyType) isClass() bool {
	switch t {
	case "", tInt, tFloat, tBool, tStr, tRange, tNone, tNoneType, tExc, tObject, tOptExc:
		return false
	}
	return !strings.ContainsAny(string(t), "[,|")
}

// classRefOf is the type of a class used as a value, such as cls in a
// classmethod. Only calls and attribute reads use it.
func classRefOf(name string) pyType { return pyType("type[" + name + "]") }
func (t pyType) isClassRef() bool   { return t.head() == "type" && strings.Contains(string(t), "[") }
func (t pyType) classRefName() string {
	return strings.TrimSuffix(strings.TrimPrefix(string(t), "type["), "]")
}

func optionalOf(elem pyType) pyType {
	if elem.isOptional() {
		return elem
	}
	return pyType("optional[" + string(elem) + "]")
}

// isOptional is Optional[T], which is T or None.
func (t pyType) isOptional() bool { return t.head() == "optional" }
func (t pyType) optElem() pyType  { return t.typeArgs()[0] }

// genOf is a generator that yields y, is sent s and returns r.
func genOf(y, s, r pyType) pyType {
	return pyType("gen[" + string(y) + ", " + string(s) + ", " + string(r) + "]")
}

func (t pyType) isGen() bool { return t.head() == "gen" }

func (t pyType) isFunc() bool       { return t.head() == "func" }
func (t pyType) isIterable() bool   { return t.head() == "iterable" }
func (t pyType) isDict() bool       { return t.head() == "dict" }
func (t pyType) isSet() bool        { return t.head() == "set" }
func (t pyType) isIter() bool       { return t.head() == "iter" }
func (t pyType) isKeysView() bool   { return t.head() == "dict_keys" }
func (t pyType) isValuesView() bool { return t.head() == "dict_values" }
func (t pyType) isItemsView() bool  { return t.head() == "dict_items" }

// isView is any of the dict views.
func (t pyType) isView() bool { return t.isKeysView() || t.isValuesView() || t.isItemsView() }

// elem is the element type of a list, set or iterator.
func (t pyType) elem() pyType { return t.typeArgs()[0] }

// elems are the element types of a tuple, in order.
func (t pyType) elems() []pyType { return t.typeArgs() }

// keyType and valType are the parameters of a dict or one of its views.
func (t pyType) keyType() pyType { return t.typeArgs()[0] }
func (t pyType) valType() pyType { return t.typeArgs()[1] }

// iterElem is the type of the items a for loop gets from a value of type t,
// or false if t cannot be iterated. A tuple cannot, because its items have
// different types.
func (t pyType) iterElem() (pyType, bool) {
	switch {
	case t == tStr:
		return tStr, true
	case t == tRange:
		return tInt, true
	case t.isList(), t.isSet(), t.isIter(), t.isVarTuple(), t.isIterable(), t.isGen():
		return t.elem(), true
	case t.isDict(), t.isKeysView():
		return t.keyType(), true
	case t.isValuesView():
		return t.valType(), true
	case t.isItemsView():
		return tupleOf([]pyType{t.keyType(), t.valType()}), true
	}
	return "", false
}

// hashable reports whether a value of type t can be a dict key or set
// element: the scalars, and tuples of hashable values.
func (t pyType) hashable() bool {
	switch {
	case t == tInt, t == tBool, t == tStr, t == tFloat, t == tNoneType:
		return true
	case t.isTuple():
		for _, e := range t.elems() {
			if !e.hashable() {
				return false
			}
		}
		return true
	}
	return false
}

// goType is the Go spelling of the type. Lists, dicts and sets are pointers
// because Python's containers alias, and tuples are the runtime's generic
// structs.
func (t pyType) goType() string {
	switch {
	case t == tInt:
		return "int64"
	case t == tFloat:
		return "float64"
	case t == tBool:
		return "bool"
	case t == tStr:
		return "string"
	case t == tExc, t == tOptExc:
		return "*runtime.PyError"
	case t == tObject:
		return "any"
	case t == tNoneType:
		return "runtime.NoneT"
	case t.isOptional():
		return "runtime.Opt[" + t.optElem().goType() + "]"
	case t.isClass():
		return string(t)
	case t.isClassRef():
		return "*runtime.ClassRef"
	case t.isGen():
		a := t.typeArgs()
		return "*runtime.Gen[" + a[0].goType() + ", " + a[1].goType() + ", " + a[2].goType() + "]"
	case t == tRange:
		return "*runtime.Range"
	case t.isFunc():
		params := t.funcParams()
		parts := make([]string, len(params))
		for i, p := range params {
			parts[i] = p.goType()
		}
		out := "func(" + strings.Join(parts, ", ") + ")"
		if r := t.funcRet(); r != tNone {
			out += " " + r.goType()
		}
		return out
	case t.isVarTuple():
		return "runtime.VarTuple[" + t.elem().goType() + "]"
	case t.isIterable():
		return "runtime.Iterable[" + t.elem().goType() + "]"
	case t.isList():
		return "*runtime.List[" + t.elem().goType() + "]"
	case t.isSet():
		return "*runtime.Set[" + t.elem().goType() + "]"
	case t.isIter():
		return "*runtime.Iterator[" + t.elem().goType() + "]"
	case t.isDict():
		return "*runtime.Dict[" + t.keyType().goType() + ", " + t.valType().goType() + "]"
	case t.isKeysView():
		return "*runtime.KeysView[" + t.keyType().goType() + ", " + t.valType().goType() + "]"
	case t.isValuesView():
		return "*runtime.ValuesView[" + t.keyType().goType() + ", " + t.valType().goType() + "]"
	case t.isItemsView():
		return "*runtime.ItemsView[" + t.keyType().goType() + ", " + t.valType().goType() + "]"
	case t.isTuple():
		elems := t.elems()
		parts := make([]string, len(elems))
		for i, e := range elems {
			parts[i] = e.goType()
		}
		return fmt.Sprintf("runtime.Tuple%d[%s]", len(elems), strings.Join(parts, ", "))
	}
	return ""
}

// typeTable maps annotation names to the scalar types. Anything not listed
// here (a class name, ...) is unsupported.
var typeTable = map[string]pyType{
	"int":   tInt,
	"float": tFloat,
	"bool":  tBool,
	"str":   tStr,
	"range": tRange,
	// Types without a Go type of their own are objects that the runtime
	// knows how to operate on.
	"complex":    tObject,
	"bytes":      tObject,
	"bytearray":  tObject,
	"memoryview": tObject,
	"frozenset":  tObject,
	"slice":      tObject,
}

// annotationType resolves a type annotation. allowNone permits the literal
// None, which is valid as a return annotation only.
func annotationType(n pyast.Node, allowNone bool) (pyType, error) {
	return (&generator{}).annotationType(n, allowNone)
}

// annotationType resolves an annotation, which can name a class the program
// defines. A class used before its definition is written as a string.
func (g *generator) annotationType(n pyast.Node, allowNone bool) (pyType, error) {
	n = typingFromABC(n)
	switch n := n.(type) {
	case *pyast.Name:
		if t, ok := g.tsubst[n.ID]; ok {
			return t, nil
		}
		if _, free := g.typeVars[n.ID]; free {
			return "", unsupported("type variable '"+n.ID+"' used outside of a generic definition", n.Line())
		}
		if al := g.typeAliases[n.ID]; al != nil {
			if len(al.params) > 0 {
				return "", unsupported("generic type alias '"+n.ID+"' needs type arguments", n.Line())
			}
			var t pyType
			err := g.withSubst(nil, func() (e error) { t, e = g.annotationType(al.value, allowNone); return })
			return t, err
		}
		if gc := g.genericClasses[n.ID]; gc != nil {
			return "", unsupported("generic class '"+n.ID+"' needs type arguments: "+n.ID+"[...]", n.Line())
		}
		if t, ok := typeTable[n.ID]; ok {
			return t, nil
		}
		if n.ID == "object" || (n.ID == "Any" && g.classes["Any"] == nil) {
			return tObject, nil
		}
		if _, user := g.classes[n.ID]; !user && runtime.IsExceptionClass(n.ID) {
			return tExc, nil
		}
		if _, ok := g.classes[n.ID]; ok {
			return pyType(n.ID), nil
		}
		if why, ok := typingNames[n.ID]; ok {
			return "", unsupported("type annotation '"+n.ID+"' ("+why+")", n.Line())
		}
		return "", unsupported("type annotation '"+n.ID+"'", n.Line())
	case *pyast.Subscript:
		return g.containerType(n)
	case *pyast.BinOp:
		if n.Op == "BitOr" {
			return g.unionType([]pyast.Node{n.Left, n.Right}, n.Line())
		}
		return "", unsupported("type annotation of kind BinOp", n.Line())
	case *pyast.Constant:
		if n.PyType == "none" && allowNone {
			return tNone, nil
		}
		if s, ok := n.Value.(string); ok && n.PyType == "str" {
			// A forward reference: "Node", or "list[Node] | None".
			if _, ok := g.classes[s]; ok && g.tsubst[s] == "" {
				return pyType(s), nil
			}
			parsed, err := parseAnnotation(s, n.Line())
			if err != nil {
				return "", err
			}
			return g.annotationType(parsed, allowNone)
		}
		return "", unsupported("type annotation "+constantText(n), n.Line())
	case *pyast.Unsupported:
		return "", unsupported("type annotation of kind "+n.Type, n.Line())
	default:
		return "", unsupported("type annotation of kind "+nodeName(n), n.Line())
	}
}

// typingAliases are the names of the typing module that mean a built-in
// container: List[int] is list[int]. The typing module itself arrives in
// phase 12; until then these names are recognised on their own.
var typingAliases = map[string]string{
	"List": "list", "MutableSequence": "list",
	"Set":  "set",
	"Dict": "dict", "Mapping": "dict", "MutableMapping": "dict",
	"Tuple": "tuple",
}

// typingNames are typing names that exist, but need something the typed
// subset does not have yet.
var typingNames = map[string]string{
	"Literal":  "needs literal types",
	"Callable": "write the parameters: Callable[[int], int]",
	"Sequence": "use list[...]; sequences are not a type of their own",
	"Iterable": "write the item type: Iterable[int]",
	"Generic":  "write the type parameters in brackets: class C[T]:",
	"TypeVar":  "write the type parameters in brackets: def f[T](x: T) -> T:",
	"List":     "write the item type: list[int]",
	"Dict":     "write the key and value types: dict[str, int]",
	"Set":      "write the item type: set[int]",
	"Tuple":    "write the item types: tuple[int, str]",
	"Type":     "write type[C]",
}

// containerType resolves list[T], set[T], dict[K, V], tuple[A, B, ...],
// tuple[T, ...], Iterable[T] and Callable[[A, B], R].
func (g *generator) containerType(n *pyast.Subscript) (pyType, error) {
	if t, handled, err := g.genericSubscript(n); handled {
		return t, err
	}
	head, ok := n.Value.(*pyast.Name)
	if !ok {
		return "", unsupported("type annotation of kind Subscript", n.Line())
	}
	args := []pyast.Node{n.Index}
	if t, ok := n.Index.(*pyast.Tuple); ok {
		args = t.Elts
	}
	resolve := func(node pyast.Node) (pyType, error) { return g.annotationType(node, false) }

	name := head.ID
	if alias, ok := typingAliases[name]; ok {
		name = alias
	}
	switch name {
	case "Callable":
		return g.callableType(n, args)
	case "Iterator":
		if len(args) != 1 {
			return "", unsupported(fmt.Sprintf("Iterator[...] with %d type arguments", len(args)), n.Line())
		}
		elem, err := resolve(args[0])
		if err != nil {
			return "", err
		}
		return iterOf(elem), nil
	case "Generator":
		if len(args) < 1 || len(args) > 3 {
			return "", unsupported(fmt.Sprintf("Generator[...] with %d type arguments", len(args)), n.Line())
		}
		parts := []pyType{"", tNoneType, tNoneType}
		for i, a := range args {
			if c, ok := a.(*pyast.Constant); ok && c.PyType == "none" {
				parts[i] = tNoneType
				continue
			}
			t, err := resolve(a)
			if err != nil {
				return "", err
			}
			parts[i] = t
		}
		return genOf(parts[0], parts[1], parts[2]), nil
	case "Iterable":
		if len(args) != 1 {
			return "", unsupported(fmt.Sprintf("Iterable[...] with %d type arguments", len(args)), n.Line())
		}
		elem, err := resolve(args[0])
		if err != nil {
			return "", err
		}
		return iterableOfType(elem), nil
	case "ClassVar":
		return "", unsupported("ClassVar outside a class body", n.Line())
	case "Optional":
		if len(args) != 1 {
			return "", unsupported(fmt.Sprintf("Optional[...] with %d type arguments", len(args)), n.Line())
		}
		return g.unionType(append(args, &pyast.Constant{PyType: "none", Lineno: n.Line()}), n.Line())
	case "Union":
		return g.unionType(args, n.Line())
	case "Literal":
		// The type of the values written in the brackets, or object.
		var lit pyType
		for _, a := range args {
			c, ok := a.(*pyast.Constant)
			if !ok {
				return tObject, nil
			}
			var ct pyType
			switch c.PyType {
			case "int":
				ct = tInt
			case "str":
				ct = tStr
			case "bool":
				ct = tBool
			default:
				return tObject, nil
			}
			if lit != "" && lit != ct {
				return tObject, nil
			}
			lit = ct
		}
		if lit == "" {
			return tObject, nil
		}
		return lit, nil
	case "Type", "Sequence":
		return "", unsupported("type annotation '"+head.ID+"[...]' ("+typingNames[head.ID]+")", n.Line())
	case "frozenset":
		// A frozenset is an object whatever it holds.
		return tObject, nil
	}

	switch name {
	case "list", "set":
		if len(args) != 1 {
			return "", unsupported(fmt.Sprintf("%s[...] with %d type arguments", head.ID, len(args)), n.Line())
		}
		elem, err := resolve(args[0])
		if err != nil {
			return "", err
		}
		if name == "list" {
			return listOf(elem), nil
		}
		if !g.hashable(elem) {
			return "", unsupported(fmt.Sprintf("set element type %s (not hashable)", elem), n.Line())
		}
		return setOf(g.keyType(elem)), nil

	case "dict":
		if len(args) != 2 {
			return "", unsupported(fmt.Sprintf("dict[...] with %d type arguments", len(args)), n.Line())
		}
		key, err := resolve(args[0])
		if err != nil {
			return "", err
		}
		val, err := resolve(args[1])
		if err != nil {
			return "", err
		}
		if !g.hashable(key) {
			return "", unsupported(fmt.Sprintf("dict key type %s (not hashable)", key), n.Line())
		}
		return dictOf(g.keyType(key), val), nil

	case "tuple":
		if len(args) == 2 {
			if isEllipsis(args[1]) {
				elem, err := resolve(args[0])
				if err != nil {
					return "", err
				}
				return varTupleOf(elem), nil
			}
		}
		if len(args) > maxTupleArity {
			return "", unsupported(fmt.Sprintf("tuple with more than %d elements", maxTupleArity), n.Line())
		}
		elems := make([]pyType, 0, len(args))
		for _, item := range args {
			if isEllipsis(item) {
				return "", unsupported("tuple annotation with '...' after more than one item", n.Line())
			}
			t, err := resolve(item)
			if err != nil {
				return "", err
			}
			elems = append(elems, t)
		}
		if len(elems) == 0 {
			return "", unsupported("empty tuple annotation", n.Line())
		}
		return tupleOf(elems), nil
	}
	return "", unsupported("type annotation '"+head.ID+"[...]'", n.Line())
}

// constantText describes a literal for error messages.
func constantText(c *pyast.Constant) string {
	if c.PyType == "none" {
		return "None"
	}
	if s, ok := c.Value.(string); ok {
		return strconv.Quote(s)
	}
	return fmt.Sprint(c.Value)
}

// callableType resolves Callable[[A, B], R].
func (g *generator) callableType(n *pyast.Subscript, args []pyast.Node) (pyType, error) {
	if len(args) != 2 {
		return "", unsupported("Callable[...] needs a parameter list and a return type", n.Line())
	}
	list, ok := args[0].(*pyast.List)
	if !ok {
		return "", unsupported("Callable[..., R] (the parameters must be written out)", n.Line())
	}
	params := make([]pyType, len(list.Elts))
	for i, p := range list.Elts {
		t, err := g.annotationType(p, false)
		if err != nil {
			return "", err
		}
		params[i] = t
	}
	ret, err := g.annotationType(args[1], true)
	if err != nil {
		return "", err
	}
	return funcOf(params, ret), nil
}

// unionType resolves Optional[T], T | None and Union[T, None]. A union of
// several types needs the dynamic object model of phase 9.
func (g *generator) unionType(parts []pyast.Node, line int) (pyType, error) {
	var members []pyType
	hasNone := false
	var flatten func(n pyast.Node) error
	flatten = func(n pyast.Node) error {
		if b, ok := n.(*pyast.BinOp); ok && b.Op == "BitOr" {
			if err := flatten(b.Left); err != nil {
				return err
			}
			return flatten(b.Right)
		}
		if c, ok := n.(*pyast.Constant); ok && c.PyType == "none" {
			hasNone = true
			return nil
		}
		t, err := g.annotationType(n, false)
		if err != nil {
			return err
		}
		members = append(members, t)
		return nil
	}
	for _, p := range parts {
		if err := flatten(p); err != nil {
			return "", err
		}
	}
	switch {
	case len(members) == 0:
		return "", unsupported("the type None where a type is needed", line)
	case len(members) > 1:
		// A value of one of several types is an object whose type is found
		// when the program runs.
		return tObject, nil
	case !hasNone:
		return members[0], nil
	}
	return optionalOf(members[0]), nil
}

// isEllipsis reports whether a node is the literal "...".
func isEllipsis(n pyast.Node) bool {
	switch n := n.(type) {
	case *pyast.Constant:
		return n.PyType == "ellipsis"
	case *pyast.Unsupported:
		return n.Type == "ellipsis constant"
	}
	return false
}

// isScalar reports whether t is int, float, str or bool.
func (t pyType) isScalar() bool { return t == tInt || t == tFloat || t == tStr || t == tBool }

// abcPrefix is how the linker names the classes of collections.abc. In an
// annotation they stand for the typing names they share (Iterable[int]).
const abcPrefix = "collections__abc_pm_"

func typingFromABC(n pyast.Node) pyast.Node {
	switch x := n.(type) {
	case *pyast.Name:
		if rest, ok := strings.CutPrefix(x.ID, abcPrefix); ok {
			return &pyast.Name{ID: rest, Lineno: x.Lineno}
		}
	case *pyast.Subscript:
		if head, ok := x.Value.(*pyast.Name); ok {
			if rest, ok := strings.CutPrefix(head.ID, abcPrefix); ok {
				return &pyast.Subscript{Value: &pyast.Name{ID: rest, Lineno: head.Lineno}, Index: x.Index, Lineno: x.Lineno}
			}
		}
	}
	return n
}

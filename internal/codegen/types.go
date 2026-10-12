package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
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

func (t pyType) isList() bool       { return t.head() == "list" }
func (t pyType) isTuple() bool      { return t.head() == "tuple" }
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
	case t.isList(), t.isSet(), t.isIter():
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
	case t == tInt, t == tBool, t == tStr:
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
	case t == tExc:
		return "*runtime.PyError"
	case t == tRange:
		return "*runtime.Range"
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
}

// annotationType resolves a type annotation. allowNone permits the literal
// None, which is valid as a return annotation only.
func annotationType(n pyast.Node, allowNone bool) (pyType, error) {
	switch n := n.(type) {
	case *pyast.Name:
		if t, ok := typeTable[n.ID]; ok {
			return t, nil
		}
		return "", unsupported("type annotation '"+n.ID+"'", n.Line())
	case *pyast.Subscript:
		return containerType(n)
	case *pyast.Constant:
		if n.PyType == "none" && allowNone {
			return tNone, nil
		}
		return "", unsupported("type annotation "+constantText(n), n.Line())
	case *pyast.Unsupported:
		return "", unsupported("type annotation of kind "+n.Type, n.Line())
	default:
		return "", unsupported("type annotation of kind "+nodeName(n), n.Line())
	}
}

// containerType resolves list[T], set[T], dict[K, V] and tuple[A, B, ...].
func containerType(n *pyast.Subscript) (pyType, error) {
	head, ok := n.Value.(*pyast.Name)
	if !ok {
		return "", unsupported("type annotation of kind Subscript", n.Line())
	}
	args := []pyast.Node{n.Index}
	if t, ok := n.Index.(*pyast.Tuple); ok {
		args = t.Elts
	}
	resolve := func(node pyast.Node) (pyType, error) { return annotationType(node, false) }

	switch head.ID {
	case "list", "set":
		if len(args) != 1 {
			return "", unsupported(fmt.Sprintf("%s[...] with %d type arguments", head.ID, len(args)), n.Line())
		}
		elem, err := resolve(args[0])
		if err != nil {
			return "", err
		}
		if head.ID == "list" {
			return listOf(elem), nil
		}
		if !elem.hashable() {
			return "", unsupported(fmt.Sprintf("set element type %s (not hashable)", elem), n.Line())
		}
		return setOf(elem), nil

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
		if !key.hashable() {
			return "", unsupported(fmt.Sprintf("dict key type %s (not hashable)", key), n.Line())
		}
		return dictOf(key, val), nil

	case "tuple":
		if len(args) > maxTupleArity {
			return "", unsupported(fmt.Sprintf("tuple with more than %d elements", maxTupleArity), n.Line())
		}
		elems := make([]pyType, 0, len(args))
		for _, item := range args {
			// "..." decodes as an Unsupported "ellipsis constant" node.
			if u, ok := item.(*pyast.Unsupported); ok && u.Type == "ellipsis constant" {
				return "", unsupported("variable-length tuple annotation", n.Line())
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

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
	tNone  pyType = "None" // only the result of a call to a function that returns None

	// tExc is the type of the name bound by "except ... as e". It cannot be
	// written as an annotation, so an exception never outlives its handler.
	tExc pyType = "exception"
)

// maxTupleArity is the largest tuple the runtime has a struct for.
const maxTupleArity = 6

func (t pyType) String() string { return string(t) }

func listOf(elem pyType) pyType { return "list[" + elem + "]" }

func tupleOf(elems []pyType) pyType {
	parts := make([]string, len(elems))
	for i, e := range elems {
		parts[i] = string(e)
	}
	return pyType("tuple[" + strings.Join(parts, ", ") + "]")
}

func (t pyType) isList() bool  { return strings.HasPrefix(string(t), "list[") }
func (t pyType) isTuple() bool { return strings.HasPrefix(string(t), "tuple[") }

// elem is the element type of a list.
func (t pyType) elem() pyType {
	return pyType(strings.TrimSuffix(strings.TrimPrefix(string(t), "list["), "]"))
}

// elems are the element types of a tuple, in order.
func (t pyType) elems() []pyType {
	inner := strings.TrimSuffix(strings.TrimPrefix(string(t), "tuple["), "]")
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

// goType is the Go spelling of the type. Lists are pointers because Python
// lists alias, and tuples are the runtime's generic structs.
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
	case t.isList():
		return "*runtime.List[" + t.elem().goType() + "]"
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

// containerType resolves list[T] and tuple[A, B, ...].
func containerType(n *pyast.Subscript) (pyType, error) {
	head, ok := n.Value.(*pyast.Name)
	if !ok {
		return "", unsupported("type annotation of kind Subscript", n.Line())
	}
	switch head.ID {
	case "list":
		elem, err := annotationType(n.Index, false)
		if err != nil {
			return "", err
		}
		return listOf(elem), nil
	case "tuple":
		items := []pyast.Node{n.Index}
		if t, ok := n.Index.(*pyast.Tuple); ok {
			items = t.Elts
		}
		if len(items) > maxTupleArity {
			return "", unsupported(fmt.Sprintf("tuple with more than %d elements", maxTupleArity), n.Line())
		}
		elems := make([]pyType, 0, len(items))
		for _, item := range items {
			// "..." decodes as an Unsupported "ellipsis constant" node.
			if u, ok := item.(*pyast.Unsupported); ok && u.Type == "ellipsis constant" {
				return "", unsupported("variable-length tuple annotation", n.Line())
			}
			t, err := annotationType(item, false)
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

package codegen

import (
	"fmt"
	"strconv"

	"github.com/amayakmt/py2go/internal/pyast"
)

// pyType is the static type of an expression or variable.
type pyType int

const (
	tInt pyType = iota + 1
	tFloat
	tBool
	tStr
	tNone // only the result of a call to a function that returns None
)

func (t pyType) String() string {
	switch t {
	case tInt:
		return "int"
	case tFloat:
		return "float"
	case tBool:
		return "bool"
	case tStr:
		return "str"
	case tNone:
		return "None"
	}
	return "invalid"
}

// goType is the Go spelling of the type.
func (t pyType) goType() string {
	switch t {
	case tInt:
		return "int64"
	case tFloat:
		return "float64"
	case tBool:
		return "bool"
	case tStr:
		return "string"
	}
	return ""
}

// typeTable maps annotation names to types. Anything not listed here
// (list, dict, a class name, ...) is unsupported.
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
		return 0, unsupported("type annotation '"+n.ID+"'", n.Line())
	case *pyast.Constant:
		if n.PyType == "none" && allowNone {
			return tNone, nil
		}
		return 0, unsupported("type annotation "+constantText(n), n.Line())
	case *pyast.Unsupported:
		return 0, unsupported("type annotation of kind "+n.Type, n.Line())
	default:
		return 0, unsupported("type annotation of kind "+nodeName(n), n.Line())
	}
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

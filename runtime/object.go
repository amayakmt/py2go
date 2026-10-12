package runtime

import (
	"fmt"
	"reflect"
)

// Instances of user-defined classes are Go pointers. The generated code makes
// each class implement the hooks below, and the runtime reaches user code
// through them, so print, ==, sorting and hash() work on instances.

// Named is implemented by every instance of a user-defined class.
type Named interface{ PyClassName() string }

// Reprer and Strer are implemented by instances: PyRepr is repr(x) and PyStr
// is str(x), which falls back to PyRepr when the class has no __str__.
type Reprer interface{ PyRepr() string }
type Strer interface{ PyStr() string }

// Equaler is __eq__. The result is false in the second position when the
// method returned NotImplemented, or when it takes another type than other.
type Equaler interface {
	PyEq(other any) (result, implemented bool)
}

// Orderer is the four ordering methods. op is "lt", "le", "gt" or "ge".
type Orderer interface {
	PyOrder(op string, other any) (result, implemented bool)
}

// Hasher is __hash__.
type Hasher interface{ PyHash() int64 }

// TypeName is the name Python shows for the type of v in error messages.
func TypeName(v any) string {
	switch v := v.(type) {
	case nil:
		return "NoneType"
	case Named:
		return v.PyClassName()
	case bool:
		return "bool"
	case int64, int:
		return "int"
	case float64:
		return "float"
	case string:
		return "str"
	case *PyError:
		return v.Type
	case Seq:
		return v.Kind()
	}
	t := reflect.TypeOf(v)
	switch {
	case t.Kind() == reflect.Func:
		return "function"
	case t.Kind() == reflect.Pointer:
		name := t.Elem().Name()
		switch {
		case len(name) >= 4 && name[:4] == "Dict":
			return "dict"
		case len(name) >= 3 && name[:3] == "Set":
			return "set"
		case len(name) >= 5 && name[:5] == "Range":
			return "range"
		}
	}
	return t.String()
}

// InstanceOf is isinstance(v, C) for a user-defined class or one of the Go
// types that stand for the built-in ones. T is the generated interface of the
// class, whose marker method makes the test nominal.
func InstanceOf[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

// Is is "a is b" for two instances: the same object.
func Is(a, b any) bool { return a == b }

// UnsupportedOperand is the TypeError of a binary operator no operand
// implements.
func UnsupportedOperand(op, left, right string) *PyError {
	return newError("TypeError", fmt.Sprintf("unsupported operand type(s) for %s: '%s' and '%s'", op, left, right))
}

// UnsupportedOperandUnary is the TypeError of a unary operator.
func UnsupportedOperandUnary(op, operand string) *PyError {
	return newError("TypeError", fmt.Sprintf("bad operand type for unary %s: '%s'", op, operand))
}

// NotOrderable is "'<' not supported between instances of 'A' and 'B'".
func NotOrderable(op string, a, b any) *PyError {
	return newError("TypeError", fmt.Sprintf("'%s' not supported between instances of '%s' and '%s'", op, TypeName(a), TypeName(b)))
}

// NoAttribute is the AttributeError of reading an attribute that is not there.
func NoAttribute(class, attr string) *PyError {
	return newError("AttributeError", fmt.Sprintf("'%s' object has no attribute '%s'", class, attr))
}

// NoSetter is the AttributeError of assigning to a read-only property.
func NoSetter(class, attr string) *PyError {
	return newError("AttributeError", fmt.Sprintf("property '%s' of '%s' object has no setter", attr, class))
}

// NoDeleter is the AttributeError of deleting a property.
func NoDeleter(class, attr string) *PyError {
	return newError("AttributeError", fmt.Sprintf("property '%s' of '%s' object has no deleter", attr, class))
}

// FrozenInstance is the error of assigning to a field of a frozen dataclass.
func FrozenInstance(attr string) *PyError {
	return newError("FrozenInstanceError", fmt.Sprintf("cannot assign to field '%s'", attr))
}

// NotIterable and friends raise the TypeError of a protocol the type lacks.
func NotCallable(class string) *PyError {
	return newError("TypeError", fmt.Sprintf("'%s' object is not callable", class))
}

// HashOf is hash(v) for an instance; instances without __hash__ hash by
// identity, which CPython derives from the address.
func HashOf(v any) int64 {
	if h, ok := v.(Hasher); ok {
		return h.PyHash()
	}
	return Hash(v)
}

// UnhashableType is the TypeError of hash() on an instance of a class that
// defines __eq__ and not __hash__.
func UnhashableType(class string) *PyError {
	return newError("TypeError", fmt.Sprintf("unhashable type: '%s'", class))
}

// DefaultRepr is the repr of an object without __repr__:
// <__main__.C object at 0x...>. The address cannot match CPython's.
func DefaultRepr(class string, self any) string {
	return fmt.Sprintf("<__main__.%s object at %p>", class, self)
}

// IdentityHash is the default hash of an instance.
func IdentityHash(self any) int64 {
	v := reflect.ValueOf(self)
	return int64(v.Pointer()>>4) & 0x1fffffffffffffff
}

// AnySeq is a tuple of values of different types, which generated code
// builds to compare or hash the fields of a dataclass.
type AnySeq []any

func (s AnySeq) Len() int64     { return int64(len(s)) }
func (s AnySeq) At(i int64) any { return s[i] }
func (s AnySeq) Kind() string   { return "tuple" }

// ExcTypeName is type(e).__name__ for a caught exception, or "NoneType" for
// the nil that e.__cause__ can be.
func ExcTypeName(e *PyError) string {
	if e == nil {
		return "NoneType"
	}
	return e.Type
}

// IsExcInstance is isinstance(x, E) for a value held as object.
func IsExcInstance(x any, class string) bool {
	if e, ok := x.(interface{ PyExc() *PyError }); ok {
		return IsSubclass(e.PyExc().Type, class)
	}
	return false
}

// RawArgs are the positional arguments an exception class was called with,
// which BaseException keeps in args whatever __init__ does with them. given
// is how many the call wrote, or -1 when it is not known and all count.
func RawArgs(given int, all ...any) []any {
	if given < 0 || given > len(all) {
		return all
	}
	return all[:given]
}

// AsInt reads an object that isinstance found to be an int: a bool is an int
// in Python.
func AsInt(x any) int64 {
	if b, ok := x.(bool); ok {
		return b2i(b)
	}
	return x.(int64)
}

// ClassRef is a class used as a value, such as the result of type(x). Each
// class has one, so two of them are the same class when they are equal.
type ClassRef struct {
	Name    string
	Builtin bool
}

var classRefs = map[string]*ClassRef{}

// ClassOf is the class object of a user-defined class.
func ClassOf(name string) *ClassRef {
	if c, ok := classRefs[name]; ok {
		return c
	}
	c := &ClassRef{Name: name}
	classRefs[name] = c
	return c
}

// BuiltinClassOf is the class object of int, str and the other built-in types.
func BuiltinClassOf(name string) *ClassRef {
	if c, ok := classRefs["builtins."+name]; ok {
		return c
	}
	c := &ClassRef{Name: name, Builtin: true}
	classRefs["builtins."+name] = c
	return c
}

func (c *ClassRef) PyRepr() string {
	if c.Builtin {
		return "<class '" + c.Name + "'>"
	}
	return "<class '__main__." + c.Name + "'>"
}

// ClassOfValue is type(x) for an instance.
func ClassOfValue(x Named) *ClassRef { return ClassOf(x.PyClassName()) }

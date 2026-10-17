package runtime

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/big"
	"reflect"
	"strings"
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
	case nil, NoneT:
		return "NoneType"
	case Named:
		return BaseTypeName(v.PyClassName())
	case bool:
		return "bool"
	case int64, int, *big.Int:
		return "int"
	case float64:
		return "float"
	case complex128:
		return "complex"
	case string:
		return "str"
	case *PyError:
		return BaseTypeName(v.Type)
	case Seq:
		return v.Kind()
	case interface{ Boxed() any }:
		return TypeName(v.Boxed()) // an Optional is the type of what it holds
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
		case strings.HasPrefix(name, "Iterator"), strings.HasPrefix(name, "Gen["):
			return "generator"
		case name == "ClassRef":
			return "type"
		case name == "Callable":
			return "function"
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
func Is(a, b any) bool {
	a, b = Unbox(a), Unbox(b)
	switch x := a.(type) {
	case nil:
		return b == nil
	case float64:
		y, ok := b.(float64)
		return ok && math.Float64bits(x) == math.Float64bits(y)
	case int64, bool, string, complex128, Bytes:
		// Equal numbers and strings are the same object where CPython shares
		// them (small ints, interned strings, constants of one function); the
		// answer for values that were computed separately is unknowable here.
		return a == b
	case *big.Int:
		y, ok := b.(*big.Int)
		return ok && x.Cmp(y) == 0
	case Seq:
		if x.Kind() == "tuple" {
			y, ok := b.(Seq)
			return ok && y.Kind() == "tuple" && Eq(a, b)
		}
	}
	if r, isFunc := sameFunc(a, b); isFunc {
		return r
	}
	defer func() { recover() }()
	return a == b
}

// UnsupportedOperand is the TypeError of a binary operator no operand
// implements.
func UnsupportedOperand(op, left, right string) *PyError {
	return newError("TypeError", fmt.Sprintf("unsupported operand type(s) for %s: '%s' and '%s'", op, errTypeName(left), errTypeName(right)))
}

// UnsupportedOperandUnary is the TypeError of a unary operator.
func UnsupportedOperandUnary(op, operand string) *PyError {
	return newError("TypeError", fmt.Sprintf("bad operand type for unary %s: '%s'", op, errTypeName(operand)))
}

// NotOrderable is "'<' not supported between instances of 'A' and 'B'".
func NotOrderable(op string, a, b any) *PyError {
	return newError("TypeError", fmt.Sprintf("'%s' not supported between instances of '%s' and '%s'", op, errTypeName(TypeName(a)), errTypeName(TypeName(b))))
}

// errTypeName is a type name as the messages of C types spell it: the types
// of the datetime and decimal modules carry the module name.
func errTypeName(name string) string {
	if c, ok := classRefs[name]; ok && (c.Module == "datetime" || c.Module == "decimal") {
		return c.Module + "." + name
	}
	return name
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
		return fixHash(h.PyHash())
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
	return fmt.Sprintf("<%s.%s object at %p>", ClassOf(class).module(), BaseTypeName(class), self)
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
	return BaseTypeName(e.Type)
}

// IsExcInstance is isinstance(x, E) for a value held as object.
func IsExcInstance(x any, class string) bool {
	if e, ok := Unbox(x).(*PyError); ok {
		return IsSubclass(e.Type, class)
	}
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
	switch v := x.(type) {
	case bool:
		return b2i(v)
	case *big.Int:
		panic(overflowError("Python int too large to convert to C long"))
	}
	return x.(int64)
}

// ClassRef is a class used as a value, such as the result of type(x). Each
// class has one, so two of them are the same class when they are equal.
type ClassRef struct {
	Name    string
	Builtin bool
	Module  string    // the module of a class defined in a module of the project
	ctor    any       // the constructor of a class of the program
	info    *FuncInfo // and the names of its parameters
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

// module is the module a class was defined in.
func (c *ClassRef) module() string {
	if c.Module != "" {
		return c.Module
	}
	return "__main__"
}

// ClassModule records the module a class of a joined module was defined in.
func ClassModule(name, module string) { ClassOf(name).Module = module }

func (c *ClassRef) PyRepr() string {
	if c.Builtin {
		return "<class '" + c.Name + "'>"
	}
	return "<class '" + c.module() + "." + BaseTypeName(c.Name) + "'>"
}

// ClassOfValue is type(x) for an instance.
func ClassOfValue(x Named) *ClassRef { return ClassOf(x.PyClassName()) }

// ExcClass is the class of a caught exception, as an object.
func ExcClass(e *PyError) *ClassRef {
	if _, user := userMRO[e.Type]; user {
		return ClassOf(e.Type)
	}
	return BuiltinClassOf(e.Type)
}

// ExcValue is the exception as the value an __exit__ method receives: the
// instance for a user-defined class, the exception itself otherwise.
func ExcValue(e *PyError) any {
	if e.Obj != nil {
		return e.Obj
	}
	return e
}

// ExitCall runs the __exit__ of a with statement whose body raised pending.
// An exception raised by __exit__ gets pending as its context.
func ExitCall(pending *PyError, exit func() bool) (suppress bool) {
	if e := Catch(func() { suppress = exit() }); e != nil {
		panic(WithContext(e, pending))
	}
	return suppress
}

// StopValue is e.value of a StopIteration: what the generator returned.
func StopValue(e *PyError) any {
	if e.Type != "StopIteration" {
		panic(newError("AttributeError", "'"+e.Type+"' object has no attribute 'value'"))
	}
	if len(e.Args) == 0 {
		return nil
	}
	return e.Args[0]
}

// Id is id(x): the address of an object that has an identity, and a number
// that depends on the value alone for the ones that compare as values here
// (ints, strings, None, tuples...), so that id(1) == id(1) as it does in
// CPython for small ints and constants.
func Id(v any) int64 {
	v = Unbox(v)
	rv := reflect.ValueOf(v)
	if v != nil {
		switch rv.Kind() {
		case reflect.Pointer, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
			return int64(rv.Pointer())
		}
	}
	h := fnv.New64a()
	fmt.Fprintf(h, "%T|%s", v, Repr(v))
	return int64(h.Sum64()>>8) | 1<<55
}

var (
	classAttrs = map[string]map[string]func() any{}
	classBases = map[string][]string{}
)

// RegisterClassAttr records how to read a class attribute of a class of the
// program, for getattr(cls, name).
func RegisterClassAttr(class, name string, get func() any) {
	m := classAttrs[class]
	if m == nil {
		m = map[string]func() any{}
		classAttrs[class] = m
	}
	m[name] = get
}

// RegisterClassBases records a class and its bases of the program, nearest
// first, so that an inherited class attribute is found.
func RegisterClassBases(chain ...string) { classBases[chain[0]] = chain }

var classMembers = map[string][]string{}

var classCallables = map[string]map[string]func() any{}

// RegisterClassCallable records how to read a classmethod or staticmethod of
// a class of the program as a value of that class.
func RegisterClassCallable(class, name string, get func() any) {
	m := classCallables[class]
	if m == nil {
		m = map[string]func() any{}
		classCallables[class] = m
	}
	m[name] = get
}

// RegisterClassMembers records the names of the methods, properties and
// class attributes a class of the program has, inherited ones included, for
// dir(cls) and getattr(cls, name).
func RegisterClassMembers(class string, names ...string) { classMembers[class] = names }

// classHasMember tells whether a class of the program has a method or
// attribute of that name.
func classHasMember(class, name string) bool {
	for _, n := range classMembers[class] {
		if n == name {
			return true
		}
	}
	return false
}

func lookupClassAttr(class, name string) (any, bool) {
	chain := classBases[class]
	if chain == nil {
		chain = []string{class}
	}
	for _, k := range chain {
		if get, ok := classAttrs[k][name]; ok {
			return get(), true
		}
	}
	return nil, false
}

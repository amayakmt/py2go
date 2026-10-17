package runtime

import "fmt"

// Opt is a value that can be None: Optional[T] and T | None.
type Opt[T any] struct {
	v    T
	some bool
}

// Some wraps a value.
func Some[T any](v T) Opt[T] { return Opt[T]{v: v, some: true} }

// NoneOf is None where an Optional[T] is expected.
func NoneOf[T any]() Opt[T] { return Opt[T]{} }

// SomeAny is an object as an Optional[object]: None when the object is None.
func SomeAny(v any) Opt[any] {
	if v == nil {
		return Opt[any]{}
	}
	if _, none := v.(NoneT); none {
		return Opt[any]{}
	}
	return Opt[any]{v: v, some: true}
}

func (o Opt[T]) IsSome() bool { return o.some }

// Get is the value, for code that has checked the Opt is not None.
func (o Opt[T]) Get() T {
	if !o.some {
		panic(newError("TypeError", "NoneType used where a value is required"))
	}
	return o.v
}

// Unwrap is the value, or the error Python raises when something that is
// None is used: errType and msg say which.
func (o Opt[T]) Unwrap(errType, msg string) T {
	if !o.some {
		panic(newError(errType, msg))
	}
	return o.v
}

// UnwrapAttr is x.attr where x is None.
func (o Opt[T]) UnwrapAttr(attr string) T {
	if !o.some {
		panic(newError("AttributeError", fmt.Sprintf("'NoneType' object has no attribute '%s'", attr)))
	}
	return o.v
}

// OrNone is the value as any, with nil for None.
func (o Opt[T]) Boxed() any {
	if !o.some {
		return nil
	}
	return o.v
}

func (o Opt[T]) PyRepr() string {
	if !o.some {
		return "None"
	}
	return Repr(o.v)
}

func (o Opt[T]) PyStr() string {
	if !o.some {
		return "None"
	}
	return Str(o.v)
}

// PyEq compares with another Optional or with a plain value.
func (o Opt[T]) PyEq(other any) (bool, bool) {
	switch x := other.(type) {
	case Opt[T]:
		if !o.some || !x.some {
			return o.some == x.some, true
		}
		return Eq(o.v, x.v), true
	case NoneT:
		return !o.some, true
	case nil:
		return !o.some, true
	}
	if !o.some {
		return false, true
	}
	return Eq(o.v, other), true
}

// PyHash hashes None like CPython (a constant), and a value like itself.
func (o Opt[T]) PyHash() int64 {
	if !o.some {
		return 0xFCA86420
	}
	return Hash(o.v)
}

// OptTruth is the truth of an Optional: None is false.
func OptTruth[T any](o Opt[T]) bool { return o.some && Truth(o.v) }

// NoneT is the type of the value None where no Optional is expected, such as
// print(None).
type NoneT struct{}

// NoneV is None.
var NoneV = NoneT{}

func (NoneT) PyRepr() string { return "None" }
func (NoneT) PyStr() string  { return "None" }
func (NoneT) PyEq(other any) (bool, bool) {
	switch x := other.(type) {
	case NoneT:
		return true, true
	case interface{ IsSomeAny() bool }:
		return !x.IsSomeAny(), true
	}
	return false, true
}

func (o Opt[T]) IsSomeAny() bool { return o.some }

func (o Opt[T]) pyTruth() bool { return o.some && Truth(o.v) }
func (NoneT) pyTruth() bool    { return false }

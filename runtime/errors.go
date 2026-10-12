// Package runtime is the support library for generated Go code.
// It is public (not internal/) because generated programs import it.
package runtime

import (
	"fmt"
	"os"
	"strings"
)

// parents maps every supported exception class to its base class, following
// CPython's hierarchy. BaseException is the root.
var parents = map[string]string{
	"BaseException":       "",
	"Exception":           "BaseException",
	"ArithmeticError":     "Exception",
	"ZeroDivisionError":   "ArithmeticError",
	"OverflowError":       "ArithmeticError",
	"FloatingPointError":  "ArithmeticError",
	"LookupError":         "Exception",
	"IndexError":          "LookupError",
	"KeyError":            "LookupError",
	"ValueError":          "Exception",
	"UnicodeError":        "ValueError",
	"TypeError":           "Exception",
	"AssertionError":      "Exception",
	"AttributeError":      "Exception",
	"EOFError":            "Exception",
	"ImportError":         "Exception",
	"MemoryError":         "Exception",
	"NameError":           "Exception",
	"UnboundLocalError":   "NameError",
	"OSError":             "Exception",
	"StopIteration":       "Exception",
	"RuntimeError":        "Exception",
	"NotImplementedError": "RuntimeError",
	"RecursionError":      "RuntimeError",
}

// IsExceptionClass reports whether name is a supported exception class.
// Codegen uses it to reject unknown classes at compile time.
func IsExceptionClass(name string) bool {
	_, ok := parents[name]
	return ok
}

// IsSubclass reports whether class is parent or derives from it.
func IsSubclass(class, parent string) bool {
	for c := class; c != ""; c = parents[c] {
		if c == parent {
			return true
		}
	}
	return false
}

// PyError is a Python exception. It is the value of every panic the runtime
// and generated code raise on purpose.
type PyError struct {
	Type string // "ZeroDivisionError"
	Args []any  // the constructor arguments: strings, numbers, bools, lists
}

// NewError is raise Class(args...).
func NewError(class string, args ...any) *PyError {
	return &PyError{Type: class, Args: args}
}

// newError builds an error with a single message, the shape of every
// error the runtime raises itself.
func newError(class, msg string) *PyError { return &PyError{Type: class, Args: []any{msg}} }

func zeroDivision(msg string) *PyError  { return newError("ZeroDivisionError", msg) }
func indexError(msg string) *PyError    { return newError("IndexError", msg) }
func valueError(msg string) *PyError    { return newError("ValueError", msg) }
func overflowError(msg string) *PyError { return newError("OverflowError", msg) }

// MissingReturn is the panic when a function annotated with a return type
// reaches its end. Python would return None there, which breaks the annotation.
func MissingReturn(fn string) *PyError {
	return newError("TypeError", "function "+fn+" reached its end without returning a value")
}

// str is str(e): empty for no arguments, the argument's str for one, and
// the tuple of arguments for several. KeyError shows the repr of its key.
func (e *PyError) str() string {
	switch len(e.Args) {
	case 0:
		return ""
	case 1:
		if e.Type == "KeyError" {
			return Repr(e.Args[0])
		}
		return Str(e.Args[0])
	}
	return "(" + e.argsRepr() + ")"
}

func (e *PyError) argsRepr() string {
	parts := make([]string, len(e.Args))
	for i, a := range e.Args {
		parts[i] = Repr(a)
	}
	return strings.Join(parts, ", ")
}

// repr is repr(e), such as ValueError('bad').
func (e *PyError) repr() string { return e.Type + "(" + e.argsRepr() + ")" }

// Error is the last line CPython prints for an uncaught exception.
func (e *PyError) Error() string {
	if msg := e.str(); msg != "" {
		return e.Type + ": " + msg
	}
	return e.Type
}

// Match is the test of an except clause: the exception is an instance of
// one of the classes, or of a class derived from one.
func Match(e *PyError, classes ...string) bool {
	for _, c := range classes {
		if IsSubclass(e.Type, c) {
			return true
		}
	}
	return false
}

// Catch runs f and returns the exception it raised, or nil. Panics that
// are not Python exceptions are bugs and keep propagating.
func Catch(f func()) (caught *PyError) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(*PyError)
			if !ok {
				panic(r)
			}
			caught = e
		}
	}()
	f()
	return nil
}

// Report is deferred by the generated main. An exception that reached it
// was never caught, so print what CPython would, minus the frames, and
// exit with status 1.
func Report() {
	r := recover()
	if r == nil {
		return
	}
	e, ok := r.(*PyError)
	if !ok {
		panic(r)
	}
	fmt.Fprintln(os.Stderr, "Traceback (most recent call last):")
	fmt.Fprintln(os.Stderr, e.Error())
	os.Exit(1)
}

// maxDepth is how many nested calls of recursive functions are allowed.
// CPython's limit is 1000 frames and the module counts as one.
const maxDepth = 999

var depth int

// Enter and Leave bracket the body of a function that can recurse, so that
// runaway recursion raises RecursionError instead of overflowing the stack.
func Enter() {
	depth++
	if depth > maxDepth {
		depth--
		panic(newError("RecursionError", "maximum recursion depth exceeded"))
	}
}

func Leave() { depth-- }

// Package runtime is the support library for generated Go code.
// It is public (not internal/) because generated programs import it.
package runtime

// PyError is the panic value for a Python exception.
// Until phase 4 adds real exceptions, runtime helpers panic with it
// and the message matches what CPython prints.
type PyError struct {
	Type string // "ZeroDivisionError"
	Msg  string
}

func (e *PyError) Error() string { return e.Type + ": " + e.Msg }

func zeroDivision(msg string) *PyError { return &PyError{"ZeroDivisionError", msg} }

// MissingReturn is the panic when a function annotated with a return type
// reaches its end. Python would return None there, which breaks the annotation.
func MissingReturn(fn string) *PyError {
	return &PyError{"TypeError", "function " + fn + " reached its end without returning a value"}
}

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
	"BaseException":                         "",
	"Exception":                             "BaseException",
	"ArithmeticError":                       "Exception",
	"ZeroDivisionError":                     "ArithmeticError",
	"OverflowError":                         "ArithmeticError",
	"FloatingPointError":                    "ArithmeticError",
	"LookupError":                           "Exception",
	"IndexError":                            "LookupError",
	"KeyError":                              "LookupError",
	"ValueError":                            "Exception",
	"UnicodeError":                          "ValueError",
	"TypeError":                             "Exception",
	"AssertionError":                        "Exception",
	"AttributeError":                        "Exception",
	"FrozenInstanceError":                   "AttributeError",
	"EOFError":                              "Exception",
	"ImportError":                           "Exception",
	"MemoryError":                           "Exception",
	"NameError":                             "Exception",
	"UnboundLocalError":                     "NameError",
	"OSError":                               "Exception",
	"FileNotFoundError":                     "OSError",
	"PermissionError":                       "OSError",
	"TimeoutError":                          "OSError",
	"ConnectionError":                       "OSError",
	"IsADirectoryError":                     "OSError",
	"NotADirectoryError":                    "OSError",
	"FileExistsError":                       "OSError",
	"InterruptedError":                      "OSError",
	"ChildProcessError":                     "OSError",
	"ProcessLookupError":                    "OSError",
	"BlockingIOError":                       "OSError",
	"BrokenPipeError":                       "ConnectionError",
	"ConnectionAbortedError":                "ConnectionError",
	"ConnectionRefusedError":                "ConnectionError",
	"ConnectionResetError":                  "ConnectionError",
	"UnicodeTranslateError":                 "UnicodeError",
	"TabError":                              "IndentationError",
	"FutureWarning":                         "Warning",
	"ImportWarning":                         "Warning",
	"PendingDeprecationWarning":             "Warning",
	"ResourceWarning":                       "Warning",
	"SyntaxWarning":                         "Warning",
	"UnicodeWarning":                        "Warning",
	"BytesWarning":                          "Warning",
	"EncodingWarning":                       "Warning",
	"PythonFinalizationError":               "RuntimeError",
	"ModuleNotFoundError":                   "ImportError",
	"SyntaxError":                           "Exception",
	"IndentationError":                      "SyntaxError",
	"UnicodeDecodeError":                    "UnicodeError",
	"UnicodeEncodeError":                    "UnicodeError",
	"BufferError":                           "Exception",
	"ReferenceError":                        "Exception",
	"SystemError":                           "Exception",
	"StopAsyncIteration":                    "Exception",
	"Warning":                               "Exception",
	"UserWarning":                           "Warning",
	"DeprecationWarning":                    "Warning",
	"RuntimeWarning":                        "Warning",
	"StopIteration":                         "Exception",
	"RuntimeError":                          "Exception",
	"NotImplementedError":                   "RuntimeError",
	"RecursionError":                        "RuntimeError",
	"GeneratorExit":                         "BaseException",
	"BaseExceptionGroup":                    "BaseException",
	"ExceptionGroup":                        "Exception",
	"KeyboardInterrupt":                     "BaseException",
	"SystemExit":                            "BaseException",
	"re_pm_PatternError":                    "Exception",
	"json__decoder_pm_JSONDecodeError":      "ValueError",
	"asyncio__exceptions_pm_CancelledError": "BaseException",
	"asyncio__exceptions_pm_InvalidStateError":        "Exception",
	"asyncio__exceptions_pm_BrokenBarrierError":       "RuntimeError",
	"threading_pm_BrokenBarrierError":                 "RuntimeError",
	"struct_pm_error":                                 "Exception",
	"statistics_pm_StatisticsError":                   "ValueError",
	"queue_pm_Empty":                                  "Exception",
	"queue_pm_Full":                                   "Exception",
	"queue_pm_ShutDown":                               "Exception",
	"concurrent__futures___base_pm_Error":             "Exception",
	"concurrent__futures___base_pm_CancelledError":    "concurrent__futures___base_pm_Error",
	"concurrent__futures___base_pm_InvalidStateError": "concurrent__futures___base_pm_Error",
	"concurrent__futures___base_pm_BrokenExecutor":    "RuntimeError",
	"concurrent__futures__thread_pm_BrokenThreadPool": "concurrent__futures___base_pm_BrokenExecutor",
	"asyncio__queues_pm_QueueEmpty":                   "Exception",
	"asyncio__queues_pm_QueueFull":                    "Exception",
	"asyncio__queues_pm_QueueShutDown":                "Exception",
}

// IsExceptionClass reports whether name is a supported built-in exception
// class. Codegen uses it to reject unknown classes at compile time.
func IsExceptionClass(name string) bool {
	_, ok := parents[name]
	return ok
}

// BuiltinMRO is the chain of built-in exception classes from name up to
// BaseException, or nil if name is not one.
func BuiltinMRO(name string) []string {
	if !IsExceptionClass(name) {
		return nil
	}
	var out []string
	for c := name; c != ""; c = parents[c] {
		out = append(out, c)
	}
	return out
}

// userMRO holds the exception classes the program defines: each maps to the
// names of all its ancestors, itself first.
var userMRO = map[string][]string{}

// RegisterClass records a user-defined exception class and its method
// resolution order, so that except clauses and isinstance can walk it.
func RegisterClass(name string, mro ...string) { userMRO[name] = mro }

// IsSubclass reports whether class is parent or derives from it.
func IsSubclass(class, parent string) bool {
	if class == "ExceptionGroup" && parent == "BaseExceptionGroup" {
		return true
	}
	if mro, ok := userMRO[class]; ok {
		for _, c := range mro {
			if c == parent {
				return true
			}
		}
		return false
	}
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

	// Obj is the instance of a user-defined exception class, whose fields and
	// methods the program reads back. It is nil for built-in exceptions.
	Obj any

	// Cause is "raise X from Cause" and Context the exception being handled
	// when this one was raised. Both appear in the report of an uncaught error.
	Cause, Context  *PyError
	SuppressContext bool

	// Attrs holds the attributes of errors that have more than args: the
	// range of a UnicodeError, the errno of an OSError.
	Attrs map[string]any

	// Excs holds the exceptions of an ExceptionGroup, and Msg its message.
	Excs    []*PyError
	Msg     string
	IsGroup bool
}

// ExcBase is embedded in the instances of exception classes the program
// defines. It ties the instance to the *PyError that gets raised.
type ExcBase struct{ E *PyError }

// PyExc is the exception to panic with when the instance is raised.
func (b *ExcBase) PyExc() *PyError { return b.E }

// Init is what BaseException.__new__ does: record the class and the
// constructor arguments. self is the instance.
func (b *ExcBase) Init(class string, self any, args ...any) {
	b.E = &PyError{Type: class, Args: args, Obj: self}
}

// SetArgs is BaseException.__init__(*args).
func (b *ExcBase) SetArgs(args ...any) { b.E.Args = args }

// BaseStr and BaseRepr are str() and repr() of an exception from its
// arguments, which user classes without __str__ and __repr__ get.
func (e *PyError) BaseStr() string  { return e.baseStr() }
func (e *PyError) BaseRepr() string { return e.baseRepr() }

// RaiseFrom is "raise e from cause": cause may be nil for "from None".
func RaiseFrom(e, cause *PyError) *PyError {
	e.Cause = cause
	e.SuppressContext = true
	return e
}

// WithContext records the exception being handled when e is raised inside
// its handler, and returns e. An exception is never its own context.
func WithContext(e, handled *PyError) *PyError {
	if e != handled && e.Context == nil {
		e.Context = handled
	}
	return e
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
	if s, ok := e.Obj.(Strer); ok {
		return s.PyStr()
	}
	return e.baseStr()
}

func (e *PyError) baseStr() string {
	if e.IsGroup {
		n := len(e.Excs)
		if n == 1 {
			return e.Msg + " (1 sub-exception)"
		}
		return fmt.Sprintf("%s (%d sub-exceptions)", e.Msg, n)
	}
	switch len(e.Args) {
	case 0:
		return ""
	case 1:
		if IsSubclass(e.Type, "KeyError") {
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
func (e *PyError) repr() string {
	if s, ok := e.Obj.(Reprer); ok {
		return s.PyRepr()
	}
	return e.baseRepr()
}

func (e *PyError) baseRepr() string { return BaseTypeName(e.Type) + "(" + e.argsRepr() + ")" }

// Error is the last line CPython prints for an uncaught exception.
func (e *PyError) Error() string {
	if msg := e.str(); msg != "" {
		return QualTypeName(e.Type) + ": " + msg
	}
	return QualTypeName(e.Type)
}

// modMarker separates the module of a class defined in another module of the
// project from its name: "pkg__mod_pm_Err" is pkg.mod.Err.
const modMarker = "_pm_"

// BaseTypeName is the name of an exception class without its module.
func BaseTypeName(t string) string {
	if i := strings.LastIndex(t, modMarker); i >= 0 {
		return t[i+len(modMarker):]
	}
	return t
}

// QualTypeName is the name with its module as CPython prints it for an
// uncaught exception: "pkg.mod.Err".
func QualTypeName(t string) string {
	i := strings.LastIndex(t, modMarker)
	if i < 0 {
		return t
	}
	return strings.ReplaceAll(t[:i], "__", ".") + "." + t[i+len(modMarker):]
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
		runExitHooks()
		return
	}
	e, ok := r.(*PyError)
	if !ok {
		panic(r)
	}
	if IsSubclass(e.Type, "SystemExit") {
		code := exitCode(e)
		runExitHooks()
		os.Exit(code)
	}
	fmt.Fprint(os.Stderr, report(e, map[*PyError]bool{}))
	runExitHooks()
	os.Exit(1)
}

// report is what CPython prints for an uncaught exception: the chain of
// causes and contexts first, oldest first, then the exception itself.
// The frames are not recorded, so each traceback is only its header.
func report(e *PyError, seen map[*PyError]bool) string {
	seen[e] = true
	out := ""
	switch {
	case e.Cause != nil && !seen[e.Cause]:
		out = report(e.Cause, seen) + "\nThe above exception was the direct cause of the following exception:\n\n"
	case e.Cause == nil && !e.SuppressContext && e.Context != nil && !seen[e.Context]:
		out = report(e.Context, seen) + "\nDuring handling of the above exception, another exception occurred:\n\n"
	}
	return out + "Traceback (most recent call last):\n" + e.Error() + "\n"
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

// Bound is the read of a module variable from a function. It raises
// NameError if the module-level assignment has not run yet.
func Bound[T any](v T, set bool, name string) T {
	if !set {
		panic(unboundName(name))
	}
	return v
}

// unboundName is the error for reading a module variable that is not bound
// yet. The name of another module of the project is an attribute of it.
func unboundName(name string) *PyError {
	if i := strings.LastIndex(name, modMarker); i >= 0 {
		mod := strings.ReplaceAll(name[:i], "__", ".")
		return newError("AttributeError", "module '"+mod+"' has no attribute '"+name[i+len(modMarker):]+"'")
	}
	return newError("NameError", "name '"+name+"' is not defined")
}

// exitCode is the status of a program that ended with SystemExit: the
// integer it was given, 0 for none, and 1 after printing any other value to
// standard error.
func exitCode(e *PyError) int {
	if len(e.Args) == 0 || IsNone(e.Args[0]) {
		return 0
	}
	switch c := Unbox(e.Args[0]).(type) {
	case int64:
		return int(c & 0xff)
	case bool:
		return int(b2i(c))
	}
	fmt.Fprintln(os.Stderr, Str(e.Args[0]))
	return 1
}

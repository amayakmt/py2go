package runtime

// The standard library modules that py2go ships are written in Python
// (stdlib/*.py) and joined to the program like the modules of the project.
// What cannot be written in Python is registered here and called with
// _native("module.function", args...).

// NativeFn is a function of a native module.
type NativeFn func(args []any, kw []KW) any

var natives = map[string]NativeFn{}

// RegisterNative adds a native function; the name is "module.function".
func RegisterNative(name string, f NativeFn) { natives[name] = f }

// NativeCall is _native(name, args...): args[0] is the name.
func NativeCall(args []any, kw []KW) any {
	name, _ := Unbox(args[0]).(string)
	f, ok := natives[name]
	if !ok {
		panic(newError("NotImplementedError", "native function '"+name+"' does not exist"))
	}
	return f(args[1:], kw)
}

// nativeArgs checks the number of arguments of a native function the way a
// built-in function does.
func nativeArgs(name string, args []any, min, max int) {
	n := len(args)
	switch {
	case n < min && min == max:
		panic(newError("TypeError", name+"() takes exactly "+itoa(min)+" argument"+plural(int64(min), "", "s")+" ("+itoa(n)+" given)"))
	case n < min:
		panic(newError("TypeError", name+"() takes at least "+itoa(min)+" argument"+plural(int64(min), "", "s")+" ("+itoa(n)+" given)"))
	case n > max && min == max:
		panic(newError("TypeError", name+"() takes exactly "+itoa(max)+" argument"+plural(int64(max), "", "s")+" ("+itoa(n)+" given)"))
	case n > max:
		panic(newError("TypeError", name+"() takes at most "+itoa(max)+" argument"+plural(int64(max), "", "s")+" ("+itoa(n)+" given)"))
	}
}

// NativeCallPair is NativeCall with the arguments collected in a CallArgs.
func NativeCallPair(a CallArgs) any { return NativeCall(a.Args, a.KW) }

// NativeBuiltin is a call of a built-in function the runtime implements
// natively, such as open().
func NativeBuiltin(name string, args []any, kw []KW) any {
	return natives["builtins."+name](args, kw)
}

func NativeBuiltinPair(name string, a CallArgs) any { return NativeBuiltin(name, a.Args, a.KW) }

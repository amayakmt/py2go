package runtime

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
)

// Calls, attributes and methods of values whose type is not known.

// ClassRef gets a constructor when the class is defined, so that a class held
// in a variable can be called.
func RegisterCtor(name string, ctor any) { ClassOf(name).ctor = ctor }

// FuncInfo describes a function of the program for calls with keyword
// arguments: the names of its parameters, in the order of its Go signature.
type FuncInfo struct {
	Name     string
	Params   []string
	Defaults map[string]func() any
	VarArgs  string // name of *args, if any
	KwArgs   string // name of **kwargs, if any
	KwOnly   map[string]bool
}

var funcInfos = map[uintptr]*FuncInfo{}

// RegisterFunc records the parameters of a function (or a bound method, whose
// code pointer is shared by all its bound forms).
func RegisterFunc(f any, info *FuncInfo) {
	funcInfos[reflect.ValueOf(f).Pointer()] = info
}

// RegisterCtorInfo records the parameters of a class's constructor.
func RegisterCtorInfo(name string, info *FuncInfo) { ClassOf(name).info = info }

func funcInfoOf(f any) *FuncInfo {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return nil
	}
	return funcInfos[v.Pointer()]
}

// WithInfo registers the parameters of a function value and returns it.
func WithInfo[F any](f F, info *FuncInfo) F {
	RegisterFunc(f, info)
	return f
}

// AnyCursor reads a cursor of T as a cursor of anything.
func AnyCursor[T any](c Cursor[T]) Cursor[any] { return anyCursor[T]{c} }

// DynCall is f(args...).
func DynCall(f any, args ...any) any { return DynCallKw(f, args, nil) }

// DynCallKw is f(args..., name=value...) where kw holds the keywords in the
// order written.
func DynCallKw(f any, args []any, kw []KW) any {
	f = Unbox(f)
	switch fn := f.(type) {
	case nil:
		panic(newError("TypeError", "'NoneType' object is not callable"))
	case func(...any) any:
		if len(kw) > 0 {
			panic(newError("TypeError", "keyword arguments are not supported for this callable"))
		}
		return fn(args...)
	case *ClassRef:
		return callClass(fn, args, kw)
	case interface {
		PyCall(args []any, kw []KW) (any, bool)
	}:
		if r, ok := fn.PyCall(args, kw); ok {
			return r
		}
	}
	v := reflect.ValueOf(f)
	if v.Kind() == reflect.Func {
		return callReflect(v, funcInfoOf(f), args, kw)
	}
	panic(newError("TypeError", "'"+TypeName(f)+"' object is not callable"))
}

// KW is one keyword argument of a dynamic call.
type KW struct {
	Name string
	Val  any
}

func callClass(c *ClassRef, args []any, kw []KW) any {
	if c.ctor == nil {
		if c.Builtin {
			return callBuiltinType(c.Name, args, kw)
		}
		panic(newError("TypeError", "cannot create '"+c.Name+"' instances here"))
	}
	return callReflect(reflect.ValueOf(c.ctor), c.info, args, kw)
}

// callReflect calls a Go function with dynamic arguments: it binds keywords by
// the registered parameter names, fills defaults, converts each argument to
// the parameter's type, and boxes the result.
func callReflect(fv reflect.Value, info *FuncInfo, args []any, kw []KW) any {
	t := fv.Type()
	if info == nil && t.NumIn() > 0 && !t.IsVariadic() && strings.HasPrefix(t.In(t.NumIn()-1).Name(), "VarTuple[") {
		// A function with *args and no registered signature: its last
		// parameter takes the extra arguments.
		params := make([]string, t.NumIn()-1)
		for i := range params {
			params[i] = fmt.Sprintf("arg%d", i+1)
		}
		info = &FuncInfo{Params: params, VarArgs: "args"}
	}
	name := "function"
	if info != nil && info.Name != "" {
		name = info.Name
	}
	n := t.NumIn()
	in := make([]reflect.Value, n)
	given := make([]bool, n)

	var params []string
	if info != nil {
		params = info.Params
	}
	positional := n
	if info != nil {
		positional = len(info.Params) - len(info.KwOnly)
		if info.VarArgs != "" {
			positional = len(info.Params) - len(info.KwOnly)
		}
	}
	// Positional arguments fill the leading parameters; extras go to *args.
	var extra []any
	for i, a := range args {
		if i < positional && i < n {
			in[i] = convertTo(a, t.In(i), name, i)
			given[i] = true
		} else {
			extra = append(extra, a)
		}
	}
	if info == nil && len(extra) > 0 {
		panic(newError("TypeError", fmt.Sprintf("%s() takes %d positional argument%s but %d %s given", name, n, plural(int64(n), "", "s"), len(args), wasWere(len(args)))))
	}
	var kwExtra []KW
	for _, k := range kw {
		idx := -1
		for i, p := range params {
			if p == k.Name {
				idx = i
			}
		}
		switch {
		case idx >= 0 && !given[idx]:
			in[idx] = convertTo(k.Val, t.In(idx), name, idx)
			given[idx] = true
		case idx >= 0:
			panic(newError("TypeError", fmt.Sprintf("%s() got multiple values for argument '%s'", name, k.Name)))
		case info != nil && info.KwArgs != "":
			kwExtra = append(kwExtra, k)
		default:
			panic(newError("TypeError", fmt.Sprintf("%s() got an unexpected keyword argument '%s'", name, k.Name)))
		}
	}
	// *args and **kwargs are the last parameters of the Go function.
	next := len(params)
	if info != nil && info.VarArgs != "" {
		in[next] = packVarArgs(extra, t.In(next))
		given[next] = true
		next++
		extra = nil
	}
	if info != nil && info.KwArgs != "" {
		in[next] = packKwArgs(kwExtra, t.In(next))
		given[next] = true
	}
	if len(extra) > 0 {
		panic(newError("TypeError", fmt.Sprintf("%s() takes %d positional argument%s but %d %s given", name, positional, plural(int64(positional), "", "s"), len(args), wasWere(len(args)))))
	}
	var missing, missingKw []string
	for i := 0; i < n; i++ {
		if given[i] {
			continue
		}
		if info != nil && i < len(params) {
			if d, ok := info.Defaults[params[i]]; ok {
				in[i] = convertTo(d(), t.In(i), name, i)
				continue
			}
			if info.KwOnly[params[i]] {
				missingKw = append(missingKw, params[i])
			} else {
				missing = append(missing, params[i])
			}
			continue
		}
		missing = append(missing, fmt.Sprintf("arg%d", i+1))
	}
	if len(missing) > 0 {
		panic(typeError(MsgMissing(name, "positional", missing...)))
	}
	if len(missingKw) > 0 {
		panic(typeError(MsgMissing(name, "keyword-only", missingKw...)))
	}
	out := fv.Call(in)
	switch len(out) {
	case 0:
		return nil
	case 1:
		return out[0].Interface()
	}
	vals := make([]any, len(out))
	for i := range out {
		vals[i] = out[i].Interface()
	}
	return NewVarTuple(vals...)
}

func packVarArgs(extra []any, t reflect.Type) reflect.Value {
	p := reflect.New(t)
	p.MethodByName("SetFromAny").Call([]reflect.Value{reflect.ValueOf(extra)})
	return p.Elem()
}

func packKwArgs(kw []KW, t reflect.Type) reflect.Value {
	p := reflect.New(t.Elem())
	p.MethodByName("InitAny").Call(nil)
	set := p.MethodByName("SetItemAny")
	for _, k := range kw {
		set.Call([]reflect.Value{reflect.ValueOf(any(k.Name)), reflect.ValueOf(&k.Val).Elem()})
	}
	return p
}

// convertTo converts a dynamic argument to the Go type of a parameter.
func convertTo(a any, t reflect.Type, fn string, i int) reflect.Value {
	a = Unbox(a)
	if a == nil {
		return reflect.Zero(t)
	}
	av := reflect.ValueOf(a)
	if t.Kind() == reflect.Interface && av.Type().Implements(t) {
		return av
	}
	if av.Type().AssignableTo(t) {
		return av.Convert(t)
	}
	switch t.Kind() {
	case reflect.Float64:
		if k, ii, _ := numeric(a); k == isInt {
			return reflect.ValueOf(float64(ii))
		}
	case reflect.Int64:
		if b, ok := a.(bool); ok {
			return reflect.ValueOf(b2i(b))
		}
	case reflect.Struct:
		// An Optional parameter takes a value or None.
		if some := reflect.ValueOf(a); some.IsValid() && t.Name() != "" && strings.HasPrefix(t.Name(), "Opt[") {
			out := reflect.New(t).Elem()
			inner := t.Field(0).Type
			if some.Type().AssignableTo(inner) {
				setUnexported(out.Field(0), some)
				setUnexported(out.Field(1), reflect.ValueOf(true))
				return out
			}
		}
	}
	panic(newError("TypeError", fmt.Sprintf("%s() argument %d: expected %s, got %s", fn, i+1, goTypeName(t), TypeName(a))))
}

func setUnexported(f reflect.Value, v reflect.Value) {
	reflect.NewAt(f.Type(), f.Addr().UnsafePointer()).Elem().Set(v)
}

func goTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int64:
		return "int"
	case reflect.Float64:
		return "float"
	case reflect.String:
		return "str"
	case reflect.Bool:
		return "bool"
	}
	return t.String()
}

// ---- attributes ----

// attrGetter, attrSetter and so on are implemented by instances of the
// classes of the program.
type attrGetter interface {
	PyGetAttr(name string) (any, bool)
}
type attrSetter interface {
	PySetAttr(name string, v any) bool
}
type attrDeleter interface {
	PyDelAttr(name string) bool
}

// DynGetAttr is obj.name.
func DynGetAttr(obj any, name string) any {
	obj = Unbox(obj)
	if g, ok := obj.(attrGetter); ok {
		if v, ok := g.PyGetAttr(name); ok {
			return v
		}
		// A class may define __getattr__.
		if h, ok := obj.(interface{ PyGetAttrHook(name string) (any, bool) }); ok {
			if v, ok := h.PyGetAttrHook(name); ok {
				return v
			}
		}
		if e, ok := obj.(interface{ PyExc() *PyError }); ok {
			if v, ok := excAttr(e.PyExc(), name); ok {
				return v
			}
		}
		if name == "__class__" {
			if n, ok := obj.(interface{ PyClassName() string }); ok {
				return ClassOf(n.PyClassName())
			}
		}
		panic(NoAttribute(TypeName(obj), name))
	}
	if v, ok := builtinAttr(obj, name); ok {
		return v
	}
	if e, ok := obj.(*PyError); ok {
		if e.Obj != nil {
			if g, isInst := e.Obj.(attrGetter); isInst {
				if v, ok := g.PyGetAttr(name); ok {
					return v
				}
			}
		}
		if v, ok := excAttr(e, name); ok {
			return v
		}
	}
	if f, ok := funcName(obj); ok {
		switch name {
		case "__name__":
			return f
		case "__qualname__":
			return f
		case "__module__":
			return "__main__"
		}
		panic(newError("AttributeError", "'function' object has no attribute '"+name+"'"))
	}
	if c, ok := obj.(*ClassRef); ok {
		if c.Builtin {
			if f := unboundMethod(c.Name, name); f != nil {
				return f
			}
		}
		switch name {
		case "__name__", "__qualname__":
			return BaseTypeName(c.Name)
		case "__module__":
			if i := strings.LastIndex(c.Name, modMarker); i >= 0 {
				return strings.ReplaceAll(c.Name[:i], "__", ".")
			}
			if c.Builtin {
				return "builtins"
			}
			return "__main__"
		}
		panic(newError("AttributeError", "type object '"+c.Name+"' has no attribute '"+name+"'"))
	}
	panic(NoAttribute(TypeName(obj), name))
}

// DynHasAttr is hasattr(obj, name).
func DynHasAttr(obj any, name string) bool {
	return Catch(func() { DynGetAttr(obj, name) }) == nil
}

// DynGetAttrDefault is getattr(obj, name, default).
func DynGetAttrDefault(obj any, name string, def any) (r any) {
	if e := Catch(func() { r = DynGetAttr(obj, name) }); e != nil {
		if e.Type != "AttributeError" {
			panic(e)
		}
		return def
	}
	return r
}

// DynSetAttr is obj.name = v.
func DynSetAttr(obj any, name string, v any) {
	obj = Unbox(obj)
	if s, ok := obj.(attrSetter); ok {
		if s.PySetAttr(name, v) {
			return
		}
	}
	panic(newError("AttributeError", fmt.Sprintf("'%s' object has no attribute '%s' and no __dict__ for setting new attributes", TypeName(obj), name)))
}

// DynDelAttr is del obj.name.
func DynDelAttr(obj any, name string) {
	obj = Unbox(obj)
	if d, ok := obj.(attrDeleter); ok {
		if d.PyDelAttr(name) {
			return
		}
	}
	panic(NoAttribute(TypeName(obj), name))
}

func excAttr(e *PyError, name string) (any, bool) {
	if v, ok := e.Attrs[name]; ok {
		return v, true
	}
	switch name {
	case "code":
		if IsSubclass(e.Type, "SystemExit") {
			switch len(e.Args) {
			case 0:
				return nil, true
			case 1:
				return e.Args[0], true
			}
			return NewVarTuple[any](e.Args...), true
		}
	case "errno", "strerror", "filename", "filename2":
		if IsSubclass(e.Type, "OSError") {
			return nil, true
		}
	case "name", "path":
		if IsSubclass(e.Type, "ImportError") {
			return nil, true
		}
	case "args":
		return NewVarTuple[any](e.Args...), true
	case "__cause__":
		if e.Cause == nil {
			return nil, true
		}
		return e.Cause, true
	case "__context__":
		if e.Context == nil {
			return nil, true
		}
		return e.Context, true
	case "__suppress_context__":
		return e.SuppressContext, true
	case "value":
		return StopValue(e), true
	case "message":
		if e.IsGroup {
			return e.Msg, true
		}
	case "exceptions":
		if e.IsGroup {
			return GroupExceptions(e), true
		}
	}
	return nil, false
}

// ---- isinstance and type ----

// DynIsInstance is isinstance(x, cls) for a class, a builtin type or a tuple.
func DynIsInstance(x, cls any) bool {
	x, cls = Unbox(x), Unbox(cls)
	if t, ok := cls.(Seq); ok && t.Kind() == "tuple" {
		for i := int64(0); i < t.Len(); i++ {
			if DynIsInstance(x, t.At(i)) {
				return true
			}
		}
		return false
	}
	c, ok := cls.(*ClassRef)
	if !ok {
		panic(newError("TypeError", "isinstance() arg 2 must be a type, a tuple of types, or a union"))
	}
	return IsInstanceName(x, c.Name, c.Builtin)
}

// IsInstanceName is isinstance(x, C) for a class named C.
func IsInstanceName(x any, name string, builtin bool) bool {
	x = Unbox(x)
	if name == "object" && builtin {
		return true
	}
	if m, ok := x.(interface{ PyMRO() []string }); ok && !builtin {
		for _, n := range m.PyMRO() {
			if n == name {
				return true
			}
		}
		return false
	}
	if e, ok := x.(*PyError); ok {
		return IsSubclass(e.Type, name)
	}
	if e, ok := x.(interface{ PyExc() *PyError }); ok && builtin {
		return IsSubclass(e.PyExc().Type, name)
	}
	if !builtin {
		return false
	}
	tn := TypeName(x)
	if tn == name {
		return true
	}
	return name == "int" && tn == "bool"
}

// DynType is type(x) as a class object.
func DynType(x any) *ClassRef {
	x = Unbox(x)
	if n, ok := x.(Named); ok {
		return ClassOf(n.PyClassName())
	}
	if e, ok := x.(*PyError); ok {
		return ExcClass(e)
	}
	return BuiltinClassOf(TypeName(x))
}

// DynTypeName is type(x).__name__.
func DynTypeName(x any) string { return TypeName(Unbox(x)) }

// ---- builtin types as callables ----

func callBuiltinType(name string, args []any, kw []KW) any {
	arg := func(i int) any {
		if i < len(args) {
			return args[i]
		}
		return nil
	}
	switch name {
	case "int":
		if len(args) == 0 {
			return int64(0)
		}
		if len(args) > 1 {
			return DynIntBase(args[0], args[1])
		}
		if v, ok := kwValue(kw, "base"); ok {
			return DynIntBase(args[0], v)
		}
		return DynIntAny(args[0])
	case "bytes":
		return NewBytes(args, kw)
	case "bytearray":
		return NewByteArray(args, kw)
	case "complex":
		return NewComplex(args, kw)
	case "frozenset":
		return NewFrozenSet(args)
	case "slice":
		return NewSlice(args)
	case "range":
		return rangeFrom(args)
	case "object":
		return &PlainObject{}
	case "float":
		if len(args) == 0 {
			return 0.0
		}
		return DynFloat(args[0])
	case "str":
		if len(args) == 0 {
			return ""
		}
		return DynStr(args[0])
	case "bool":
		return len(args) > 0 && DynBool(args[0])
	case "list":
		if len(args) == 0 {
			return NewList[any]()
		}
		return DynList(args[0])
	case "tuple":
		if len(args) == 0 {
			return NewVarTuple[any]()
		}
		return DynTuple(args[0])
	case "set":
		if len(args) == 0 {
			return NewSet[any]()
		}
		return DynSet(args[0])
	case "dict":
		if len(args) == 0 {
			d := NewDict[any, any]()
			for _, k := range kw {
				d.Set(k.Name, k.Val)
			}
			return d
		}
		return DynDict(arg(0))
	}
	if IsExceptionClass(name) {
		return NewError(name, args...)
	}
	panic(newError("TypeError", "cannot call "+name+" with these arguments"))
}

// ---- sorting of names, for dir() ----

func sortedNames(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// builtinFn is a built-in function as the runtime calls it.
type builtinFn func(args []any, kw []KW) any

// BuiltinFunc is a built-in function used as a value, such as the len in
// map(len, words) or key=len.
func BuiltinFunc(name string) any {
	if f := builtinFunc(name); f != nil {
		return &Callable{Fn: f, Info: &FuncInfo{Name: name}}
	}
	return nil
}

func builtinFunc(name string) builtinFn {
	plain := func(f func(args ...any) any) builtinFn {
		return func(args []any, kw []KW) any {
			if len(kw) > 0 {
				panic(newError("TypeError", name+"() takes no keyword arguments"))
			}
			return f(args...)
		}
	}
	one := func(f func(a any) any) builtinFn {
		return plain(func(args ...any) any {
			if len(args) != 1 {
				panic(newError("TypeError", name+"() takes exactly one argument ("+itoa(len(args))+" given)"))
			}
			return f(args[0])
		})
	}
	kwOr := func(kw []KW, key string, def any) any {
		if v, ok := kwValue(kw, key); ok {
			return v
		}
		return def
	}
	switch name {
	case "len":
		return one(func(a any) any { return DynLen(a) })
	case "abs":
		return one(func(a any) any { return DynUnary("Abs", a) })
	case "repr":
		return one(func(a any) any { return DynRepr(a) })
	case "ascii":
		return one(func(a any) any { return Ascii(a) })
	case "any":
		return one(func(a any) any { return DynAny(a) })
	case "all":
		return one(func(a any) any { return DynAll(a) })
	case "hash":
		return one(func(a any) any { return Hash(Unbox(a)) })
	case "reversed":
		return one(func(a any) any { return DynReversed(a) })
	case "print":
		return func(args []any, kw []KW) any {
			sep, _ := kwOr(kw, "sep", " ").(string)
			end, _ := kwOr(kw, "end", "\n").(string)
			PrintOpts(sep, end, args...)
			return nil
		}
	case "sorted":
		return func(args []any, kw []KW) any {
			return DynSorted(argAt(args, 0), nilIfNone(kwOr(kw, "key", nil)), Truth(Unbox(kwOr(kw, "reverse", false))))
		}
	case "sum":
		return func(args []any, kw []KW) any {
			return DynSum(argAt(args, 0), kwOr(kw, "start", argOr(args, 1, int64(0))))
		}
	case "min", "max":
		return func(args []any, kw []KW) any {
			items := any(NewVarTuple(args...))
			if len(args) == 1 {
				items = args[0]
			}
			def, hasDef := kwValue(kw, "default")
			return DynMinMax(name == "max", items, nilIfNone(kwOr(kw, "key", nil)), def, hasDef)
		}
	case "round":
		return plain(func(args ...any) any { return DynRound(argAt(args, 0), argAt(args, 1)) })
	case "enumerate":
		return func(args []any, kw []KW) any {
			start, _ := intOf(Unbox(kwOr(kw, "start", argOr(args, 1, int64(0)))))
			return DynEnumerate(argAt(args, 0), start)
		}
	case "pow":
		return plain(func(args ...any) any { return DynPow(argAt(args, 0), argAt(args, 1), argAt(args, 2)) })
	case "divmod":
		return plain(func(args ...any) any { return DynDivmod(argAt(args, 0), argAt(args, 1)) })
	case "ord":
		return one(func(a any) any { return DynOrd(a) })
	case "chr":
		return one(func(a any) any { return DynChr(a) })
	case "bin":
		return one(func(a any) any { return DynBin(a) })
	case "hex":
		return one(func(a any) any { return DynHex(a) })
	case "oct":
		return one(func(a any) any { return DynOct(a) })
	case "format":
		return plain(func(args ...any) any {
			spec := ""
			if len(args) > 1 {
				spec = Unbox(args[1]).(string)
			}
			return DynFormat(argAt(args, 0), spec)
		})
	case "isinstance":
		return plain(func(args ...any) any { return DynIsInstance(argAt(args, 0), argAt(args, 1)) })
	case "issubclass":
		return plain(func(args ...any) any { return DynIsSubclass(argAt(args, 0), argAt(args, 1)) })
	case "callable":
		return one(func(a any) any { return DynCallable(a) })
	case "iter":
		return one(func(a any) any { return DynIterator(a) })
	case "next":
		return plain(func(args ...any) any { return DynNext(argAt(args, 0), argAt(args, 1), len(args) > 1) })
	case "getattr":
		return plain(func(args ...any) any {
			if len(args) > 2 {
				return DynGetAttrDefault(args[0], Unbox(args[1]).(string), args[2])
			}
			return DynGetAttr(args[0], Unbox(args[1]).(string))
		})
	case "hasattr":
		return plain(func(args ...any) any { return DynHasAttr(args[0], Unbox(args[1]).(string)) })
	case "zip":
		return plain(func(args ...any) any { return DynZip(args...) })
	case "map":
		return plain(func(args ...any) any { return DynMap(args[0], args[1:]...) })
	case "filter":
		return plain(func(args ...any) any { return DynFilter(argAt(args, 0), argAt(args, 1)) })
	case "list", "tuple", "set", "dict":
		return func(args []any, kw []KW) any { return callBuiltinType(name, args, kw) }
	}
	return nil
}

func argOr(args []any, i int, def any) any {
	if i < len(args) {
		return args[i]
	}
	return def
}

func nilIfNone(v any) any {
	if IsNone(v) {
		return nil
	}
	return v
}

// DynBuiltinCall is a call of a built-in function whose arguments were
// collected at run time.
func DynBuiltinCall(name string, a CallArgs) any {
	f := BuiltinFunc(name)
	if f == nil {
		panic(newError("NameError", "name '"+name+"' is not defined"))
	}
	return DynCallKw(f, a.Args, a.KW)
}

// AppendStar adds the items of an iterable to the arguments of a call: f(*xs).
func AppendStar(args []any, xs any) []any {
	for c := DynIter(xs).Iter(); c.Next(); {
		args = append(args, c.Value())
	}
	return args
}

// AppendDoubleStar adds the items of a dict to the keywords of a call:
// f(**d).
func AppendDoubleStar(kw []KW, d any) []KW {
	src := DynDict(d)
	for c := src.Iter(); c.Next(); {
		k := c.Value()
		name, ok := k.(string)
		if !ok {
			panic(newError("TypeError", "keywords must be strings"))
		}
		kw = append(kw, KW{Name: name, Val: src.Get(k)})
	}
	return kw
}

// CallArgs are the arguments of a call that was built up at run time.
type CallArgs struct {
	Args []any
	KW   []KW
}

// DynCallPair is DynCallKw with the arguments collected in a CallArgs.
func DynCallPair(f any, a CallArgs) any { return DynCallKw(f, a.Args, a.KW) }

// DynMethodPair is DynCallMethod for the same.
func DynMethodPair(obj any, name string, a CallArgs) any {
	return DynCallMethod(obj, name, a.Args, a.KW)
}

// Callable is a function value that carries the parameters it was defined
// with, so that calls through an object can use keywords and defaults. Each
// closure gets its own, since its defaults were evaluated when it was made.
type Callable struct {
	Fn   any
	Info *FuncInfo
}

func (c *Callable) PyCall(args []any, kw []KW) (any, bool) {
	if f, ok := c.Fn.(builtinFn); ok {
		return f(args, kw), true
	}
	return callReflect(reflect.ValueOf(c.Fn), c.Info, args, kw), true
}

func (c *Callable) PyRepr() string { return "<function " + c.Info.Name + ">" }

func wasWere(n int) string {
	if n == 1 {
		return "was"
	}
	return "were"
}

// unboundMethod is str.upper and the like: the method as a function of the
// receiver and its arguments.
func unboundMethod(class, name string) func(args ...any) any {
	var table map[string]bool
	switch class {
	case "str":
		table = strMethodNames
	case "list":
		table = listMethodNames
	case "dict":
		table = dictMethodNames
	case "set":
		table = setMethodNames
	case "tuple":
		table = tupleMethodNames
	case "int", "float":
		table = numMethodNames
	}
	if !table[name] {
		return nil
	}
	return func(args ...any) any {
		if len(args) == 0 {
			panic(newError("TypeError", "unbound method "+class+"."+name+"() needs an argument"))
		}
		if !IsInstanceName(args[0], class, true) {
			panic(newError("TypeError", "descriptor '"+name+"' for '"+class+"' objects doesn't apply to a '"+TypeName(Unbox(args[0]))+"' object"))
		}
		return DynCallMethod(args[0], name, args[1:], nil)
	}
}

// ---- more built-in functions ----

func DynOrd(v any) int64 {
	switch x := Unbox(v).(type) {
	case string:
		r := []rune(x)
		if len(r) != 1 {
			panic(newError("TypeError", "ord() expected a character, but string of length "+itoa(len(r))+" found"))
		}
		return int64(r[0])
	case Bytes, *ByteArray:
		bs, _ := bytesOf(x)
		if len(bs) != 1 {
			panic(newError("TypeError", "ord() expected a character, but string of length "+itoa(len(bs))+" found"))
		}
		return int64(bs[0])
	}
	panic(newError("TypeError", "ord() expected string of length 1, but "+TypeName(Unbox(v))+" found"))
}

func DynChr(v any) string {
	if isBig(Unbox(v)) {
		panic(valueError("chr() arg not in range(0x110000)"))
	}
	n, ok := intOf(Unbox(v))
	if !ok {
		panic(newError("TypeError", "'"+TypeName(v)+"' object cannot be interpreted as an integer"))
	}
	if n < 0 || n > 0x10ffff {
		panic(valueError("chr() arg not in range(0x110000)"))
	}
	return string(rune(n))
}

func dynBase(v any, base int, prefix string) string {
	n, ok := bigOf(Unbox(v))
	if !ok {
		panic(newError("TypeError", "'"+TypeName(v)+"' object cannot be interpreted as an integer"))
	}
	sign := ""
	if n.Sign() < 0 {
		sign = "-"
	}
	return sign + prefix + new(big.Int).Abs(n).Text(base)
}

func DynBin(v any) string { return dynBase(v, 2, "0b") }
func DynOct(v any) string { return dynBase(v, 8, "0o") }
func DynHex(v any) string { return dynBase(v, 16, "0x") }

// DynPow is pow(a, b) and pow(a, b, mod).
func DynPow(a, b, mod any) any {
	if mod == nil || IsNone(mod) {
		if !isNumberVal(a) || !isNumberVal(b) {
			if _, ok := Unbox(a).(interface {
				PyBinop(string, any, bool) (any, bool)
			}); !ok {
				panic(newError("TypeError", "unsupported operand type(s) for ** or pow(): '"+TypeName(a)+"' and '"+TypeName(b)+"'"))
			}
		}
		return DynBinary("Pow", a, b)
	}
	x, ok1 := bigOf(Unbox(a))
	y, ok2 := bigOf(Unbox(b))
	m, ok3 := bigOf(Unbox(mod))
	if !ok1 || !ok2 || !ok3 {
		panic(newError("TypeError", "pow() 3rd argument not allowed unless all arguments are integers"))
	}
	if m.Sign() == 0 {
		panic(valueError("pow() 3rd argument cannot be 0"))
	}
	absM := new(big.Int).Abs(m)
	base := new(big.Int).Mod(x, absM) // non-negative, as Python's % for a positive modulus
	if y.Sign() < 0 {
		inv := new(big.Int).ModInverse(base, absM)
		if inv == nil {
			panic(valueError("base is not invertible for the given modulus"))
		}
		base, y = inv, new(big.Int).Neg(y)
	}
	r := new(big.Int).Exp(base, y, absM)
	if m.Sign() < 0 && r.Sign() != 0 {
		r.Add(r, m)
	}
	return NormBig(r)
}

// DynDivmod is divmod(a, b).
func DynDivmod(a, b any) VarTuple[any] {
	isNum := func(v any) bool { v = Unbox(v); return isIntVal(v) || isFloatValue(v) }
	if !isNum(a) || !isNum(b) {
		panic(newError("TypeError", "unsupported operand type(s) for divmod(): '"+TypeName(a)+"' and '"+TypeName(b)+"'"))
	}
	if bf, ok := Unbox(b).(float64); ok && bf == 0 {
		panic(zeroDivision("float divmod()"))
	}
	if _, ok := Unbox(a).(float64); ok {
		if bi, ok := bigOf(Unbox(b)); ok && bi.Sign() == 0 {
			panic(zeroDivision("float divmod()"))
		}
	}
	return NewVarTuple[any](DynBinary("FloorDiv", a, b), DynBinary("Mod", a, b))
}

// DynFormat is format(v, spec).
func DynFormat(v any, spec string) string { return FormatDyn(v, spec) }

// DynCallable is callable(x).
func DynCallable(x any) bool {
	x = Unbox(x)
	switch x.(type) {
	case nil:
		return false
	case *ClassRef, *Callable:
		return true
	case interface {
		PyCall(args []any, kw []KW) (any, bool)
	}:
		return true
	}
	return reflect.ValueOf(x).Kind() == reflect.Func
}

// AsException turns an object into the error a raise statement throws: an
// exception, an instance of an exception class, or an exception class.
func AsException(x any) *PyError {
	x = Unbox(x)
	switch e := x.(type) {
	case *PyError:
		return e
	case interface{ PyExc() *PyError }:
		return e.PyExc()
	case *ClassRef:
		if v, ok := DynCallKw(e, nil, nil).(interface{ PyExc() *PyError }); ok {
			return v.PyExc()
		}
		if v, ok := DynCallKw(e, nil, nil).(*PyError); ok {
			return v
		}
	}
	return newError("TypeError", "exceptions must derive from BaseException")
}

// DynNext is next(it) and next(it, default).
func DynNext(it any, def any, hasDef bool) (result any) {
	it = Unbox(it)
	var cur Cursor[any]
	switch x := it.(type) {
	case interface{ PyNext() any }:
		if e := Catch(func() { result = x.PyNext() }); e != nil {
			if hasDef && IsSubclass(e.Type, "StopIteration") {
				return def
			}
			panic(e)
		}
		return result
	case interface{ NextAny() (any, bool) }:
		if v, ok := x.NextAny(); ok {
			return v
		}
		if hasDef {
			return def
		}
		panic(newError("StopIteration", ""))
	case interface{ IterAny() Cursor[any] }:
		cur = x.IterAny()
		_ = cur
	}
	panic(newError("TypeError", "'"+TypeName(it)+"' object is not an iterator"))
}

// funcName is the name of a function value, from the parameters registered
// with it.
func funcName(v any) (string, bool) {
	if c, ok := v.(*Callable); ok {
		return nameOf(c.Info), true
	}
	rv := reflect.ValueOf(v)
	if v == nil || rv.Kind() != reflect.Func {
		return "", false
	}
	return nameOf(funcInfoOf(v)), true
}

func nameOf(info *FuncInfo) string {
	if info == nil || info.Name == "" {
		return "function"
	}
	return info.Name
}

// DynIterator is iter(x): an iterator over x.
func DynIterator(v any) any {
	v = Unbox(v)
	if it, ok := v.(*Iterator[any]); ok {
		return it
	}
	c := DynIter(v).Iter()
	return NewIterator(func() (any, bool) {
		if c.Next() {
			return c.Value(), true
		}
		return nil, false
	})
}

// DynIntBase is int(text, base) for bases 2 to 36 and 0.
func DynIntBase(v, base any) any {
	var text string
	switch x := Unbox(v).(type) {
	case string:
		text = x
	case Bytes:
		text = string(x)
	case *ByteArray:
		text = string(x.b)
	default:
		panic(newError("TypeError", "int() can't convert non-string with explicit base"))
	}
	b, ok := intOf(Unbox(base))
	if !ok {
		panic(newError("TypeError", "'"+TypeName(base)+"' object cannot be interpreted as an integer"))
	}
	if b != 0 && (b < 2 || b > 36) {
		panic(valueError("int() base must be >= 2 and <= 36, or 0"))
	}
	body := strings.TrimSpace(text)
	sign := ""
	if strings.HasPrefix(body, "-") || strings.HasPrefix(body, "+") {
		sign, body = body[:1], body[1:]
	}
	lower := strings.ToLower(body)
	origBase := b
	for prefix, pb := range map[string]int64{"0x": 16, "0o": 8, "0b": 2} {
		if strings.HasPrefix(lower, prefix) && (b == pb || b == 0) {
			body, b = body[2:], pb
			body = strings.TrimPrefix(body, "_") // 0x_ff
		}
	}
	if b == 0 {
		b = 10
		if strings.HasPrefix(body, "0") && strings.Trim(body, "0_") != "" {
			panic(valueError("invalid literal for int() with base 0: " + quote(text)))
		}
	}
	if body == "" || strings.HasPrefix(body, "_") || strings.HasSuffix(body, "_") || strings.Contains(body, "__") {
		panic(valueError("invalid literal for int() with base " + itoa(int(origBase)) + ": " + quote(text)))
	}
	return IntFromText(sign+body, int(b), text)
}

// DynIsSubclass is issubclass(sub, sup) for classes held in objects.
func DynIsSubclass(sub, sup any) bool {
	sub, sup = Unbox(sub), Unbox(sup)
	c, ok := sub.(*ClassRef)
	if !ok {
		panic(newError("TypeError", "issubclass() arg 1 must be a class"))
	}
	if t, ok := sup.(Seq); ok && t.Kind() == "tuple" {
		for i := int64(0); i < t.Len(); i++ {
			if DynIsSubclass(c, t.At(i)) {
				return true
			}
		}
		return false
	}
	p, ok := sup.(*ClassRef)
	if !ok {
		panic(newError("TypeError", "issubclass() arg 2 must be a class, a tuple of classes, or a union"))
	}
	if p.Name == "object" && p.Builtin || c.Name == p.Name || c.Name == "bool" && p.Name == "int" {
		return true
	}
	return IsSubclass(c.Name, p.Name)
}

// UndefinedName is reading a name that nothing defines.
func UndefinedName(name string) any {
	panic(unboundName(name))
}

// PlainObject is object(): an instance with nothing in it.
type PlainObject struct{ _ byte }

func (*PlainObject) PyClassName() string { return "object" }

func rangeFrom(args []any) any {
	n := make([]int64, len(args))
	var bigs []*big.Int
	for i, a := range args {
		v := Unbox(a)
		bi, ok := bigOf(v)
		if !ok {
			panic(newError("TypeError", "'"+TypeName(v)+"' object cannot be interpreted as an integer"))
		}
		bigs = append(bigs, bi)
		if !bi.IsInt64() {
			n = nil
		} else if n != nil {
			n[i] = bi.Int64()
		}
	}
	if n == nil {
		switch len(bigs) {
		case 1:
			return rangeOfBig(new(big.Int), bigs[0], big.NewInt(1))
		case 2:
			return rangeOfBig(bigs[0], bigs[1], big.NewInt(1))
		default:
			return rangeOfBig(bigs[0], bigs[1], bigs[2])
		}
	}
	switch len(n) {
	case 1:
		return NewRange(0, n[0], 1)
	case 2:
		return NewRange(n[0], n[1], 1)
	case 3:
		return NewRange(n[0], n[1], n[2])
	}
	panic(newError("TypeError", "range expected at least 1 argument, got "+itoa(len(n))))
}

// Ascii is ascii(x): repr with every non-ASCII character escaped.
func Ascii(v any) string {
	r := Repr(Unbox(v))
	var b strings.Builder
	for _, c := range r {
		if c < 0x80 {
			b.WriteRune(c)
		} else {
			b.WriteString(charEscape(c))
		}
	}
	return b.String()
}

// StrDecode is str(bytes, encoding[, errors]).
func StrDecode(args ...any) string {
	s, ok := bytesOf(args[0])
	if !ok {
		panic(newError("TypeError", "decoding to str: need a bytes-like object, "+TypeName(Unbox(args[0]))+" found"))
	}
	enc := argNamed("str", "encoding", args, 1, "utf-8")
	errs := argNamed("str", "errors", args, 2, "strict")
	return DecodeBytes(s, enc, errs)
}

// DynRange is range(...) with arguments of any size.
func DynRange(args ...any) any { return rangeFrom(args) }

func isNumberVal(v any) bool {
	switch Unbox(v).(type) {
	case int64, float64, bool, complex128, *big.Int:
		return true
	}
	return false
}

// ExcAttr is e.name for the attributes of the built-in exceptions that the
// compiler does not know the type of.
func ExcAttr(e *PyError, name string) any {
	if v, ok := excAttr(e, name); ok {
		return v
	}
	panic(NoAttribute(TypeName(e), name))
}

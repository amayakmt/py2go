package runtime

import "strings"

// Support for pickle: finding classes and functions by module and name, and
// making an instance without running __init__.

var (
	globalsReg = map[string]any{}
	allocReg   = map[string]func() any{}
)

func globalKey(module, name string) string { return module + "\x00" + name }

// RegisterGlobal records a function of a module of the program, so pickle can
// name it and find it again.
func RegisterGlobal(module, name string, v any) { globalsReg[globalKey(module, name)] = v }

// RegisterAlloc records how to make an empty instance of a class.
func RegisterAlloc(class string, alloc func() any) { allocReg[class] = alloc }

func moduleOf(c *ClassRef) string {
	if c.Builtin {
		return "builtins"
	}
	if c.Module != "" {
		return c.Module
	}
	return "__main__"
}

var pickleBuiltinNames = map[string]bool{
	"int": true, "float": true, "str": true, "bool": true, "list": true, "tuple": true, "set": true, "dict": true,
	"bytes": true, "bytearray": true, "complex": true, "frozenset": true, "slice": true, "range": true, "object": true, "type": true,
}

func init() {
	// pickle.global_name(obj): (module, name) of a class or function.
	RegisterNative("pickle.global_name", func(args []any, kw []KW) any {
		nativeArgs("global_name", args, 1, 1)
		obj := Unbox(args[0])
		if c, ok := obj.(*ClassRef); ok {
			return NewVarTuple[any](moduleOf(c), BaseTypeName(c.Name))
		}
		for k, v := range globalsReg {
			if r, isFunc := sameFunc(v, obj); isFunc && r {
				i := strings.IndexByte(k, 0)
				return NewVarTuple[any](k[:i], k[i+1:])
			}
		}
		if name, ok := funcName(obj); ok && !strings.HasPrefix(name, "<") && name != "function" {
			return NewVarTuple[any]("builtins", name)
		}
		return nil
	})
	RegisterNative("pickle.find_global", func(args []any, kw []KW) any {
		nativeArgs("find_global", args, 2, 2)
		module, name := Str(args[0]), Str(args[1])
		if module == "builtins" || module == "__builtin__" {
			if pickleBuiltinNames[name] || IsExceptionClass(name) {
				return BuiltinClassOf(name)
			}
			return BuiltinFunc(name)
		}
		if v, ok := globalsReg[globalKey(module, name)]; ok {
			return v
		}
		if c, ok := classRefs[name]; ok && !c.Builtin && (moduleOf(c) == module) {
			return c
		}
		// A class whose name other modules' classes share is registered
		// under its mangled name.
		for key, c := range classRefs {
			if !c.Builtin && BaseTypeName(key) == name && moduleOf(c) == module {
				return c
			}
		}
		if _, known := moduleKnown(module); !known {
			panic(newError("ModuleNotFoundError", "No module named '"+module+"'"))
		}
		panic(newError("AttributeError", "Can't get attribute '"+name+"' on <module '"+module+"'>"))
	})
	RegisterNative("pickle.is_instance", func(args []any, kw []KW) any {
		nativeArgs("is_instance", args, 1, 1)
		_, ok := Unbox(args[0]).(attrNamer)
		return ok
	})
	RegisterNative("pickle.is_class", func(args []any, kw []KW) any {
		nativeArgs("is_class", args, 1, 1)
		_, ok := Unbox(args[0]).(*ClassRef)
		return ok
	})
	RegisterNative("pickle.is_exception", func(args []any, kw []KW) any {
		nativeArgs("is_exception", args, 1, 1)
		switch Unbox(args[0]).(type) {
		case *PyError, interface{ PyExc() *PyError }:
			return true
		}
		return false
	})
	// pickle.func_name(f): the name f was registered under, "" if unknown.
	RegisterNative("pickle.func_name", func(args []any, kw []KW) any {
		nativeArgs("func_name", args, 1, 1)
		obj := Unbox(args[0])
		for k, v := range globalsReg {
			if r, isFunc := sameFunc(v, obj); isFunc && r {
				return k[strings.IndexByte(k, 0)+1:]
			}
		}
		return ""
	})
	RegisterNative("pickle.newobj", func(args []any, kw []KW) any {
		nativeArgs("newobj", args, 2, 2)
		c, ok := Unbox(args[0]).(*ClassRef)
		if !ok {
			panic(newError("TypeError", "NEWOBJ class argument isn't a type object"))
		}
		var cargs []any
		for it := DynIter(args[1]).Iter(); it.Next(); {
			cargs = append(cargs, it.Value())
		}
		if c.Builtin {
			return DynCallKw(c, cargs, nil)
		}
		// object.__new__: the instance is allocated without running __init__.
		if alloc, ok := allocReg[c.Name]; ok {
			return alloc()
		}
		return DynCallKw(c, nil, nil)
	})
}

func moduleKnown(module string) (struct{}, bool) {
	for k := range globalsReg {
		if strings.HasPrefix(k, module+"\x00") {
			return struct{}{}, true
		}
	}
	for _, c := range classRefs {
		if !c.Builtin && moduleOf(c) == module {
			return struct{}{}, true
		}
	}
	return struct{}{}, module == "__main__"
}

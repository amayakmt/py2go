package runtime

func init() {
	// inspect.sigdata(f) is None, or (name, [(param, kind, has_default, default)], varargs, kwargs, is_async):
	// what the program knows about the parameters of a function value.
	RegisterNative("inspect.sigdata", func(args []any, kw []KW) any {
		f := Unbox(args[0])
		var info *FuncInfo
		if c, ok := f.(*Callable); ok {
			if _, builtin := c.Fn.(builtinFn); builtin {
				return nil
			}
			info = c.Info
		} else {
			info = funcInfoOf(f)
		}
		if info == nil {
			return nil
		}
		params := NewList[any]()
		for _, p := range info.Params {
			kind := "pos"
			if info.KwOnly[p] {
				kind = "kwonly"
			}
			def, has := info.Defaults[p]
			var dv any
			if has && def != nil {
				dv = def()
			}
			params.Append(NewVarTuple[any](p, kind, has, dv))
		}
		return NewVarTuple[any](info.Name, params, info.VarArgs, info.KwArgs, info.Async)
	})
}

package runtime

// NewGroup is ExceptionGroup(msg, excs) and BaseExceptionGroup(msg, excs).
func NewGroup(class, msg string, excs []*PyError) *PyError {
	if len(excs) == 0 {
		panic(newError("ValueError", "second argument (exceptions) must be a non-empty sequence"))
	}
	if class == "ExceptionGroup" {
		for _, x := range excs {
			if !IsSubclass(x.Type, "Exception") {
				panic(newError("TypeError", "Cannot nest BaseExceptions in an ExceptionGroup"))
			}
		}
	}
	list := NewList(excs...)
	return &PyError{Type: class, Args: []any{msg, list}, Excs: excs, Msg: msg, IsGroup: true}
}

// Subgroup is eg.subgroup(match) and Split is eg.split(match): the exceptions
// of a group for which match is true, in a group of the same shape and message.
// Leaves that are not groups are tested themselves.
func Split(e *PyError, match func(*PyError) bool) (matched, rest *PyError) {
	if e == nil {
		return nil, nil
	}
	if !e.IsGroup {
		if match(e) {
			return e, nil
		}
		return nil, e
	}
	if match(e) {
		return e, nil
	}
	var in, out []*PyError
	for _, x := range e.Excs {
		m, r := Split(x, match)
		if m != nil {
			in = append(in, m)
		}
		if r != nil {
			out = append(out, r)
		}
	}
	derive := func(items []*PyError) *PyError {
		if len(items) == 0 {
			return nil
		}
		g := NewGroup(e.Type, e.Msg, items)
		g.Cause, g.Context, g.SuppressContext = e.Cause, e.Context, e.SuppressContext
		return g
	}
	return derive(in), derive(out)
}

// SplitClasses splits by exception class names.
func SplitClasses(e *PyError, classes ...string) (matched, rest *PyError) {
	return Split(e, func(x *PyError) bool { return Match(x, classes...) })
}

// SplitStar is one except* clause: the part of rem that the classes catch,
// and what is left. An exception that is not a group is wrapped in one when
// it matches, as Python does.
func SplitStar(rem *PyError, classes ...string) (matched, rest *PyError) {
	if rem == nil {
		return nil, nil
	}
	if !rem.IsGroup {
		if Match(rem, classes...) {
			g := NewGroup("ExceptionGroup", "", []*PyError{rem})
			if !IsSubclass(rem.Type, "Exception") {
				g = NewGroup("BaseExceptionGroup", "", []*PyError{rem})
			}
			return g, nil
		}
		return nil, rem
	}
	return SplitClasses(rem, classes...)
}

// GroupExceptions is eg.exceptions.
func GroupExceptions(e *PyError) VarTuple[*PyError] {
	if !e.IsGroup {
		panic(newError("AttributeError", "'"+e.Type+"' object has no attribute 'exceptions'"))
	}
	return NewVarTuple(e.Excs...)
}

// GroupMessage is eg.message.
func GroupMessage(e *PyError) string {
	if !e.IsGroup {
		panic(newError("AttributeError", "'"+e.Type+"' object has no attribute 'message'"))
	}
	return e.Msg
}

// Optional wraps an exception that may be nil.
func OptExc(e *PyError) Opt[*PyError] {
	if e == nil {
		return NoneOf[*PyError]()
	}
	return Some(e)
}

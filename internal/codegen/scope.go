package codegen

// scope is one block's variables. Blocks nest, and a function body starts
// a fresh chain with no parent: Python functions cannot see the variables
// of main() here, since module-level code is compiled into main().
//
// A name is looked up through every enclosing block of the function, so
// "x: int = 1" inside an if reassigns an x declared before it, as in Python,
// instead of shadowing it as Go's := would.
type scope struct {
	vars   map[string]pyType
	sigs   map[string]*funcSig // nested functions, which keep their full signature
	parent *scope
	// boundary marks the root scope of a function. Its parent is the
	// enclosing function, whose variables the function can read, and assign
	// only after "nonlocal".
	boundary bool
	aliases  map[string]aliasInfo // names narrowed in this scope, see narrow.go
}

func newScope(parent *scope) *scope {
	return &scope{vars: map[string]pyType{}, sigs: map[string]*funcSig{}, parent: parent}
}

// lookupSig finds the innermost binding of name. If that binding is a
// function with a known signature it is returned; a variable of the same
// name in an inner scope hides an outer function.
func (s *scope) lookupSig(name string) (sig *funcSig, found bool) {
	for ; s != nil; s = s.parent {
		if _, ok := s.vars[name]; ok {
			return s.sigs[name], true
		}
	}
	return nil, false
}

// lookupWithinFunction is lookup that stops at the root of the current
// function: the variables an assignment can reach without "nonlocal".
func (s *scope) lookupWithinFunction(name string) (pyType, bool) {
	for ; s != nil; s = s.parent {
		if t, ok := s.vars[name]; ok {
			return t, true
		}
		if s.boundary {
			break
		}
	}
	return "", false
}

// functionRoot is the root scope of the function s belongs to.
func (s *scope) functionRoot() *scope {
	for ; s != nil; s = s.parent {
		if s.boundary || s.parent == nil {
			return s
		}
	}
	return nil
}

func (s *scope) lookup(name string) (pyType, bool) {
	for ; s != nil; s = s.parent {
		if t, ok := s.vars[name]; ok {
			return t, true
		}
	}
	return "", false
}

// declare adds a name to this block. It reports false if the name already
// exists in this block.
func (s *scope) declare(name string, t pyType) bool {
	if _, dup := s.vars[name]; dup {
		return false
	}
	s.vars[name] = t
	return true
}

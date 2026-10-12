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
	parent *scope
}

func newScope(parent *scope) *scope {
	return &scope{vars: map[string]pyType{}, parent: parent}
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

// funcSig is the signature of a function defined at module level.
type funcSig struct {
	params []pyType
	ret    pyType // tNone for "-> None"
}

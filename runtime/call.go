package runtime

import (
	"fmt"
	"strings"
)

// The messages of the TypeErrors CPython raises for a call that does not
// match the function's parameters. Codegen uses the Msg functions for calls it
// can check at compile time, and the generated code panics with the others
// for calls that depend on the length of a *args list or the keys of a
// **kwargs dict.

func plural(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func quoteNames(names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = "'" + n + "'"
	}
	switch len(q) {
	case 1:
		return q[0]
	case 2:
		return q[0] + " and " + q[1]
	}
	return strings.Join(q[:len(q)-1], ", ") + ", and " + q[len(q)-1]
}

// MsgTooManyPositional is "f() takes 2 positional arguments but 3 were given".
// min and max bound the accepted count; kwonly is how many keyword-only
// arguments the call supplied.
func MsgTooManyPositional(fn string, given, min, max, kwonly int64) string {
	var takes string
	if min == max {
		takes = fmt.Sprintf("%d positional %s", max, plural(max, "argument", "arguments"))
	} else {
		takes = fmt.Sprintf("from %d to %d positional arguments", min, max)
	}
	if kwonly > 0 {
		return fmt.Sprintf("%s() takes %s but %d positional %s (and %d keyword-only %s) %s given", fn, takes, given,
			plural(given, "argument", "arguments"), kwonly, plural(kwonly, "argument", "arguments"), plural(given, "was", "were"))
	}
	return fmt.Sprintf("%s() takes %s but %d %s given", fn, takes, given, plural(given, "was", "were"))
}

// MsgMissing is "f() missing 2 required positional arguments: 'a' and 'b'".
// kind is "positional" or "keyword-only".
func MsgMissing(fn, kind string, names ...string) string {
	return fmt.Sprintf("%s() missing %d required %s %s: %s", fn, len(names), kind,
		plural(int64(len(names)), "argument", "arguments"), quoteNames(names))
}

// MsgUnexpectedKeyword is "f() got an unexpected keyword argument 'c'", with
// CPython's suggestion when one parameter name is close.
func MsgUnexpectedKeyword(fn, name, suggestion string) string {
	msg := fmt.Sprintf("%s() got an unexpected keyword argument '%s'", fn, name)
	if suggestion != "" {
		msg += fmt.Sprintf(". Did you mean '%s'?", suggestion)
	}
	return msg
}

func MsgMultipleValues(fn, name string) string {
	return fmt.Sprintf("%s() got multiple values for argument '%s'", fn, name)
}

func MsgPositionalOnlyAsKeyword(fn string, names ...string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = "'" + n + "'"
	}
	return fmt.Sprintf("%s() got some positional-only arguments passed as keyword arguments: %s", fn, strings.Join(q, ", "))
}

func typeError(msg string) *PyError { return newError("TypeError", msg) }

// The Raise functions are what generated code calls for the checks it cannot
// make at compile time.

func TooManyPositional(fn string, given, min, max, kwonly int64) {
	panic(typeError(MsgTooManyPositional(fn, given, min, max, kwonly)))
}

func MissingPositional(fn string, names ...string) {
	panic(typeError(MsgMissing(fn, "positional", names...)))
}

func MissingKeywordOnly(fn string, names ...string) {
	panic(typeError(MsgMissing(fn, "keyword-only", names...)))
}

func UnexpectedKeyword(fn, name, suggestion string) {
	panic(typeError(MsgUnexpectedKeyword(fn, name, suggestion)))
}

func MultipleValues(fn, name string) { panic(typeError(MsgMultipleValues(fn, name))) }

// Suggest is CPython's "Did you mean" search: the candidate nearest to name by
// edit distance, if it is near enough.
func Suggest(name string, candidates []string) string {
	const moveCost, caseCost, maxSize = 2, 1, 40
	sub := func(a, b byte) int {
		if a == b {
			return 0
		}
		if lower(a) == lower(b) {
			return caseCost
		}
		return moveCost
	}
	distance := func(a, b string, maxCost int) int {
		if a == b {
			return 0
		}
		for len(a) > 0 && len(b) > 0 && a[0] == b[0] {
			a, b = a[1:], b[1:]
		}
		for len(a) > 0 && len(b) > 0 && a[len(a)-1] == b[len(b)-1] {
			a, b = a[:len(a)-1], b[:len(b)-1]
		}
		if len(a) == 0 || len(b) == 0 {
			return (len(a) + len(b)) * moveCost
		}
		if len(a) > maxSize || len(b) > maxSize {
			return maxCost + 1
		}
		if len(b) < len(a) {
			a, b = b, a
		}
		if (len(b)-len(a))*moveCost > maxCost {
			return maxCost + 1
		}
		buf := make([]int, len(a))
		tmp := moveCost
		for i := range buf {
			buf[i] = tmp
			tmp += moveCost
		}
		result := 0
		for bi := 0; bi < len(b); bi++ {
			code := b[bi]
			dist := bi * moveCost
			result = dist
			minimum := 1 << 30
			for i := 0; i < len(a); i++ {
				substitute := dist + sub(code, a[i])
				dist = buf[i]
				insertDelete := min(result, dist) + moveCost
				result = min(insertDelete, substitute)
				buf[i] = result
				minimum = min(minimum, result)
			}
			if minimum > maxCost {
				return maxCost + 1
			}
		}
		return result
	}

	best, bestDistance := "", 1<<30
	for _, c := range candidates {
		if c == name {
			continue
		}
		maxDistance := min((len(name)+len(c)+3)*moveCost/6, bestDistance-1)
		d := distance(name, c, maxDistance)
		if d > maxDistance {
			continue
		}
		if best == "" || d < bestDistance {
			best, bestDistance = c, d
		}
	}
	return best
}

func lower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

// SeqList turns a tuple of items that share a type into a list.
func SeqList[T any](s Seq) *List[T] {
	out := NewList[T]()
	for i := int64(0); i < s.Len(); i++ {
		out.Append(s.At(i).(T))
	}
	return out
}

// JoinLists is the list of the items of several lists, in order.
func JoinLists[T any](lists ...*List[T]) *List[T] {
	out := NewList[T]()
	for _, l := range lists {
		out.items = append(out.items, l.items...)
	}
	return out
}

// VarTupleOf builds the *args tuple of a call: the extra positional arguments
// written out, then the items of the unpacked list from index from onward.
func VarTupleOf[T any](extra *List[T], tail *List[T], from int64) VarTuple[T] {
	items := append([]T(nil), extra.items...)
	if tail != nil && from < int64(len(tail.items)) {
		items = append(items, tail.items[from:]...)
	}
	return VarTuple[T]{items: items}
}

// SplitKeys separates the keys of a **mapping argument that match no
// parameter. With a **kwargs parameter to take them (collect) they are
// returned; without one, the first is an unexpected keyword argument.
func SplitKeys[V any](d *Dict[string, V], fn string, known, posonly []string, collect bool) *Dict[string, V] {
	extra := NewDict[string, V]()
	isKnown := make(map[string]bool, len(known))
	for _, k := range known {
		isKnown[k] = true
	}
	for _, e := range d.entries {
		if e.dead || isKnown[e.key] {
			continue
		}
		if !collect {
			var asKeyword []string
			for _, o := range d.entries {
				for _, name := range posonly {
					if !o.dead && o.key == name {
						asKeyword = append(asKeyword, name)
					}
				}
			}
			if len(asKeyword) > 0 {
				panic(typeError(MsgPositionalOnlyAsKeyword(fn, asKeyword...)))
			}
			UnexpectedKeyword(fn, e.key, Suggest(e.key, known))
		}
		extra.Set(e.key, e.val)
	}
	return extra
}

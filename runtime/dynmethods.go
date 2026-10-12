package runtime

import (
	"math/bits"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The methods of the built-in types, for objects whose type is known only at
// run time.

var strMethodNames = names("upper lower strip lstrip rstrip split rsplit join replace find rfind index rindex startswith endswith count isdigit isalpha isalnum isspace isupper islower isnumeric isdecimal isidentifier istitle title capitalize swapcase center ljust rjust zfill partition rpartition splitlines casefold removeprefix removesuffix expandtabs format")
var listMethodNames = names("append pop insert remove index count sort reverse extend copy clear")
var dictMethodNames = names("get pop popitem setdefault update copy clear keys values items")
var setMethodNames = names("add remove discard pop union intersection difference symmetric_difference issubset issuperset isdisjoint update intersection_update difference_update symmetric_difference_update copy clear")
var tupleMethodNames = names("count index")
var numMethodNames = names("bit_length is_integer conjugate")

func names(s string) map[string]bool {
	m := map[string]bool{}
	for _, n := range strings.Fields(s) {
		m[n] = true
	}
	return m
}

// builtinAttr reads an attribute of a built-in value: its methods, as functions
// that remember the receiver.
func builtinAttr(obj any, name string) (any, bool) {
	var table map[string]bool
	switch obj.(type) {
	case string:
		table = strMethodNames
	case int64, bool, float64:
		table = numMethodNames
		switch name {
		case "real":
			return obj, true
		case "imag":
			if _, isF := obj.(float64); isF {
				return 0.0, true
			}
			return int64(0), true
		}
	case interface{ PyDynTable() map[string]bool }:
		table = obj.(interface{ PyDynTable() map[string]bool }).PyDynTable()
	case Seq:
		if obj.(Seq).Kind() == "tuple" {
			table = tupleMethodNames
		}
	}
	if table == nil || !table[name] {
		return nil, false
	}
	return func(args ...any) any {
		if r, ok := callMethod(obj, name, args, nil); ok {
			return r
		}
		panic(NoAttribute(TypeName(obj), name))
	}, true
}

// DynCallMethod is obj.name(args..., kw...).
func DynCallMethod(obj any, name string, args []any, kw []KW) any {
	if r, ok := callMethod(obj, name, args, kw); ok {
		return r
	}
	return DynCallKw(DynGetAttr(obj, name), args, kw)
}

// callMethod runs a method the runtime implements itself.
func callMethod(obj any, name string, args []any, kw []KW) (any, bool) {
	obj = Unbox(obj)
	switch x := obj.(type) {
	case string:
		if r, ok := strMethod(x, name, args, kw); ok {
			return r, true
		}
	case int64, bool, float64:
		if r, ok := numMethod(x, name, args); ok {
			return r, true
		}
	case interface {
		DynMethod(name string, args []any, kw []KW) (any, bool)
	}:
		if r, ok := x.DynMethod(name, args, kw); ok {
			return r, true
		}
	case Seq:
		if x.Kind() == "tuple" {
			if r, ok := tupleMethod(x, name, args); ok {
				return r, true
			}
		}
	case *PyError:
		if r, ok := excMethod(x, name, args); ok {
			return r, true
		}
	case *ClassRef:
		if r, ok := classMethod(x, name, args, kw); ok {
			return r, true
		}
	}
	return nil, false
}

func argInt(args []any, i int, def int64) int64 {
	if i < len(args) && args[i] != nil {
		n, ok := intOf(Unbox(args[i]))
		if !ok {
			panic(newError("TypeError", "'"+TypeName(args[i])+"' object cannot be interpreted as an integer"))
		}
		return n
	}
	return def
}

func argStr(args []any, i int, def string) string {
	if i < len(args) && args[i] != nil {
		s, ok := Unbox(args[i]).(string)
		if !ok {
			panic(newError("TypeError", "must be str, not "+TypeName(args[i])))
		}
		return s
	}
	return def
}

func kwValue(kw []KW, name string) (any, bool) {
	for _, k := range kw {
		if k.Name == name {
			return k.Val, true
		}
	}
	return nil, false
}

func countArgs(name string, args []any, min, max int) {
	n := len(args)
	if n >= min && n <= max {
		return
	}
	switch {
	case min == 1 && max == 1:
		panic(newError("TypeError", methodOwner(name)+name+"() takes exactly one argument ("+itoa(n)+" given)"))
	case min == max:
		panic(newError("TypeError", name+" expected "+itoa(min)+" argument"+plural(int64(min), "", "s")+", got "+itoa(n)))
	case n > max:
		panic(newError("TypeError", name+" expected at most "+itoa(max)+" argument"+plural(int64(max), "", "s")+", got "+itoa(n)))
	}
	panic(newError("TypeError", name+" expected at least "+itoa(min)+" argument"+plural(int64(min), "", "s")+", got "+itoa(n)))
}

// methodOwner is the class name CPython puts before the methods that take
// exactly one argument.
func methodOwner(name string) string {
	switch name {
	case "append", "remove", "extend", "count", "index":
		return "list."
	case "join":
		return "str."
	}
	return ""
}

func itoa(n int) string { return Str(int64(n)) }

func strMethod(s, name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "strip", "lstrip", "rstrip":
		countArgs(name, args, 0, 1)
	}
	switch name {
	case "upper":
		return strings.ToUpper(s), true
	case "lower", "casefold":
		return strings.ToLower(s), true
	case "strip":
		if len(args) == 0 || args[0] == nil {
			return StrStrip(s), true
		}
		return StrStripChars(s, argStr(args, 0, "")), true
	case "lstrip":
		if len(args) == 0 || args[0] == nil {
			return strings.TrimLeftFunc(s, isSpace), true
		}
		return strings.TrimLeft(s, argStr(args, 0, "")), true
	case "rstrip":
		if len(args) == 0 || args[0] == nil {
			return strings.TrimRightFunc(s, isSpace), true
		}
		return strings.TrimRight(s, argStr(args, 0, "")), true
	case "split", "rsplit":
		sep, hasSep := "", false
		if len(args) > 0 && args[0] != nil {
			sep, hasSep = argStr(args, 0, ""), true
		} else if v, ok := kwValue(kw, "sep"); ok && v != nil {
			sep, hasSep = v.(string), true
		}
		max := argInt(args, 1, -1)
		if v, ok := kwValue(kw, "maxsplit"); ok {
			max, _ = intOf(v)
		}
		return splitStr(s, sep, hasSep, max, name == "rsplit"), true
	case "join":
		countArgs("join", args, 1, 1)
		return DynJoin(s, args[0]), true
	case "replace":
		if len(args) == 3 {
			return strings.Replace(s, argStr(args, 0, ""), argStr(args, 1, ""), int(argInt(args, 2, -1))), true
		}
		countArgs("replace", args, 2, 3)
		return strings.ReplaceAll(s, argStr(args, 0, ""), argStr(args, 1, "")), true
	case "find", "index", "rfind", "rindex":
		sub := argStr(args, 0, "")
		runes := []rune(s)
		start, end := clampRange(len(runes), args, 1)
		hay := string(runes[start:end])
		var at int
		if name == "find" || name == "index" {
			at = strings.Index(hay, sub)
		} else {
			at = strings.LastIndex(hay, sub)
		}
		if at < 0 {
			if name == "index" || name == "rindex" {
				panic(valueError("substring not found"))
			}
			return int64(-1), true
		}
		return int64(start + utf8.RuneCountInString(hay[:at])), true
	case "startswith", "endswith":
		prefixes := []string{}
		switch p := Unbox(argAt(args, 0)).(type) {
		case string:
			prefixes = append(prefixes, p)
		case Seq:
			for i := int64(0); i < p.Len(); i++ {
				prefixes = append(prefixes, p.At(i).(string))
			}
		default:
			panic(newError("TypeError", name+" first arg must be str or a tuple of str, not "+TypeName(p)))
		}
		runes := []rune(s)
		start, end := clampRange(len(runes), args, 1)
		hay := string(runes[start:end])
		for _, p := range prefixes {
			if (name == "startswith" && strings.HasPrefix(hay, p)) || (name == "endswith" && strings.HasSuffix(hay, p)) {
				return true, true
			}
		}
		return false, true
	case "count":
		sub := argStr(args, 0, "")
		if sub == "" {
			return StrLen(s) + 1, true
		}
		return int64(strings.Count(s, sub)), true
	case "isdigit", "isnumeric", "isdecimal":
		return s != "" && allRunes(s, unicode.IsDigit), true
	case "isalpha":
		return s != "" && allRunes(s, unicode.IsLetter), true
	case "isalnum":
		return s != "" && allRunes(s, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }), true
	case "isspace":
		return s != "" && allRunes(s, isSpace), true
	case "isupper":
		return strings.ToUpper(s) == s && strings.ToLower(s) != s, true
	case "islower":
		return strings.ToLower(s) == s && strings.ToUpper(s) != s, true
	case "isidentifier":
		return isIdentifier(s), true
	case "istitle":
		return s != "" && titleCase(s) == s && strings.ToLower(s) != s, true
	case "title":
		return titleCase(s), true
	case "capitalize":
		if s == "" {
			return s, true
		}
		r, n := utf8.DecodeRuneInString(s)
		return string(unicode.ToUpper(r)) + strings.ToLower(s[n:]), true
	case "swapcase":
		return strings.Map(func(r rune) rune {
			if unicode.IsUpper(r) {
				return unicode.ToLower(r)
			}
			return unicode.ToUpper(r)
		}, s), true
	case "center", "ljust", "rjust":
		w := int(argInt(args, 0, 0))
		fill := argStr(args, 1, " ")
		n := utf8.RuneCountInString(s)
		if w <= n {
			return s, true
		}
		pad := w - n
		switch name {
		case "ljust":
			return s + strings.Repeat(fill, pad), true
		case "rjust":
			return strings.Repeat(fill, pad) + s, true
		}
		left := pad/2 + (pad & w & 1)
		return strings.Repeat(fill, left) + s + strings.Repeat(fill, pad-left), true
	case "zfill":
		w := int(argInt(args, 0, 0))
		n := utf8.RuneCountInString(s)
		if w <= n {
			return s, true
		}
		sign := ""
		body := s
		if s != "" && (s[0] == '-' || s[0] == '+') {
			sign, body = s[:1], s[1:]
		}
		return sign + strings.Repeat("0", w-n) + body, true
	case "partition", "rpartition":
		sep := argStr(args, 0, "")
		if sep == "" {
			panic(valueError("empty separator"))
		}
		var i int
		if name == "partition" {
			i = strings.Index(s, sep)
		} else {
			i = strings.LastIndex(s, sep)
		}
		if i < 0 {
			if name == "partition" {
				return Tup3(s, "", ""), true
			}
			return Tup3("", "", s), true
		}
		return Tup3(s[:i], sep, s[i+len(sep):]), true
	case "splitlines":
		keep := false
		if v, ok := kwValue(kw, "keepends"); ok {
			keep = Truth(v)
		}
		out := NewList[string]()
		for len(s) > 0 {
			i := strings.IndexAny(s, "\n\r")
			if i < 0 {
				out.Append(s)
				break
			}
			j := i + 1
			if s[i] == '\r' && j < len(s) && s[j] == '\n' {
				j++
			}
			if keep {
				out.Append(s[:j])
			} else {
				out.Append(s[:i])
			}
			s = s[j:]
		}
		return out, true
	case "removeprefix":
		return strings.TrimPrefix(s, argStr(args, 0, "")), true
	case "removesuffix":
		return strings.TrimSuffix(s, argStr(args, 0, "")), true
	case "format":
		return strFormat(s, args, kw), true
	case "expandtabs":
		return strings.ReplaceAll(s, "\t", strings.Repeat(" ", int(argInt(args, 0, 8)))), true
	}
	return nil, false
}

func argAt(args []any, i int) any {
	if i < len(args) {
		return args[i]
	}
	return nil
}

func clampRange(n int, args []any, from int) (int, int) {
	start, end := int(argInt(args, from, 0)), int(argInt(args, from+1, int64(n)))
	if start < 0 {
		start = max(start+n, 0)
	}
	if end < 0 {
		end = max(end+n, 0)
	}
	start, end = min(start, n), min(end, n)
	if start > end {
		start = end
	}
	return start, end
}

func allRunes(s string, f func(rune) bool) bool {
	for _, r := range s {
		if !f(r) {
			return false
		}
	}
	return true
}

func isIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if !(r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r))) {
			return false
		}
	}
	return true
}

func titleCase(s string) string {
	var b strings.Builder
	prevLetter := false
	for _, r := range s {
		switch {
		case !unicode.IsLetter(r):
			b.WriteRune(r)
			prevLetter = false
		case prevLetter:
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteRune(unicode.ToUpper(r))
			prevLetter = true
		}
	}
	return b.String()
}

func splitStr(s, sep string, hasSep bool, max int64, right bool) *List[string] {
	out := NewList[string]()
	if !hasSep {
		fields := strings.FieldsFunc(s, isSpace)
		if max >= 0 && int64(len(fields)) > max+1 {
			if right {
				head := strings.TrimRightFunc(s, isSpace)
				var parts []string
				rest := head
				for i := int64(0); i < max; i++ {
					j := strings.LastIndexFunc(rest, isSpace)
					if j < 0 {
						break
					}
					parts = append([]string{rest[j+1:]}, parts...)
					rest = strings.TrimRightFunc(rest[:j], isSpace)
				}
				parts = append([]string{rest}, parts...)
				out.items = parts
				return out
			}
			var parts []string
			rest := strings.TrimLeftFunc(s, isSpace)
			for i := int64(0); i < max; i++ {
				j := strings.IndexFunc(rest, isSpace)
				if j < 0 {
					break
				}
				parts = append(parts, rest[:j])
				rest = strings.TrimLeftFunc(rest[j:], isSpace)
			}
			parts = append(parts, rest)
			out.items = parts
			return out
		}
		out.items = fields
		return out
	}
	if sep == "" {
		panic(valueError("empty separator"))
	}
	if max < 0 {
		out.items = strings.Split(s, sep)
		return out
	}
	if right {
		var parts []string
		rest := s
		for i := int64(0); i < max; i++ {
			j := strings.LastIndex(rest, sep)
			if j < 0 {
				break
			}
			parts = append([]string{rest[j+len(sep):]}, parts...)
			rest = rest[:j]
		}
		out.items = append([]string{rest}, parts...)
		return out
	}
	out.items = strings.SplitN(s, sep, int(max)+1)
	return out
}

func numMethod(x any, name string, args []any) (any, bool) {
	switch name {
	case "bit_length":
		if n, ok := intOf(x); ok {
			if n < 0 {
				n = -n
			}
			return int64(bits.Len64(uint64(n))), true
		}
	case "is_integer":
		if f, ok := x.(float64); ok {
			return f == float64(int64(f)), true
		}
	case "conjugate":
		return x, true
	}
	return nil, false
}

func tupleMethod(s Seq, name string, args []any) (any, bool) {
	switch name {
	case "count":
		var n int64
		for i := int64(0); i < s.Len(); i++ {
			if Eq(s.At(i), argAt(args, 0)) {
				n++
			}
		}
		return n, true
	case "index":
		for i := int64(0); i < s.Len(); i++ {
			if Eq(s.At(i), argAt(args, 0)) {
				return i, true
			}
		}
		panic(valueError("tuple.index(x): x not in tuple"))
	}
	return nil, false
}

func excMethod(e *PyError, name string, args []any) (any, bool) {
	switch name {
	case "with_traceback":
		return e, true
	case "add_note":
		return nil, true
	}
	return nil, false
}

// ---- lists ----

func (l *List[T]) PyDynTable() map[string]bool { return listMethodNames }

func (l *List[T]) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "append":
		countArgs("append", args, 1, 1)
		l.Append(conv[T](args[0]))
		return nil, true
	case "pop":
		countArgs("pop", args, 0, 1)
		if len(args) == 0 {
			return l.Pop(), true
		}
		return l.PopAt(argInt(args, 0, -1)), true
	case "insert":
		countArgs("insert", args, 2, 2)
		l.Insert(argInt(args, 0, 0), conv[T](args[1]))
		return nil, true
	case "remove":
		countArgs("remove", args, 1, 1)
		l.Remove(conv[T](args[0]))
		return nil, true
	case "index":
		countArgs("index", args, 1, 1)
		v, ok := convOpt[T](args[0])
		if !ok {
			panic(valueError(Repr(args[0]) + " is not in list"))
		}
		return l.Index(v), true
	case "count":
		countArgs("count", args, 1, 1)
		v, ok := convOpt[T](args[0])
		if !ok {
			return int64(0), true
		}
		return l.Count(v), true
	case "sort":
		if len(args) > 0 {
			panic(newError("TypeError", "sort() takes no positional arguments"))
		}
		reverse := false
		var key any
		for _, k := range kw {
			switch k.Name {
			case "reverse":
				reverse = Truth(Unbox(k.Val))
			case "key":
				key = k.Val
			default:
				panic(newError("TypeError", "sort() got an unexpected keyword argument '"+k.Name+"'"))
			}
		}
		if key == nil || IsNone(key) {
			l.Sort(reverse)
		} else {
			SortKey(l, func(v T) any { return DynCall(key, v) }, reverse)
		}
		return nil, true
	case "reverse":
		l.Reverse()
		return nil, true
	case "extend":
		countArgs("extend", args, 1, 1)
		for c := DynIter(args[0]).Iter(); c.Next(); {
			l.Append(conv[T](c.Value()))
		}
		return nil, true
	case "copy":
		return l.Copy(), true
	case "clear":
		l.Clear()
		return nil, true
	}
	return nil, false
}

// ---- dicts ----

func (d *Dict[K, V]) PyDynTable() map[string]bool { return dictMethodNames }

func (d *Dict[K, V]) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "get":
		countArgs("get", args, 1, 2)
		k, ok := convOpt[K](args[0])
		if ok && d.Has(k) {
			return d.Get(k), true
		}
		if len(args) == 2 {
			return args[1], true
		}
		return nil, true
	case "pop":
		countArgs("pop", args, 1, 2)
		k, ok := convOpt[K](args[0])
		if ok && d.Has(k) {
			return d.Pop(k), true
		}
		if len(args) == 2 {
			return args[1], true
		}
		panic(NewError("KeyError", args[0]))
	case "popitem":
		return d.PopItem(), true
	case "setdefault":
		countArgs("setdefault", args, 1, 2)
		k := conv[K](args[0])
		var def V
		if len(args) == 2 {
			def = conv[V](args[1])
		}
		return d.SetDefault(k, def), true
	case "update":
		if len(args) == 1 {
			src := Unbox(args[0])
			if o, ok := src.(*Dict[K, V]); ok {
				d.Update(o)
			} else if m, ok := src.(interface{ IterAny() Cursor[any] }); ok {
				if g, isDict := src.(interface{ GetItemAny(any) any }); isDict {
					for c := m.IterAny(); c.Next(); {
						d.Set(conv[K](c.Value()), conv[V](g.GetItemAny(c.Value())))
					}
				} else {
					for c := DynIter(src).Iter(); c.Next(); {
						pair := Unbox(c.Value()).(Seq)
						d.Set(conv[K](pair.At(0)), conv[V](pair.At(1)))
					}
				}
			}
		}
		for _, k := range kw {
			d.Set(conv[K](k.Name), conv[V](k.Val))
		}
		return nil, true
	case "copy":
		return d.Copy(), true
	case "clear":
		d.Clear()
		return nil, true
	case "keys":
		return d.Keys(), true
	case "values":
		return d.Values(), true
	case "items":
		return d.Items(), true
	}
	return nil, false
}

// ---- sets ----

func (s *Set[T]) PyDynTable() map[string]bool { return setMethodNames }

func (s *Set[T]) DynMethod(name string, args []any, kw []KW) (any, bool) {
	others := func() []Iterable[T] {
		var out []Iterable[T]
		for _, a := range args {
			l := NewList[T]()
			for c := DynIter(a).Iter(); c.Next(); {
				l.Append(conv[T](c.Value()))
			}
			out = append(out, l)
		}
		return out
	}
	one := func() Iterable[T] {
		countArgs(name, args, 1, 1)
		return others()[0]
	}
	switch name {
	case "add":
		countArgs("add", args, 1, 1)
		s.Add(conv[T](args[0]))
		return nil, true
	case "remove":
		countArgs("remove", args, 1, 1)
		s.Remove(conv[T](args[0]))
		return nil, true
	case "discard":
		countArgs("discard", args, 1, 1)
		if v, ok := convOpt[T](args[0]); ok {
			s.Discard(v)
		}
		return nil, true
	case "pop":
		return s.Pop(), true
	case "union":
		return s.Union(others()...), true
	case "intersection":
		return s.Intersection(others()...), true
	case "difference":
		return s.Difference(others()...), true
	case "symmetric_difference":
		return s.SymmetricDifference(one()), true
	case "issubset":
		return s.IsSubset(one()), true
	case "issuperset":
		return s.IsSuperset(one()), true
	case "isdisjoint":
		return s.IsDisjoint(one()), true
	case "update":
		s.Update(others()...)
		return nil, true
	case "intersection_update":
		s.IntersectionUpdate(others()...)
		return nil, true
	case "difference_update":
		s.DifferenceUpdate(others()...)
		return nil, true
	case "symmetric_difference_update":
		s.SymmetricDifferenceUpdate(one())
		return nil, true
	case "copy":
		return s.Copy(), true
	case "clear":
		s.Clear()
		return nil, true
	}
	return nil, false
}

// InplaceBinary implements list += iterable and list *= int.
func (l *List[T]) InplaceBinary(op string, b any) bool {
	switch op {
	case "Add":
		items := []T{}
		for c := DynIter(b).Iter(); c.Next(); {
			items = append(items, conv[T](c.Value()))
		}
		l.items = append(l.items, items...)
		return true
	case "Mult":
		n, ok := intOf(b)
		if !ok {
			return false
		}
		if n <= 0 {
			l.items = l.items[:0]
			return true
		}
		orig := append([]T{}, l.items...)
		for i := int64(1); i < n; i++ {
			l.items = append(l.items, orig...)
		}
		return true
	}
	return false
}

// ---- generators and iterators ----

func (g *Gen[Y, S, R]) PyDynTable() map[string]bool { return genMethodNames }
func (g *Gen[Y, S, R]) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "send":
		countArgs("send", args, 1, 1)
		return g.Send(conv[S](args[0])), true
	case "throw":
		countArgs("throw", args, 1, 1)
		e, ok := Unbox(args[0]).(*PyError)
		if !ok {
			if x, isInst := Unbox(args[0]).(interface{ PyExc() *PyError }); isInst {
				e = x.PyExc()
			} else {
				panic(newError("TypeError", "exceptions must derive from BaseException"))
			}
		}
		return g.Throw(e), true
	case "close":
		g.Close()
		return nil, true
	case "__next__":
		return g.NextValue(), true
	}
	return nil, false
}

var genMethodNames = names("send throw close __next__")
var iterMethodNames = names("__next__")

func (it *Iterator[T]) PyDynTable() map[string]bool { return iterMethodNames }
func (it *Iterator[T]) DynMethod(name string, args []any, kw []KW) (any, bool) {
	if name == "__next__" {
		return it.NextValue(), true
	}
	return nil, false
}

// classMethod calls a method of a built-in type itself, such as dict.fromkeys.
func classMethod(c *ClassRef, name string, args []any, kw []KW) (any, bool) {
	if !c.Builtin {
		return nil, false
	}
	if c.Name == "dict" && name == "fromkeys" {
		countArgs("fromkeys", args, 1, 2)
		out := NewDict[any, any]()
		var val any
		if len(args) == 2 {
			val = args[1]
		}
		for it := DynIter(args[0]).Iter(); it.Next(); {
			out.Set(it.Value(), val)
		}
		return out, true
	}
	return nil, false
}

// strFormat is str.format: "{}", "{0}", "{name}", "{x.attr}", "{x[0]}", with
// a conversion (!r, !s) and a format spec, which may itself hold fields.
func strFormat(format string, args []any, kw []KW) string {
	var out strings.Builder
	auto := 0
	lookup := func(field string) any {
		head, rest := field, ""
		if k := strings.IndexAny(field, ".["); k >= 0 {
			head, rest = field[:k], field[k:]
		}
		var v any
		switch {
		case head == "":
			if auto >= len(args) {
				panic(newError("IndexError", "Replacement index "+itoa(auto)+" out of range for positional args tuple"))
			}
			v = args[auto]
			auto++
		case head[0] >= '0' && head[0] <= '9':
			n, _ := strconv.Atoi(head)
			if n >= len(args) {
				panic(newError("IndexError", "Replacement index "+itoa(n)+" out of range for positional args tuple"))
			}
			v = args[n]
		default:
			found := false
			for _, k := range kw {
				if k.Name == head {
					v, found = k.Val, true
				}
			}
			if !found {
				panic(NewError("KeyError", head))
			}
		}
		for rest != "" {
			switch rest[0] {
			case '.':
				end := strings.IndexAny(rest[1:], ".[")
				name := rest[1:]
				if end >= 0 {
					name, rest = rest[1:1+end], rest[1+end:]
				} else {
					rest = ""
				}
				v = DynGetAttr(v, name)
			case '[':
				end := strings.IndexByte(rest, ']')
				if end < 0 {
					panic(valueError("Missing ']' in format string"))
				}
				key := rest[1:end]
				rest = rest[end+1:]
				if n, err := strconv.Atoi(key); err == nil {
					v = DynGetItem(v, int64(n))
				} else {
					v = DynGetItem(v, key)
				}
			default:
				panic(valueError("Only '.' or '[' may follow ']' in format field specifier"))
			}
		}
		return v
	}
	var render func(f string) string
	render = func(f string) string {
		var out strings.Builder
		for i := 0; i < len(f); i++ {
			c := f[i]
			switch {
			case c == '{' && i+1 < len(f) && f[i+1] == '{':
				out.WriteByte('{')
				i++
			case c == '}' && i+1 < len(f) && f[i+1] == '}':
				out.WriteByte('}')
				i++
			case c == '}':
				panic(valueError("Single '}' encountered in format string"))
			case c == '{':
				depth, j := 1, i+1
				for ; j < len(f) && depth > 0; j++ {
					if f[j] == '{' {
						depth++
					} else if f[j] == '}' {
						depth--
					}
				}
				if depth != 0 {
					panic(valueError("Single '{' encountered in format string"))
				}
				field := f[i+1 : j-1]
				i = j - 1
				spec := ""
				if k := strings.IndexByte(field, ':'); k >= 0 && !strings.Contains(field[:k], "[") {
					field, spec = field[:k], field[k+1:]
				} else if k >= 0 {
					// a colon inside [..] belongs to the key
					if close := strings.IndexByte(field, ']'); close >= 0 && k > close {
						field, spec = field[:k], field[k+1:]
					}
				}
				conv := ""
				if k := strings.IndexByte(field, '!'); k >= 0 {
					field, conv = field[:k], field[k+1:]
				}
				v := lookup(field)
				switch conv {
				case "r":
					v = Repr(Unbox(v))
				case "s":
					v = Str(Unbox(v))
				case "a":
					v = Repr(Unbox(v))
				case "":
				default:
					panic(valueError("Unknown conversion specifier " + conv))
				}
				out.WriteString(FormatDyn(v, render(spec)))
			default:
				out.WriteByte(c)
			}
		}
		return out.String()
	}
	out.WriteString(render(format))
	return out.String()
}

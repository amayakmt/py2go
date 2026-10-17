package runtime

import (
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The methods of the built-in types, for objects whose type is known only at
// run time.

var strMethodNames = names("encode translate isprintable isascii format_map upper lower strip lstrip rstrip split rsplit join replace find rfind index rindex startswith endswith count isdigit isalpha isalnum isspace isupper islower isnumeric isdecimal isidentifier istitle title capitalize swapcase center ljust rjust zfill partition rpartition splitlines casefold removeprefix removesuffix expandtabs format")
var listMethodNames = names("append pop insert remove index count sort reverse extend copy clear")
var dictMethodNames = names("get pop popitem setdefault update copy clear keys values items")
var setMethodNames = names("add remove discard pop union intersection difference symmetric_difference issubset issuperset isdisjoint update intersection_update difference_update symmetric_difference_update copy clear")
var dunderNames = names("__getitem__ __setitem__ __delitem__ __len__ __contains__ __iter__ __eq__ __ne__ __str__ __repr__ __bool__")
var tupleMethodNames = names("count index")
var numMethodNames = names("bit_length bit_count is_integer conjugate as_integer_ratio to_bytes hex fromhex")

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
	case complex128:
		if v, ok := complexAttr(obj.(complex128), name); ok {
			return v, true
		}
		return nil, false
	case int64, bool, float64, *big.Int:
		table = numMethodNames
		switch name {
		case "real", "numerator":
			if name == "numerator" && !isIntVal(obj) {
				break
			}
			if b, isBool := obj.(bool); isBool {
				return b2i(b), true
			}
			return obj, true
		case "imag":
			if _, isF := obj.(float64); isF {
				return 0.0, true
			}
			return int64(0), true
		case "denominator":
			if isIntVal(obj) {
				return int64(1), true
			}
		}
	case interface{ PyDynTable() map[string]bool }:
		table = obj.(interface{ PyDynTable() map[string]bool }).PyDynTable()
	case Seq:
		if obj.(Seq).Kind() == "tuple" {
			table = tupleMethodNames
		}
	}
	if table == nil || !table[name] {
		if _, user := obj.(Named); !user && dunderNames[name] && obj != nil {
			return func(args ...any) any {
				if r, ok := dunderMethod(obj, name, args); ok {
					return r
				}
				panic(NoAttribute(TypeName(obj), name))
			}, true
		}
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
	if _, user := Unbox(obj).(Named); !user {
		if r, ok := dunderMethod(obj, name, args); ok {
			return r
		}
	}
	return DynCallKw(DynGetAttr(obj, name), args, kw)
}

// dunderMethod is obj.__getitem__(k) and the other special methods that
// built-in objects answer by running the operation.
func dunderMethod(obj any, name string, args []any) (any, bool) {
	n := func(k int) {
		if len(args) != k {
			panic(newError("TypeError", "expected "+itoa(k)+" argument"+map[bool]string{true: "", false: "s"}[k == 1]+", got "+itoa(len(args))))
		}
	}
	switch name {
	case "__getitem__":
		n(1)
		return DynGetItem(obj, args[0]), true
	case "__setitem__":
		n(2)
		DynSetItem(obj, args[0], args[1])
		return nil, true
	case "__delitem__":
		n(1)
		DynDelItem(obj, args[0])
		return nil, true
	case "__len__":
		n(0)
		return DynLen(obj), true
	case "__contains__":
		n(1)
		return DynContains(obj, args[0]), true
	case "__iter__":
		n(0)
		return DynIter(obj), true
	case "__eq__":
		n(1)
		return Eq(obj, args[0]), true
	case "__ne__":
		n(1)
		return !Eq(obj, args[0]), true
	case "__str__":
		n(0)
		return Str(obj), true
	case "__repr__":
		n(0)
		return Repr(obj), true
	case "__bool__":
		n(0)
		return DynBool(obj), true
	}
	return nil, false
}

// callMethod runs a method the runtime implements itself.
func callMethod(obj any, name string, args []any, kw []KW) (any, bool) {
	obj = Unbox(obj)
	switch x := obj.(type) {
	case string:
		if r, ok := strMethod(x, name, args, kw); ok {
			return r, true
		}
	case int64, bool, float64, *big.Int:
		if r, ok := numMethod(x, name, args, kw); ok {
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
	if i < len(args) && !IsNone(args[i]) {
		n, ok := intOf(Unbox(args[i]))
		if !ok {
			panic(newError("TypeError", "'"+TypeName(args[i])+"' object cannot be interpreted as an integer"))
		}
		return n
	}
	return def
}

func argStr(args []any, i int, def string) string {
	if i < len(args) && !IsNone(args[i]) {
		s, ok := BaseValue(Unbox(args[i])).(string)
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
	case "center", "ljust", "rjust":
		countArgsClinic(name, args, 1, 2)
	case "zfill":
		countArgsClinic(name, args, 1, 1)
	case "replace":
		if len(args) < 2 {
			panic(newError("TypeError", "replace() takes at least 2 positional arguments ("+itoa(len(args))+" given)"))
		}
	case "count", "find", "rfind", "index", "rindex":
		if len(args) == 0 {
			panic(newError("TypeError", name+"() takes at least 1 argument (0 given)"))
		}
	}
	// A str argument, with the message of the method when it is not one.
	argStr := func(args []any, i int, def string) string {
		if i >= len(args) || IsNone(args[i]) && name != "split" && name != "rsplit" {
			return def
		}
		if IsNone(args[i]) {
			return def
		}
		s, ok := Unbox(args[i]).(string)
		if !ok {
			what := "must be str, not "
			switch name {
			case "strip", "lstrip", "rstrip":
				panic(newError("TypeError", name+" arg must be None or str"))
			case "split", "rsplit":
				panic(newError("TypeError", "must be str or None, not "+TypeName(Unbox(args[i]))))
			case "removeprefix", "removesuffix":
				panic(newError("TypeError", name+"() argument must be str, not "+TypeName(Unbox(args[i]))))
			}
			panic(newError("TypeError", name+"() argument "+itoa(i+1)+" "+what+TypeName(Unbox(args[i]))))
		}
		return s
	}
	switch name {
	case "upper":
		return FullUpper(s), true
	case "lower":
		return FullLower(s), true
	case "casefold":
		return FullCasefold(s), true
	case "strip":
		if len(args) == 0 || IsNone(args[0]) {
			return StrStrip(s), true
		}
		return StrStripChars(s, argStr(args, 0, "")), true
	case "lstrip":
		if len(args) == 0 || IsNone(args[0]) {
			return strings.TrimLeftFunc(s, isSpace), true
		}
		return strings.TrimLeft(s, argStr(args, 0, "")), true
	case "rstrip":
		if len(args) == 0 || IsNone(args[0]) {
			return strings.TrimRightFunc(s, isSpace), true
		}
		return strings.TrimRight(s, argStr(args, 0, "")), true
	case "split", "rsplit":
		sep, hasSep := "", false
		if len(args) > 0 && !IsNone(args[0]) {
			sep, hasSep = argStr(args, 0, ""), true
		} else if v, ok := kwValue(kw, "sep"); ok && v != nil {
			sep, hasSep = v.(string), true
		}
		max := argInt(args, 1, -1)
		if v, ok := kwValue(kw, "maxsplit"); ok {
			max = argInt([]any{v}, 0, -1)
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
		start, end, found := searchRange(len(runes), args, 1)
		if !found {
			if name == "index" || name == "rindex" {
				panic(valueError("substring not found"))
			}
			return int64(-1), true
		}
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
		runes := []rune(s)
		start, end, found := searchRange(len(runes), args, 1)
		matches := func(p string) bool {
			if !found {
				return false
			}
			hay := string(runes[start:end])
			return (name == "startswith" && strings.HasPrefix(hay, p)) || (name == "endswith" && strings.HasSuffix(hay, p))
		}
		switch p := Unbox(argAt(args, 0)).(type) {
		case string:
			return matches(p), true
		case Seq:
			if p.Kind() != "tuple" {
				panic(newError("TypeError", name+" first arg must be str or a tuple of str, not "+TypeName(p)))
			}
			for i := int64(0); i < p.Len(); i++ {
				ps, isStr := Unbox(p.At(i)).(string)
				if !isStr {
					panic(newError("TypeError", "tuple for "+name+" must only contain str, not "+TypeName(Unbox(p.At(i)))))
				}
				if matches(ps) {
					return true, true
				}
			}
			return false, true
		default:
			panic(newError("TypeError", name+" first arg must be str or a tuple of str, not "+TypeName(p)))
		}
	case "count":
		sub := argStr(args, 0, "")
		runes := []rune(s)
		start, end, found := searchRange(len(runes), args, 1)
		if !found {
			return int64(0), true
		}
		if sub == "" {
			return int64(end-start) + 1, true
		}
		return int64(strings.Count(string(runes[start:end]), sub)), true
	case "isdecimal":
		return s != "" && allRunes(s, decimalRanges.has), true
	case "isdigit":
		return s != "" && allRunes(s, digitRanges.has), true
	case "isnumeric":
		return s != "" && allRunes(s, numericRanges.has), true
	case "isalpha":
		return s != "" && allRunes(s, alphaRanges.has), true
	case "isalnum":
		return s != "" && allRunes(s, func(r rune) bool {
			return alphaRanges.has(r) || decimalRanges.has(r) || digitRanges.has(r) || numericRanges.has(r)
		}), true
	case "isspace":
		return s != "" && allRunes(s, spaceRanges.has), true
	case "isupper":
		return StrIsUpper(s), true
	case "islower":
		return StrIsLower(s), true
	case "isidentifier":
		return isIdentifier(s), true
	case "istitle":
		return StrIsTitle(s), true
	case "title":
		return FullTitle(s), true
	case "capitalize":
		return FullCapitalize(s), true
	case "swapcase":
		return FullSwapcase(s), true
	case "center", "ljust", "rjust":
		w := int(argInt(args, 0, 0))
		fill := " "
		if len(args) > 1 {
			f, ok := Unbox(args[1]).(string)
			if !ok {
				panic(newError("TypeError", "The fill character must be a unicode character, not "+TypeName(Unbox(args[1]))))
			}
			if utf8.RuneCountInString(f) != 1 {
				panic(newError("TypeError", "The fill character must be exactly one character long"))
			}
			fill = f
		}
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
		if len(args) > 0 {
			keep = Truth(Unbox(args[0]))
		}
		if v, ok := kwValue(kw, "keepends"); ok {
			keep = Truth(Unbox(v))
		}
		out := NewList[string]()
		for len(s) > 0 {
			i, n := lineBreak(s)
			if i < 0 {
				out.Append(s)
				break
			}
			if keep {
				out.Append(s[:i+n])
			} else {
				out.Append(s[:i])
			}
			s = s[i+n:]
		}
		return out, true
	case "removeprefix":
		return strings.TrimPrefix(s, argStr(args, 0, "")), true
	case "removesuffix":
		return strings.TrimSuffix(s, argStr(args, 0, "")), true
	case "format":
		return strFormat(s, args, kw), true
	case "format_map":
		countArgs("format_map", args, 1, 1)
		return strFormatMap(s, args[0]), true
	case "encode":
		enc, errs := argNamed("encode", "encoding", args, 0, "utf-8"), argNamed("encode", "errors", args, 1, "strict")
		if v, ok := kwValue(kw, "encoding"); ok {
			enc = Unbox(v).(string)
		}
		if v, ok := kwValue(kw, "errors"); ok {
			errs = Unbox(v).(string)
		}
		return EncodeStr(s, enc, errs), true
	case "translate":
		countArgs("translate", args, 1, 1)
		return strTranslate(s, args[0]), true
	case "isprintable":
		return allRunes(s, printableRanges.has), true
	case "isascii":
		return isASCII(s), true
	case "expandtabs":
		return strExpandTabs(s, int(argInt(args, 0, 8))), true
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
	start, end, _ := searchRange(n, args, from)
	return start, end
}

// argIdx reads a slice index argument: an int or None.
func argIdx(args []any, i int, def int64) int64 {
	if i < len(args) && !IsNone(args[i]) {
		v := Unbox(args[i])
		if bi, isBig := v.(*big.Int); isBig {
			if bi.Sign() < 0 {
				return -1 << 62
			}
			return 1 << 62
		}
		n, ok := intOf(v)
		if !ok {
			panic(newError("TypeError", "slice indices must be integers or None or have an __index__ method"))
		}
		return n
	}
	return def
}

// searchRange is the start and end that find, count and startswith take, out
// of a sequence of n items. ok is false when start lies beyond end, where
// nothing is found, not even an empty string.
func searchRange(n int, args []any, from int) (start, end int, ok bool) {
	st, en := argIdx(args, from, 0), argIdx(args, from+1, int64(n))
	if st < 0 {
		st = max(st+int64(n), 0)
	}
	if en < 0 {
		en = max(en+int64(n), 0)
	}
	en = min(en, int64(n))
	if st > en {
		return int(min(st, int64(n))), int(en), false
	}
	return int(st), int(en), true
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
		if i == 0 && !identStartRanges.has(r) {
			return false
		}
		if i > 0 && !identContinueRanges.has(r) {
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
			b.WriteString(fullTitleRune(r))
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

func numMethod(x any, name string, args []any, kw []KW) (any, bool) {
	if f, ok := x.(float64); ok {
		return floatMethod(f, name, args)
	}
	b, ok := bigOf(x)
	if !ok {
		return nil, false
	}
	switch name {
	case "bit_length", "bit_count", "is_integer", "conjugate", "as_integer_ratio", "__abs__":
		if len(args) > 0 {
			panic(newError("TypeError", "int."+name+"() takes no arguments ("+itoa(len(args))+" given)"))
		}
	}
	switch name {
	case "bit_length":
		return int64(new(big.Int).Abs(b).BitLen()), true
	case "bit_count":
		n := 0
		for _, w := range new(big.Int).Abs(b).Bits() {
			n += bits.OnesCount(uint(w))
		}
		return int64(n), true
	case "is_integer":
		return true, true
	case "conjugate", "__index__", "__int__", "__trunc__", "__floor__", "__ceil__":
		return NormBig(b), true
	case "as_integer_ratio":
		return NewVarTuple[any](NormBig(b), int64(1)), true
	case "to_bytes":
		return intToBytes(b, args, kw), true
	case "__abs__":
		return NormBig(new(big.Int).Abs(b)), true
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
		for i, x := range l.items {
			if Eq(x, args[0]) {
				l.items = append(l.items[:i], l.items[i+1:]...)
				return nil, true
			}
		}
		panic(valueError("list.remove(x): x not in list"))
	case "index":
		countArgs("index", args, 1, 3)
		start, end := clampRange(len(l.items), args, 1)
		for i := start; i < end; i++ {
			if Eq(l.items[i], args[0]) {
				return int64(i), true
			}
		}
		panic(valueError(Repr(Unbox(args[0])) + " is not in list"))
	case "count":
		countArgs("count", args, 1, 1)
		var n int64
		for _, x := range l.items {
			if Eq(x, args[0]) {
				n++
			}
		}
		return n, true
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
			} else if keys, isMap := userMappingKeys(src); isMap {
				for _, k := range keys {
					d.Set(conv[K](k), conv[V](DynGetItem(src, k)))
				}
			} else if m, ok := src.(interface{ IterAny() Cursor[any] }); ok {
				_, isSeq := src.(Seq)
				if g, isDict := src.(interface{ GetItemAny(any) any }); isDict && !isSeq {
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
	switch {
	case c.Name == "int" && name == "from_bytes":
		return IntFromBytes(args, kw), true
	case (c.Name == "bytes" || c.Name == "bytearray") && name == "fromhex":
		countArgs("fromhex", args, 1, 1)
		s, pos, ok := FromHex(Unbox(args[0]).(string))
		if !ok {
			panic(valueError("non-hexadecimal number found in fromhex() arg at position " + itoa(pos)))
		}
		if c.Name == "bytearray" {
			return &ByteArray{b: []byte(s)}, true
		}
		return Bytes(s), true
	case (c.Name == "bytes" || c.Name == "bytearray") && name == "maketrans":
		countArgs("maketrans", args, 2, 2)
		from, to := needBytes(args[0], name), needBytes(args[1], name)
		if len(from) != len(to) {
			panic(valueError("maketrans arguments must have same length"))
		}
		tbl := make([]byte, 256)
		for i := range tbl {
			tbl[i] = byte(i)
		}
		for i := 0; i < len(from); i++ {
			tbl[from[i]] = to[i]
		}
		return Bytes(tbl), true
	case c.Name == "float" && name == "fromhex":
		countArgs("fromhex", args, 1, 1)
		return FloatFromHex(Unbox(args[0]).(string)), true
	case c.Name == "str" && name == "maketrans":
		return strMaketrans(args), true
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
					v = Ascii(Unbox(v))
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

// strMaketrans is str.maketrans(x[, y[, z]]): a dict from code points to
// code points, strings or None.
func strMaketrans(args []any) any {
	out := NewDict[any, any]()
	switch len(args) {
	case 1:
		d, ok := Unbox(args[0]).(interface{ IterAny() Cursor[any] })
		get, isDict := Unbox(args[0]).(interface{ GetItemAny(any) any })
		if !ok || !isDict {
			panic(newError("TypeError", "if you give only one argument to maketrans it must be a dict"))
		}
		for c := d.IterAny(); c.Next(); {
			k := Unbox(c.Value())
			if s, isStr := k.(string); isStr {
				r := []rune(s)
				if len(r) != 1 {
					panic(valueError("string keys in translate table must be of length 1"))
				}
				k = int64(r[0])
			}
			out.Set(k, get.GetItemAny(c.Value()))
		}
	case 2, 3:
		str := func(i int) string {
			v, ok := Unbox(args[i]).(string)
			if !ok {
				if i == 0 {
					panic(newError("TypeError", "first maketrans argument must be a string if there is a second argument"))
				}
				panic(newError("TypeError", "maketrans() argument "+itoa(i+1)+" must be str, not "+TypeName(Unbox(args[i]))))
			}
			return v
		}
		_ = str(1)
		from, to := str(0), str(1)
		fr, tr := []rune(from), []rune(to)
		if len(fr) != len(tr) {
			panic(valueError("the first two maketrans arguments must have equal length"))
		}
		for i := range fr {
			out.Set(int64(fr[i]), int64(tr[i]))
		}
		if len(args) == 3 {
			for _, r := range str(2) {
				out.Set(int64(r), nil)
			}
		}
	default:
		panic(newError("TypeError", "maketrans expected 1 to 3 arguments"))
	}
	return out
}

func strExpandTabs(s string, size int) string {
	var b strings.Builder
	col := 0
	for _, r := range s {
		switch r {
		case '\t':
			if size > 0 {
				n := size - col%size
				b.WriteString(strings.Repeat(" ", n))
				col += n
			}
		case '\n', '\r':
			b.WriteRune(r)
			col = 0
		default:
			b.WriteRune(r)
			col++
		}
	}
	return b.String()
}

// strTranslate is s.translate(table): the table maps code points to code
// points, strings or None (delete); characters it has no entry for stay.
func strTranslate(s string, table any) string {
	var b strings.Builder
	for _, r := range s {
		v, found := tableLookup(table, int64(r))
		switch {
		case !found:
			b.WriteRune(r)
		case IsNone(v):
		default:
			switch x := Unbox(v).(type) {
			case string:
				b.WriteString(x)
			case int64:
				if x < 0 || x > 0x10ffff {
					panic(valueError("character mapping must be in range(0x110000)"))
				}
				b.WriteRune(rune(x))
			default:
				panic(newError("TypeError", "character mapping must return integer, None or str"))
			}
		}
	}
	return b.String()
}

func tableLookup(table any, key int64) (v any, found bool) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*PyError); ok && (e.Type == "LookupError" || IsSubclass(e.Type, "LookupError")) {
				v, found = nil, false
				return
			}
			panic(r)
		}
	}()
	return DynGetItem(table, key), true
}

func strFormatMap(format string, mapping any) string {
	// {name} reads the mapping; a missing key raises KeyError.
	var kw []KW
	m := DynDict(mapping)
	for c := m.Iter(); c.Next(); {
		if k, ok := c.Value().(string); ok {
			kw = append(kw, KW{Name: k, Val: m.Get(k)})
		}
	}
	return strFormat(format, nil, kw)
}

// lineBreak finds the first line boundary of splitlines: its position and
// length.
func lineBreak(s string) (int, int) {
	for i, r := range s {
		switch r {
		case '\n', '\v', '\f', 0x1c, 0x1d, 0x1e, 0x85, 0x2028, 0x2029:
			return i, utf8.RuneLen(r)
		case '\r':
			if i+1 < len(s) && s[i+1] == '\n' {
				return i, 2
			}
			return i, 1
		}
	}
	return -1, 0
}

// argNamed reads a str argument, with the message of a built-in method when it
// is something else.
func argNamed(fn, param string, args []any, i int, def string) string {
	if i >= len(args) || IsNone(args[i]) {
		return def
	}
	s, ok := Unbox(args[i]).(string)
	if !ok {
		panic(newError("TypeError", fn+"() argument '"+param+"' must be str, not "+TypeName(Unbox(args[i]))))
	}
	return s
}

// countArgsClinic checks the number of arguments of a method generated with
// Argument Clinic, whose message is "name expected at least 1 argument, got 0".
func countArgsClinic(name string, args []any, min, max int) {
	n := len(args)
	if n >= min && n <= max {
		return
	}
	switch {
	case n < min && min == max:
		panic(newError("TypeError", name+" expected "+itoa(min)+" argument"+plural(int64(min), "", "s")+", got "+itoa(n)))
	case n < min:
		panic(newError("TypeError", name+" expected at least "+itoa(min)+" argument"+plural(int64(min), "", "s")+", got "+itoa(n)))
	}
	panic(newError("TypeError", name+" expected at most "+itoa(max)+" argument"+plural(int64(max), "", "s")+", got "+itoa(n)))
}

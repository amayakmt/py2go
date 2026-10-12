package runtime

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Python strings index by code point and Go strings by byte, so every
// operation that counts or indexes goes through this file. Each is O(n).

func StrLen(s string) int64 { return int64(utf8.RuneCountInString(s)) }

// Chars splits a string into its code points, one string each. It is what
// "for c in s" iterates over.
func Chars(s string) *List[string] {
	out := make([]string, 0, len(s))
	for _, r := range s {
		out = append(out, string(r))
	}
	return &List[string]{items: out}
}

func StrIndex(s string, i int64) string {
	n := StrLen(s)
	if i < 0 {
		i += n
	}
	if i < 0 || i >= n {
		panic(indexError("string index out of range"))
	}
	for _, r := range s {
		if i == 0 {
			return string(r)
		}
		i--
	}
	panic("unreachable")
}

func StrSlice(s string, b Bounds) string {
	runes := []rune(s)
	start, step, count := b.indices(int64(len(runes)))
	out := make([]rune, 0, count)
	for i := int64(0); i < count; i++ {
		out = append(out, runes[start+i*step])
	}
	return string(out)
}

// StrRepeat is str * n. A count below 1 gives the empty string.
func StrRepeat(s string, n int64) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(s, int(n))
}

func StrContains(s, sub string) bool { return strings.Contains(s, sub) }

// isSpace is Python's str.isspace for one character: Go's unicode.IsSpace
// plus the four ASCII separators 0x1c to 0x1f that Python also counts.
func isSpace(r rune) bool { return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f) }

func StrUpper(s string) string { return strings.ToUpper(s) }
func StrLower(s string) string { return strings.ToLower(s) }

func StrStrip(s string) string { return strings.TrimFunc(s, isSpace) }

func StrStripChars(s, chars string) string {
	return strings.TrimFunc(s, func(r rune) bool { return strings.ContainsRune(chars, r) })
}

// StrSplit is split() with no separator: runs of whitespace separate
// items and leading or trailing whitespace yields no empty items.
func StrSplit(s string) *List[string] {
	return NewList(strings.FieldsFunc(s, isSpace)...)
}

// StrSplitSep is split(sep). Unlike strings.Split it rejects an empty
// separator, and "".split(",") is [”] in both.
func StrSplitSep(s, sep string) *List[string] {
	if sep == "" {
		panic(valueError("empty separator"))
	}
	return NewList(strings.Split(s, sep)...)
}

func StrJoin(sep string, parts *List[string]) string { return strings.Join(parts.items, sep) }

func StrReplace(s, old, repl string) string { return strings.ReplaceAll(s, old, repl) }

// StrFind returns the code point index of the first match, or -1.
func StrFind(s, sub string) int64 {
	i := strings.Index(s, sub)
	if i < 0 {
		return -1
	}
	return int64(utf8.RuneCountInString(s[:i]))
}

func StrStartsWith(s, prefix string) bool { return strings.HasPrefix(s, prefix) }
func StrEndsWith(s, suffix string) bool   { return strings.HasSuffix(s, suffix) }
func StrCount(s, sub string) int64        { return int64(strings.Count(s, sub)) }

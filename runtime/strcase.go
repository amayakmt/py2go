package runtime

import (
	"strings"
	"unicode"
)

// Full Unicode case mapping, where one character maps to several: the
// unconditional entries of SpecialCasing.txt that Python applies, which Go's
// one-to-one tables lack.

// FullUpper is s.upper().
func FullUpper(s string) string {
	if isASCII(s) {
		return strings.ToUpper(s)
	}
	var b strings.Builder
	for _, r := range s {
		if m, ok := specialUpper[r]; ok {
			b.WriteString(m)
		} else {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

// FullLower is s.lower(): İ becomes i and a combining dot, and a capital
// sigma at the end of a word becomes the final sigma.
func FullLower(s string) string {
	if isASCII(s) {
		return strings.ToLower(s)
	}
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		switch {
		case specialLower[r] != "":
			b.WriteString(specialLower[r])
		case r == 'Σ' && finalSigma(runes, i):
			b.WriteRune('ς')
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// finalSigma is the Final_Sigma condition: preceded by a cased letter and not
// followed by one (ignoring case-ignorable characters between).
func finalSigma(runes []rune, i int) bool {
	cased := func(r rune) bool { return isCased(r) }
	ignorable := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r) || r == '\'' || r == '.' || r == ':' || r == '·' || r == '’'
	}
	j := i - 1
	for j >= 0 && ignorable(runes[j]) {
		j--
	}
	if j < 0 || !cased(runes[j]) {
		return false
	}
	k := i + 1
	for k < len(runes) && ignorable(runes[k]) {
		k++
	}
	return k >= len(runes) || !cased(runes[k])
}

// FullCasefold is s.casefold().
func FullCasefold(s string) string {
	if isASCII(s) {
		return strings.ToLower(s)
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case specialFold[r] != "":
			b.WriteString(specialFold[r])
		case simpleFold[r] != 0:
			b.WriteRune(simpleFold[r])
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// fullTitleRune is the titlecase mapping of one character.
func fullTitleRune(r rune) string {
	if m, ok := specialTitle[r]; ok {
		return m
	}
	return string(unicode.ToTitle(r))
}

// FullSwapcase is s.swapcase().
func FullSwapcase(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		switch {
		case isUpperP(r):
			if r == 'Σ' && finalSigma(runes, i) {
				b.WriteRune('ς')
			} else if r == 'İ' {
				b.WriteString("i\u0307")
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
		case isLowerP(r):
			if m, ok := specialUpper[r]; ok {
				b.WriteString(m)
			} else {
				b.WriteRune(unicode.ToUpper(r))
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// lowerAt is the lower case of runes[i], where a capital sigma depends on its
// neighbours.
func lowerAt(runes []rune, i int) string {
	r := runes[i]
	switch {
	case r == 'İ':
		return "i\u0307"
	case r == 'Σ' && finalSigma(runes, i):
		return "ς"
	}
	return string(unicode.ToLower(r))
}

// FullTitle is s.title(): the first cased letter of each word in title case,
// the rest in lower case.
func FullTitle(s string) string {
	runes := []rune(s)
	var b strings.Builder
	prevCased := false
	for i, r := range runes {
		cased := isCased(r)
		switch {
		case !cased:
			b.WriteRune(r)
		case prevCased:
			b.WriteString(lowerAt(runes, i))
		default:
			b.WriteString(fullTitleRune(r))
		}
		prevCased = cased
	}
	return b.String()
}

// FullCapitalize is s.capitalize().
func FullCapitalize(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		if i == 0 {
			b.WriteString(fullTitleRune(r))
		} else {
			b.WriteString(lowerAt(runes, i))
		}
	}
	return b.String()
}

// isLowerP and isUpperP are the Lowercase and Uppercase properties, which
// reach beyond the Ll and Lu categories.
func isLowerP(r rune) bool { return unicode.IsLower(r) || unicode.Is(unicode.Other_Lowercase, r) }
func isUpperP(r rune) bool { return unicode.IsUpper(r) || unicode.Is(unicode.Other_Uppercase, r) }
func isCased(r rune) bool  { return isLowerP(r) || isUpperP(r) || unicode.IsTitle(r) }

// StrIsLower is s.islower().
func StrIsLower(s string) bool {
	cased := false
	for _, r := range s {
		if isUpperP(r) || unicode.IsTitle(r) {
			return false
		}
		if isLowerP(r) {
			cased = true
		}
	}
	return cased
}

// StrIsUpper is s.isupper().
func StrIsUpper(s string) bool {
	cased := false
	for _, r := range s {
		if isLowerP(r) || unicode.IsTitle(r) {
			return false
		}
		if isUpperP(r) {
			cased = true
		}
	}
	return cased
}

// StrIsTitle is s.istitle().
func StrIsTitle(s string) bool {
	cased, prev := false, false
	for _, r := range s {
		switch {
		case isUpperP(r) || unicode.IsTitle(r):
			if prev {
				return false
			}
			prev, cased = true, true
		case isLowerP(r):
			if !prev {
				return false
			}
			prev, cased = true, true
		default:
			prev = false
		}
	}
	return cased
}

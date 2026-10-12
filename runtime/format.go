package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// spec is a parsed format specification, the part after ":" in an f-string:
//
//	[[fill]align][sign][0][width][,][.precision][type]
//
// The types are b, c, d, e, E, f, F, g, G, n, o, s, x, X and %.
type spec struct {
	fill     rune
	align    byte // '<', '>', '^', '=' or 0 for the type's default
	sign     byte // '+', '-' or ' '
	zero     bool
	width    int
	comma    bool
	under    bool // '_' as the grouping character
	alt      bool // '#'
	hasPrec  bool
	prec     int
	typ      byte // the type letter, or 0
	explicit bool // the spec named a fill or align itself
}

func parseSpec(s string) (spec, error) {
	sp := spec{fill: ' ', sign: '-'}
	runes := []rune(s)
	i := 0
	isAlign := func(r rune) bool { return r == '<' || r == '>' || r == '^' || r == '=' }
	switch {
	case len(runes) >= 2 && isAlign(runes[1]):
		sp.fill, sp.align, i = runes[0], byte(runes[1]), 2
		sp.explicit = true
	case len(runes) >= 1 && isAlign(runes[0]):
		sp.align, i = byte(runes[0]), 1
		sp.explicit = true
	}
	if i < len(runes) && (runes[i] == '+' || runes[i] == '-' || runes[i] == ' ') {
		sp.sign = byte(runes[i])
		i++
	}
	if i < len(runes) && runes[i] == '#' {
		sp.alt = true
		i++
	}
	if i < len(runes) && runes[i] == '0' {
		sp.zero = true
		i++
	}
	for ; i < len(runes) && runes[i] >= '0' && runes[i] <= '9'; i++ {
		sp.width = sp.width*10 + int(runes[i]-'0')
	}
	if i < len(runes) && runes[i] == ',' {
		sp.comma = true
		i++
	} else if i < len(runes) && runes[i] == '_' {
		sp.under = true
		i++
	}
	if i < len(runes) && runes[i] == '.' {
		i++
		start := i
		for ; i < len(runes) && runes[i] >= '0' && runes[i] <= '9'; i++ {
			sp.prec = sp.prec*10 + int(runes[i]-'0')
		}
		if i == start {
			return sp, fmt.Errorf("format specifier %q is missing precision digits", s)
		}
		sp.hasPrec = true
	}
	if i < len(runes) {
		if i != len(runes)-1 {
			return sp, fmt.Errorf("invalid format specifier %q", s)
		}
		sp.typ = byte(runes[i])
		if runes[i] > 0x7f || !strings.ContainsRune("bcdeEfFgGnosxX%", runes[i]) {
			return sp, fmt.Errorf("Unknown format code '%s' for object", string(runes[i]))
		}
	}
	return sp, nil
}

// ValidateSpec checks a format spec against the kind of value it will
// format ("int", "float" or "str"), the way Python would at run time.
// Codegen calls it so that a bad spec is a compile-time error.
func ValidateSpec(kind, s string) error {
	sp, err := parseSpec(s)
	if err != nil {
		if sp.typ == 0 && strings.HasPrefix(err.Error(), "Unknown format code") {
			code := err.Error()
			code = code[strings.Index(code, "'")+1:]
			code = code[:strings.Index(code, "'")]
			return fmt.Errorf("Unknown format code '%s' for object of type '%s'", code, kindTypeName(kind))
		}
		return err
	}
	unknown := func() error {
		return fmt.Errorf("Unknown format code '%s' for object of type '%s'", string(sp.typ), kindTypeName(kind))
	}
	switch kind {
	case "str":
		switch {
		case sp.typ != 0 && sp.typ != 's':
			return unknown()
		case sp.sign != '-' || sp.zero || sp.comma || sp.under || sp.alt || sp.align == '=':
			return fmt.Errorf("Sign not allowed in string format specifier")
		}
	case "int":
		switch {
		case strings.IndexByte("sn", sp.typ) >= 0 && sp.typ == 's':
			return unknown()
		case sp.hasPrec && !strings.ContainsRune("eEfFgG%", rune(sp.typ)) && sp.typ != 0:
			return fmt.Errorf("Precision not allowed in integer format specifier")
		case sp.hasPrec && sp.typ == 0:
			return fmt.Errorf("Precision not allowed in integer format specifier")
		case sp.typ == 'c' && (sp.sign != '-' || sp.alt):
			return fmt.Errorf("Sign not allowed with integer format specifier 'c'")
		}
	case "float":
		switch {
		case strings.ContainsRune("bcdosxX", rune(sp.typ)) && sp.typ != 0:
			return unknown()
		}
	}
	return nil
}

func kindTypeName(kind string) string { return kind }

// SpecIsFloat reports whether the spec ends in the f type, which makes an
// int format as a float.
func SpecIsFloat(s string) bool {
	return s != "" && strings.ContainsRune("eEfFgG%", rune(s[len(s)-1]))
}

func mustSpec(kind, s string) spec {
	if err := ValidateSpec(kind, s); err != nil {
		panic(valueError(err.Error()))
	}
	sp, _ := parseSpec(s)
	return sp
}

// pad places sign and body in a field of the spec's width.
func (sp spec) pad(defaultAlign byte, sign, body string) string {
	fill, align := sp.fill, sp.align
	if sp.zero && !sp.explicit {
		fill, align = '0', '='
	} else if sp.zero {
		fill = '0'
	}
	if align == 0 {
		align = defaultAlign
	}
	n := sp.width - utf8.RuneCountInString(sign) - utf8.RuneCountInString(body)
	if n <= 0 {
		return sign + body
	}
	padding := strings.Repeat(string(fill), n)
	switch align {
	case '<':
		return sign + body + padding
	case '^':
		left := n / 2
		return strings.Repeat(string(fill), left) + sign + body + strings.Repeat(string(fill), n-left)
	case '=':
		return sign + padding + body
	}
	return padding + sign + body
}

func (sp spec) signFor(negative bool) string {
	switch {
	case negative:
		return "-"
	case sp.sign == '+':
		return "+"
	case sp.sign == ' ':
		return " "
	}
	return ""
}

// group inserts thousands separators into the integer part of a number.
func group(num string) string { return groupBy(num, ',', 3) }

// groupBy inserts sep between every n digits of the integer part.
func groupBy(num string, sep byte, n int) string {
	intPart, rest, _ := strings.Cut(num, ".")
	// An exponent stays out of the grouping.
	if k := strings.IndexAny(intPart, "eE"); k >= 0 {
		return num
	}
	var b strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%n == 0 {
			b.WriteByte(sep)
		}
		b.WriteRune(c)
	}
	if strings.Contains(num, ".") {
		return b.String() + "." + rest
	}
	return b.String()
}

// separate groups the digits as the spec asks. With 0 padding the zeros are
// part of the grouped number, so the field is full width with separators
// among them.
func (sp spec) separate(num string, base int, sign string) string {
	if sp.zero && (sp.comma || sp.under) && !sp.explicit {
		intPart, rest, hasFrac := strings.Cut(num, ".")
		tail := ""
		if hasFrac {
			tail = "." + rest
		}
		for {
			grouped := sp.group(intPart, base)
			if len(sign)+len(grouped)+len(tail) >= sp.width {
				if grouped[0] == ',' || grouped[0] == '_' {
					intPart = "0" + intPart
					continue
				}
				return grouped + tail
			}
			intPart = "0" + intPart
		}
	}
	return sp.group(num, base)
}

func (sp spec) group(num string, base int) string {
	switch {
	case sp.comma:
		return group(num)
	case sp.under && base == 10:
		return groupBy(num, '_', 3)
	case sp.under:
		return groupBy(num, '_', 4)
	}
	return num
}

// FormatInt is format(v, spec) for an int.
func FormatInt(v int64, s string) string {
	sp := mustSpec("int", s)
	if strings.ContainsRune("eEfFgG%", rune(sp.typ)) && sp.typ != 0 {
		return FormatFloat(float64(v), s)
	}
	if sp.typ == 'c' {
		return sp.pad('>', "", string(rune(v)))
	}
	base, prefix := 10, ""
	switch sp.typ {
	case 'b':
		base, prefix = 2, "0b"
	case 'o':
		base, prefix = 8, "0o"
	case 'x', 'X':
		base, prefix = 16, "0x"
	}
	digits := strconv.FormatUint(absUint(v), base)
	if sp.typ == 'X' {
		digits = strings.ToUpper(digits)
		prefix = "0X"
	}
	sign := sp.signFor(v < 0)
	if sp.alt && base != 10 {
		sign += prefix
	}
	digits = sp.separate(digits, base, sign)
	return sp.pad('>', sign, digits)
}

func absUint(v int64) uint64 {
	if v < 0 {
		return -uint64(v)
	}
	return uint64(v)
}

// FormatFloat is format(v, spec) for a float.
func FormatFloat(v float64, s string) string {
	sp := mustSpec("float", s)
	negative := math.Signbit(v) && !math.IsNaN(v)
	var body string
	a := math.Abs(v)
	switch {
	case math.IsNaN(v):
		body = "nan"
		if sp.typ == 'F' || sp.typ == 'E' || sp.typ == 'G' {
			body = "NAN"
		}
	case math.IsInf(v, 0):
		body = "inf"
		if sp.typ == 'F' || sp.typ == 'E' || sp.typ == 'G' {
			body = "INF"
		}
		if sp.typ == '%' {
			body += "%"
		}
	default:
		prec := 6
		if sp.hasPrec {
			prec = sp.prec
		}
		switch sp.typ {
		case 'f', 'F':
			body = strconv.FormatFloat(a, 'f', prec, 64)
		case 'e', 'E':
			body = strconv.FormatFloat(a, 'e', prec, 64)
			if sp.typ == 'E' {
				body = strings.ToUpper(body)
			}
		case 'g', 'G', 'n':
			body = generalFormat(a, prec, sp.alt)
			if sp.typ == 'G' {
				body = strings.ToUpper(body)
			}
		case '%':
			body = strconv.FormatFloat(a*100, 'f', prec, 64) + "%"
		default:
			if sp.hasPrec {
				body = generalFormat(a, prec, sp.alt)
			} else {
				body = FloatRepr(a)
			}
		}
		if sp.alt && !strings.ContainsAny(body, ".eEn") {
			body += "."
		}
	}
	sign := sp.signFor(negative)
	if !math.IsNaN(v) && !math.IsInf(v, 0) {
		body = sp.separate(body, 10, sign)
	}
	return sp.pad('>', sign, body)
}

// generalFormat is the g format: the shortest of the fixed and exponent forms
// for the given number of significant digits.
func generalFormat(a float64, prec int, keepZeros bool) string {
	if prec == 0 {
		prec = 1
	}
	if a == 0 {
		if keepZeros {
			return "0." + strings.Repeat("0", prec-1)
		}
		return "0"
	}
	exp := strconv.FormatFloat(a, 'e', prec-1, 64)
	x, _ := strconv.Atoi(exp[strings.IndexByte(exp, 'e')+1:])
	var out string
	if x >= -4 && x < prec {
		out = strconv.FormatFloat(a, 'f', prec-1-x, 64)
	} else {
		out = exp
	}
	if !keepZeros {
		mant, e, hasExp := strings.Cut(out, "e")
		if strings.Contains(mant, ".") {
			mant = strings.TrimRight(strings.TrimRight(mant, "0"), ".")
		}
		out = mant
		if hasExp {
			out += "e" + e
		}
	}
	return out
}

// FormatStr is format(v, spec) for a str.
func FormatStr(v string, s string) string {
	sp := mustSpec("str", s)
	if sp.hasPrec {
		if r := []rune(v); len(r) > sp.prec {
			v = string(r[:sp.prec])
		}
	}
	return sp.pad('<', "", v)
}

// FormatIntDyn formats an int with a spec known only at run time: a spec that
// ends in f formats it as a float.
func FormatIntDyn(v int64, s string) string {
	if SpecIsFloat(s) {
		return FormatFloat(float64(v), s)
	}
	return FormatInt(v, s)
}

// FormatOther formats a value that takes no spec: only the empty one.
func FormatOther(v any, s string) string {
	if s == "" {
		return Str(v)
	}
	panic(newError("TypeError", "unsupported format string passed to "+TypeName(v)+".__format__"))
}

// FormatDyn formats a value of any type with a spec.
func FormatDyn(v any, s string) string {
	v = Unbox(v)
	switch x := v.(type) {
	case string:
		return FormatStr(x, s)
	case int64:
		return FormatIntDyn(x, s)
	case float64:
		return FormatFloat(x, s)
	case bool:
		if s == "" {
			return Str(x)
		}
		return FormatIntDyn(b2i(x), s)
	case interface{ PyFormat(spec string) string }:
		return x.PyFormat(s)
	}
	return FormatOther(v, s)
}

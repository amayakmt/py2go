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
// Only the subset the transpiler supports is accepted: types d, f and s.
type spec struct {
	fill     rune
	align    byte // '<', '>', '^', '=' or 0 for the type's default
	sign     byte // '+', '-' or ' '
	zero     bool
	width    int
	comma    bool
	hasPrec  bool
	prec     int
	typ      byte // 'd', 'f', 's' or 0
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
		if runes[i] > 0x7f || !strings.ContainsRune("dfs", runes[i]) {
			return sp, fmt.Errorf("format type %q is not supported (use d, f or s)", string(runes[i]))
		}
	}
	if sp.zero && sp.comma {
		return sp, fmt.Errorf("format specifier %q: '0' together with ',' is not supported", s)
	}
	return sp, nil
}

// ValidateSpec checks a format spec against the kind of value it will
// format ("int", "float" or "str"), the way Python would at run time.
// Codegen calls it so that a bad spec is a compile-time error.
func ValidateSpec(kind, s string) error {
	sp, err := parseSpec(s)
	if err != nil {
		return err
	}
	switch kind {
	case "str":
		switch {
		case sp.typ != 0 && sp.typ != 's':
			return fmt.Errorf("unknown format code %q for object of type 'str'", string(sp.typ))
		case sp.sign != '-' || sp.zero || sp.comma || sp.align == '=':
			return fmt.Errorf("sign, '0', ',' and '=' are not allowed in a string format specifier")
		}
	case "int":
		switch {
		case sp.typ == 's':
			return fmt.Errorf("unknown format code 's' for object of type 'int'")
		case sp.hasPrec && sp.typ != 'f':
			return fmt.Errorf("precision not allowed in integer format specifier")
		}
	case "float":
		switch {
		case sp.typ == 'd' || sp.typ == 's':
			return fmt.Errorf("unknown format code %q for object of type 'float'", string(sp.typ))
		case sp.hasPrec && sp.typ != 'f':
			return fmt.Errorf("a precision needs the 'f' type (g-style formatting is not supported)")
		}
	}
	return nil
}

// SpecIsFloat reports whether the spec ends in the f type, which makes an
// int format as a float.
func SpecIsFloat(s string) bool { return strings.HasSuffix(s, "f") }

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
func group(num string) string {
	intPart, rest, _ := strings.Cut(num, ".")
	var b strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	if strings.Contains(num, ".") {
		return b.String() + "." + rest
	}
	return b.String()
}

// FormatInt is format(v, spec) for an int.
func FormatInt(v int64, s string) string {
	sp := mustSpec("int", s)
	digits := strconv.FormatUint(absUint(v), 10)
	if sp.comma {
		digits = group(digits)
	}
	return sp.pad('>', sp.signFor(v < 0), digits)
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
	switch {
	case math.IsNaN(v):
		body = "nan"
	case math.IsInf(v, 0):
		body = "inf"
	case sp.typ == 'f':
		prec := 6
		if sp.hasPrec {
			prec = sp.prec
		}
		body = strconv.FormatFloat(math.Abs(v), 'f', prec, 64)
	default:
		body = FloatRepr(math.Abs(v))
	}
	if sp.comma && !math.IsNaN(v) && !math.IsInf(v, 0) {
		body = group(body)
	}
	return sp.pad('>', sp.signFor(negative), body)
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

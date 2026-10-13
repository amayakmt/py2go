package runtime

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

//go:embed unicodedata.json.gz
var unicodedataGz []byte

type udRun struct {
	lo, hi rune
	val    any
}

type udData struct {
	version    string
	cat        []udRun
	bidi       []udRun
	east       []udRun
	comb       []udRun
	mirrored   []udRun
	decomp     map[rune]string
	decimal    map[rune]int64
	digit      map[rune]int64
	numeric    map[rune]string
	names      map[rune]string
	nameRanges []udRun
	byName     map[string]rune
	pairs      map[[2]rune]rune
	decompMap  map[rune][]rune // canonical full decompositions, computed on demand
}

var (
	udOnce sync.Once
	ud     *udData
)

func loadUD() *udData {
	udOnce.Do(func() {
		zr, err := gzip.NewReader(bytes.NewReader(unicodedataGz))
		if err != nil {
			panic(newError("SystemError", "unicodedata: "+err.Error()))
		}
		var raw struct {
			Version    string            `json:"version"`
			Cat        [][3]any          `json:"cat"`
			Bidi       [][3]any          `json:"bidi"`
			East       [][3]any          `json:"east"`
			Comb       [][3]any          `json:"comb"`
			Mirrored   [][3]any          `json:"mirrored"`
			Decomp     map[string]string `json:"decomp"`
			Decimal    map[string]int64  `json:"decimal"`
			Digit      map[string]int64  `json:"digit"`
			Numeric    map[string]string `json:"numeric"`
			Name       map[string]string `json:"name"`
			NameRanges [][3]any          `json:"nameranges"`
			Pairs      [][3]int64        `json:"pairs"`
		}
		if err := json.NewDecoder(zr).Decode(&raw); err != nil {
			panic(newError("SystemError", "unicodedata: "+err.Error()))
		}
		d := &udData{version: raw.Version}
		conv := func(in [][3]any) []udRun {
			out := make([]udRun, len(in))
			for i, r := range in {
				out[i] = udRun{rune(r[0].(float64)), rune(r[1].(float64)), r[2]}
			}
			return out
		}
		d.cat, d.bidi, d.east, d.comb, d.mirrored = conv(raw.Cat), conv(raw.Bidi), conv(raw.East), conv(raw.Comb), conv(raw.Mirrored)
		d.nameRanges = conv(raw.NameRanges)
		key := func(s string) rune { n, _ := strconv.Atoi(s); return rune(n) }
		d.decomp = map[rune]string{}
		for k, v := range raw.Decomp {
			d.decomp[key(k)] = v
		}
		d.decimal, d.digit = map[rune]int64{}, map[rune]int64{}
		for k, v := range raw.Decimal {
			d.decimal[key(k)] = v
		}
		for k, v := range raw.Digit {
			d.digit[key(k)] = v
		}
		d.numeric = map[rune]string{}
		for k, v := range raw.Numeric {
			d.numeric[key(k)] = v
		}
		d.names = map[rune]string{}
		d.byName = map[string]rune{}
		for k, v := range raw.Name {
			d.names[key(k)] = v
			d.byName[v] = key(k)
		}
		d.pairs = map[[2]rune]rune{}
		for _, p := range raw.Pairs {
			d.pairs[[2]rune{rune(p[0]), rune(p[1])}] = rune(p[2])
		}
		d.decompMap = map[rune][]rune{}
		ud = d
	})
	return ud
}

func (d *udData) find(runs []udRun, r rune) (any, bool) {
	i := sort.Search(len(runs), func(i int) bool { return runs[i].hi >= r })
	if i < len(runs) && runs[i].lo <= r {
		return runs[i].val, true
	}
	return nil, false
}

func udChar(fn string, v any) rune {
	s, ok := Unbox(v).(string)
	if !ok {
		panic(newError("TypeError", fn+"() argument must be a unicode character, not "+TypeName(Unbox(v))))
	}
	if utf8.RuneCountInString(s) != 1 {
		panic(newError("TypeError", fn+"() argument must be a unicode character, not str"))
	}
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

const (
	hangulSBase  = 0xAC00
	hangulLBase  = 0x1100
	hangulVBase  = 0x1161
	hangulTBase  = 0x11A7
	hangulLCount = 19
	hangulVCount = 21
	hangulTCount = 28
	hangulNCount = hangulVCount * hangulTCount
	hangulSCount = hangulLCount * hangulNCount
)

var jamoL = []string{"G", "GG", "N", "D", "DD", "R", "M", "B", "BB", "S", "SS", "", "J", "JJ", "C", "K", "T", "P", "H"}
var jamoV = []string{"A", "AE", "YA", "YAE", "EO", "E", "YEO", "YE", "O", "WA", "WAE", "OE", "YO", "U", "WEO", "WE", "WI", "YU", "EU", "YI", "I"}
var jamoT = []string{"", "G", "GG", "GS", "N", "NJ", "NH", "D", "L", "LG", "LM", "LB", "LS", "LT", "LP", "LH", "M", "B", "BS", "S", "SS", "NG", "J", "C", "K", "T", "P", "H"}

func (d *udData) name(r rune) (string, bool) {
	if r >= hangulSBase && r < hangulSBase+hangulSCount {
		s := r - hangulSBase
		return "HANGUL SYLLABLE " + jamoL[s/hangulNCount] + jamoV[(s%hangulNCount)/hangulTCount] + jamoT[s%hangulTCount], true
	}
	if n, ok := d.names[r]; ok {
		return n, true
	}
	if v, ok := d.find(d.nameRanges, r); ok {
		return v.(string) + strings.ToUpper(strconv.FormatInt(int64(r), 16)), true
	}
	return "", false
}

func (d *udData) lookup(name string) (rune, bool) {
	name = strings.ToUpper(name)
	if r, ok := d.byName[name]; ok {
		return r, true
	}
	if strings.HasPrefix(name, "HANGUL SYLLABLE ") {
		rest := name[len("HANGUL SYLLABLE "):]
		for l, ls := range jamoL {
			if !strings.HasPrefix(rest, ls) {
				continue
			}
			for v, vs := range jamoV {
				if !strings.HasPrefix(rest[len(ls):], vs) {
					continue
				}
				for t, ts := range jamoT {
					if rest[len(ls)+len(vs):] == ts {
						return rune(hangulSBase + (l*hangulVCount+v)*hangulTCount + t), true
					}
				}
			}
		}
	}
	for _, rg := range d.nameRanges {
		prefix := rg.val.(string)
		if strings.HasPrefix(name, prefix) {
			n, err := strconv.ParseInt(name[len(prefix):], 16, 32)
			if err == nil && rune(n) >= rg.lo && rune(n) <= rg.hi {
				return rune(n), true
			}
		}
	}
	return 0, false
}

// decompOf is the decomposition mapping of r, with its tag.
func (d *udData) decompOf(r rune) (tag string, runes []rune) {
	s, ok := d.decomp[r]
	if !ok {
		return "", nil
	}
	fields := strings.Fields(s)
	if strings.HasPrefix(fields[0], "<") {
		tag = fields[0]
		fields = fields[1:]
	}
	for _, f := range fields {
		n, _ := strconv.ParseInt(f, 16, 32)
		runes = append(runes, rune(n))
	}
	return tag, runes
}

func (d *udData) combining(r rune) int {
	if v, ok := d.find(d.comb, r); ok {
		return int(v.(float64))
	}
	return 0
}

func (d *udData) decompose(out []rune, r rune, compat bool) []rune {
	if r >= hangulSBase && r < hangulSBase+hangulSCount {
		s := r - hangulSBase
		out = append(out, hangulLBase+s/hangulNCount, hangulVBase+(s%hangulNCount)/hangulTCount)
		if t := s % hangulTCount; t != 0 {
			out = append(out, hangulTBase+t)
		}
		return out
	}
	tag, parts := d.decompOf(r)
	if parts == nil || (tag != "" && !compat) {
		return append(out, r)
	}
	for _, p := range parts {
		out = d.decompose(out, p, compat)
	}
	return out
}

func (d *udData) compose(rs []rune) []rune {
	if len(rs) == 0 {
		return rs
	}
	out := []rune{rs[0]}
	starter := -1
	if d.combining(rs[0]) == 0 {
		starter = 0
	}
	lastCC := 0
	if starter < 0 {
		lastCC = d.combining(rs[0])
	}
	for _, c := range rs[1:] {
		cc := d.combining(c)
		if starter >= 0 {
			sc := out[starter]
			// blocked: a character between with cc >= this one's (or a starter)
			blocked := len(out)-1 > starter && (lastCC >= cc || lastCC == 0)
			if !blocked || len(out)-1 == starter {
				var comp rune
				found := false
				switch {
				case sc >= hangulLBase && sc < hangulLBase+hangulLCount && c >= hangulVBase && c < hangulVBase+hangulVCount && len(out)-1 == starter:
					comp, found = hangulSBase+((sc-hangulLBase)*hangulVCount+(c-hangulVBase))*hangulTCount, true
				case sc >= hangulSBase && sc < hangulSBase+hangulSCount && (sc-hangulSBase)%hangulTCount == 0 && c > hangulTBase && c < hangulTBase+hangulTCount && len(out)-1 == starter:
					comp, found = sc+(c-hangulTBase), true
				default:
					comp, found = d.pairs[[2]rune{sc, c}]
				}
				if found {
					out[starter] = comp
					continue
				}
			}
		}
		if cc == 0 {
			starter = len(out)
		}
		out = append(out, c)
		lastCC = cc
	}
	return out
}

func (d *udData) normalize(form, s string) string {
	compat := form == "NFKC" || form == "NFKD"
	var rs []rune
	for _, r := range s {
		rs = d.decompose(rs, r, compat)
	}
	// canonical ordering
	for i := 1; i < len(rs); i++ {
		cc := d.combining(rs[i])
		if cc == 0 {
			continue
		}
		j := i
		for j > 0 {
			pc := d.combining(rs[j-1])
			if pc <= cc {
				break
			}
			rs[j], rs[j-1] = rs[j-1], rs[j]
			j--
		}
	}
	if form == "NFC" || form == "NFKC" {
		rs = d.compose(rs)
	}
	return string(rs)
}

func init() {
	one := func(name string, f func(d *udData, r rune) any) {
		RegisterNative("unicodedata."+name, func(args []any, kw []KW) any {
			nativeArgs(name, args, 1, 1)
			return f(loadUD(), udChar(name, args[0]))
		})
	}
	one("category", func(d *udData, r rune) any {
		if v, ok := d.find(d.cat, r); ok {
			return v.(string)
		}
		return "Cn"
	})
	one("bidirectional", func(d *udData, r rune) any {
		if v, ok := d.find(d.bidi, r); ok {
			return v.(string)
		}
		return ""
	})
	one("east_asian_width", func(d *udData, r rune) any {
		if v, ok := d.find(d.east, r); ok {
			return v.(string)
		}
		return "N"
	})
	one("combining", func(d *udData, r rune) any { return int64(d.combining(r)) })
	one("mirrored", func(d *udData, r rune) any {
		_, ok := d.find(d.mirrored, r)
		return int64(b2i(ok))
	})
	one("decomposition", func(d *udData, r rune) any { return d.decomp[r] })
	withDefault := func(name string, f func(d *udData, r rune) (any, bool), msg string) {
		RegisterNative("unicodedata."+name, func(args []any, kw []KW) any {
			nativeArgs(name, args, 1, 2)
			r := udChar(name, args[0])
			if v, ok := f(loadUD(), r); ok {
				return v
			}
			if len(args) == 2 {
				return args[1]
			}
			panic(valueError(msg))
		})
	}
	withDefault("decimal", func(d *udData, r rune) (any, bool) { v, ok := d.decimal[r]; return v, ok }, "not a decimal")
	withDefault("digit", func(d *udData, r rune) (any, bool) { v, ok := d.digit[r]; return v, ok }, "not a digit")
	withDefault("numeric", func(d *udData, r rune) (any, bool) {
		v, ok := d.numeric[r]
		if !ok {
			return nil, false
		}
		f, _ := strconv.ParseFloat(v, 64)
		return f, true
	}, "not a numeric character")
	withDefault("name", func(d *udData, r rune) (any, bool) { return d.name(r) }, "no such name")
	RegisterNative("unicodedata.lookup", func(args []any, kw []KW) any {
		nativeArgs("lookup", args, 1, 1)
		name := Str(args[0])
		if r, ok := loadUD().lookup(name); ok {
			return string(r)
		}
		panic(newError("KeyError", "undefined character name '"+name+"'"))
	})
	RegisterNative("unicodedata.normalize", func(args []any, kw []KW) any {
		nativeArgs("normalize", args, 2, 2)
		form, ok := Unbox(args[0]).(string)
		if !ok {
			panic(newError("TypeError", "normalize() argument 1 must be str, not "+TypeName(Unbox(args[0]))))
		}
		s, ok := Unbox(args[1]).(string)
		if !ok {
			panic(newError("TypeError", "normalize() argument 2 must be str, not "+TypeName(Unbox(args[1]))))
		}
		switch form {
		case "NFC", "NFD", "NFKC", "NFKD":
		default:
			panic(valueError("invalid normalization form"))
		}
		return loadUD().normalize(form, s)
	})
	RegisterNative("unicodedata.is_normalized", func(args []any, kw []KW) any {
		nativeArgs("is_normalized", args, 2, 2)
		form, s := Str(args[0]), Str(args[1])
		switch form {
		case "NFC", "NFD", "NFKC", "NFKD":
		default:
			panic(valueError("invalid normalization form"))
		}
		return loadUD().normalize(form, s) == s
	})
	RegisterNative("unicodedata.unidata_version", func(args []any, kw []KW) any { return loadUD().version })
	_ = math.Inf
}

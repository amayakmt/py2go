package runtime

import (
	"fmt"
	"math"
	"os"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata" // the zone database, for machines that have none
)

// tm is a broken-down time: the nine fields of a struct_time and its zone.
type tm struct {
	year, mon, mday, hour, min, sec, wday, yday, isdst int
	zone                                               string
	gmtoff                                             int
}

func tmFromTime(t time.Time) tm {
	name, off := t.Zone()
	dst := 0
	if t.IsDST() {
		dst = 1
	}
	return tm{t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second(), (int(t.Weekday()) + 6) % 7, t.YearDay(), dst, name, off}
}

func (t tm) tuple() any {
	return NewVarTuple[any](int64(t.year), int64(t.mon), int64(t.mday), int64(t.hour), int64(t.min), int64(t.sec),
		int64(t.wday), int64(t.yday), int64(t.isdst), t.zone, int64(t.gmtoff))
}

// secondsArg reads the seconds of gmtime and localtime: the current time for
// None, otherwise a number cut down to whole seconds.
func secondsArg(args []any) int64 {
	if len(args) == 0 || IsNone(args[0]) {
		return time.Now().Unix()
	}
	switch v := Unbox(args[0]).(type) {
	case float64:
		if math.IsNaN(v) {
			panic(valueError("Invalid value NaN (not a number)"))
		}
		if math.IsInf(v, 0) || math.Abs(v) > 9.2e18 {
			panic(overflowError("timestamp out of range for platform time_t"))
		}
		return int64(math.Floor(v))
	}
	n, ok := intOf(Unbox(args[0]))
	if !ok {
		panic(newError("TypeError", "'"+TypeName(Unbox(args[0]))+"' object cannot be interpreted as an integer"))
	}
	return n
}

func timeFromSecs(sec int64, loc *time.Location) time.Time {
	t := time.Unix(sec, 0).In(loc)
	if y := t.Year(); y > 9999 || y < -9999 {
		panic(newError("OSError", "[Errno 75] Value too large for defined data type"))
	}
	return t
}

// tmArg reads the tuple strftime and mktime take.
func tmArg(v any, fn string) tm {
	var f []int64
	var extra []any
	i := 0
	for c := DynIter(v).Iter(); c.Next(); i++ {
		x := c.Value()
		if i < 9 {
			n, ok := intOf(Unbox(x))
			if !ok {
				panic(newError("TypeError", "'"+TypeName(Unbox(x))+"' object cannot be interpreted as an integer"))
			}
			f = append(f, n)
		} else {
			extra = append(extra, x)
		}
	}
	if i < 9 || i > 11 {
		panic(newError("TypeError", fn+"(): illegal time tuple argument"))
	}
	t := tm{year: int(f[0]), mon: int(f[1]), mday: int(f[2]), hour: int(f[3]), min: int(f[4]), sec: int(f[5]), wday: int(f[6]), yday: int(f[7]), isdst: int(f[8])}
	if len(extra) > 0 && !IsNone(extra[0]) {
		t.zone = Str(extra[0])
	}
	if len(extra) > 1 && !IsNone(extra[1]) {
		t.gmtoff = int(DynInt(extra[1]))
	}
	return t
}

func checkTm(t *tm) {
	if t.mon == 0 {
		t.mon = 1
	}
	if t.mday == 0 {
		t.mday = 1
	}
	switch {
	case t.mon < 1 || t.mon > 12:
		panic(valueError("month out of range"))
	case t.mday < 1 || t.mday > 31:
		panic(valueError("day of month out of range"))
	case t.hour < 0 || t.hour > 23:
		panic(valueError("hour out of range"))
	case t.min < 0 || t.min > 59:
		panic(valueError("minute out of range"))
	case t.sec < 0 || t.sec > 61:
		panic(valueError("seconds out of range"))
	case t.wday < -1 || t.wday > 6 && false:
		panic(valueError("day of week out of range"))
	case t.yday < 0 || t.yday > 366:
		panic(valueError("day of year out of range"))
	}
	if t.wday < 0 {
		t.wday = 6
	} else {
		t.wday = t.wday % 7
	}
	if t.yday == 0 {
		t.yday = 1
	}
}

func timeOf(t tm, loc *time.Location) time.Time {
	return time.Date(t.year, time.Month(t.mon), t.mday, t.hour, t.min, t.sec, 0, loc)
}

// tzOffsets is time.timezone, time.altzone, time.daylight and time.tzname.
func tzOffsets() (timezone, altzone int, daylight bool, names [2]string) {
	year := time.Now().Year()
	jan := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	jul := time.Date(year, 7, 1, 0, 0, 0, 0, time.Local)
	jn, jo := jan.Zone()
	ln, lo := jul.Zone()
	if jan.IsDST() == jul.IsDST() {
		// no daylight saving time in effect in this zone this year
		return -jo, -jo, false, [2]string{jn, jn}
	}
	if jan.IsDST() {
		return -lo, -jo, true, [2]string{ln, jn}
	}
	return -jo, -lo, true, [2]string{jn, ln}
}

func init() {
	RegisterNative("time.gmtime", func(args []any, kw []KW) any {
		t := tmFromTime(timeFromSecs(secondsArg(args), time.UTC))
		t.zone = "GMT"
		return t.tuple()
	})
	RegisterNative("time.localtime", func(args []any, kw []KW) any {
		return tmFromTime(timeFromSecs(secondsArg(args), time.Local)).tuple()
	})
	RegisterNative("time.mktime", func(args []any, kw []KW) any {
		t := tmArg(args[0], "mktime")
		if t.year < 1 || t.year > 9999 {
			panic(overflowError("mktime argument out of range"))
		}
		tt := timeOf(t, time.Local)
		// tm_isdst forces the other offset when the clock time is ambiguous
		if t.isdst == 1 && !tt.IsDST() {
			tt = tt.Add(-time.Hour)
		} else if t.isdst == 0 && tt.IsDST() {
			tt = tt.Add(time.Hour)
		}
		return float64(tt.Unix())
	})
	RegisterNative("time.timegm", func(args []any, kw []KW) any {
		t := tmArg(args[0], "timegm")
		return timeOf(t, time.UTC).Unix()
	})
	RegisterNative("time.strftime", func(args []any, kw []KW) any {
		f := Str(args[0])
		t := tmArg(args[1], "strftime")
		checkTm(&t)
		if strings.IndexByte(f, 0) >= 0 {
			panic(valueError("embedded null character"))
		}
		return strftime(f, t)
	})
	RegisterNative("time.tzinfo", func(args []any, kw []KW) any {
		tz, alt, dl, names := tzOffsets()
		return NewVarTuple[any](int64(tz), int64(alt), dl, names[0], names[1])
	})
	RegisterNative("time.tzset", func(args []any, kw []KW) any {
		tz, set := os.LookupEnv("TZ")
		switch {
		case !set:
			if l, err := time.LoadLocation("Local"); err == nil {
				time.Local = l
			}
		case tz == "":
			time.Local = time.UTC
		default:
			name := strings.TrimPrefix(tz, ":")
			if l, err := time.LoadLocation(name); err == nil {
				time.Local = l
			} else if l := posixTZ(name); l != nil {
				time.Local = l
			} else {
				time.Local = time.FixedZone(name, 0)
			}
		}
		return nil
	})
	RegisterNative("time.process_time", func(args []any, kw []KW) any {
		var ru syscall.Rusage
		syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
		return float64(ru.Utime.Sec+ru.Stime.Sec) + float64(ru.Utime.Usec+ru.Stime.Usec)/1e6
	})
	RegisterNative("time.process_time_ns", func(args []any, kw []KW) any {
		var ru syscall.Rusage
		syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
		return (int64(ru.Utime.Sec)+int64(ru.Stime.Sec))*1_000_000_000 + (int64(ru.Utime.Usec)+int64(ru.Stime.Usec))*1000
	})
}

// posixTZ reads TZ strings such as "EST5EDT" or "UTC0": a name and the hours
// west of UTC. A daylight name is accepted and ignored.
func posixTZ(s string) *time.Location {
	i := 0
	for i < len(s) && (s[i] >= 'A' && s[i] <= 'Z' || s[i] >= 'a' && s[i] <= 'z') {
		i++
	}
	if i < 3 || i >= len(s) {
		return nil
	}
	name := s[:i]
	j := i
	if j < len(s) && (s[j] == '+' || s[j] == '-') {
		j++
	}
	k := j
	for k < len(s) && s[k] >= '0' && s[k] <= '9' {
		k++
	}
	if k == j {
		return nil
	}
	var hours, mins int
	fmt.Sscanf(s[j:k], "%d", &hours)
	if k < len(s) && s[k] == ':' {
		fmt.Sscanf(s[k+1:], "%d", &mins)
	}
	off := hours*3600 + mins*60
	if s[i] == '-' {
		off = -off
	}
	return time.FixedZone(name, -off)
}

// strftime is the C library's strftime for the conversions Python allows.
func strftime(format string, t tm) string {
	var b strings.Builder
	date := timeOf(tm{year: t.year, mon: t.mon, mday: t.mday}, time.UTC)
	weekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	hour12 := func() int {
		h := t.hour % 12
		if h == 0 {
			h = 12
		}
		return h
	}
	ampm := func() string {
		if t.hour < 12 {
			return "AM"
		}
		return "PM"
	}
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' || i+1 >= len(format) {
			b.WriteByte(c)
			continue
		}
		i++
		flag := byte(0)
		if (format[i] == '-' || format[i] == '_' || format[i] == '^' || format[i] == '0' || format[i] == '#') && i+1 < len(format) {
			flag = format[i]
			i++
		}
		width := 0
		for i < len(format) && format[i] >= '0' && format[i] <= '9' {
			width = width*10 + int(format[i]-'0')
			i++
		}
		if i >= len(format) {
			b.WriteString("%")
			break
		}
		var piece string
		switch format[i] {
		case 'Y':
			piece = fmt.Sprintf("%d", t.year)
		case 'y':
			piece = fmt.Sprintf("%02d", ((t.year%100)+100)%100)
		case 'C':
			piece = fmt.Sprintf("%02d", t.year/100)
		case 'm':
			piece = fmt.Sprintf("%02d", t.mon)
		case 'd':
			piece = fmt.Sprintf("%02d", t.mday)
		case 'e':
			piece = fmt.Sprintf("%2d", t.mday)
		case 'H':
			piece = fmt.Sprintf("%02d", t.hour)
		case 'k':
			piece = fmt.Sprintf("%2d", t.hour)
		case 'I':
			piece = fmt.Sprintf("%02d", hour12())
		case 'l':
			piece = fmt.Sprintf("%2d", hour12())
		case 'M':
			piece = fmt.Sprintf("%02d", t.min)
		case 'S':
			piece = fmt.Sprintf("%02d", t.sec)
		case 'p':
			piece = ampm()
		case 'P':
			piece = strings.ToLower(ampm())
		case 'a':
			piece = weekdays[t.wday][:3]
		case 'A':
			piece = weekdays[t.wday]
		case 'b', 'h':
			piece = months[t.mon-1][:3]
		case 'B':
			piece = months[t.mon-1]
		case 'j':
			piece = fmt.Sprintf("%03d", t.yday)
		case 'w':
			piece = fmt.Sprintf("%d", (t.wday+1)%7)
		case 'u':
			piece = fmt.Sprintf("%d", t.wday+1)
		case 'U':
			piece = fmt.Sprintf("%02d", (t.yday-1+7-(t.wday+1)%7)/7)
		case 'W':
			piece = fmt.Sprintf("%02d", (t.yday-1+7-t.wday)/7)
		case 'V':
			_, w := date.ISOWeek()
			piece = fmt.Sprintf("%02d", w)
		case 'G':
			y, _ := date.ISOWeek()
			piece = fmt.Sprintf("%d", y)
		case 'g':
			y, _ := date.ISOWeek()
			piece = fmt.Sprintf("%02d", y%100)
		case 'Z':
			piece = t.zone
		case 'z':
			off := t.gmtoff
			sign := '+'
			if off < 0 {
				sign, off = '-', -off
			}
			piece = fmt.Sprintf("%c%02d%02d", sign, off/3600, off%3600/60)
		case 'c':
			piece = strftime("%a %b %e %H:%M:%S %Y", t)
		case 'x', 'D':
			piece = strftime("%m/%d/%y", t)
		case 'X', 'T':
			piece = strftime("%H:%M:%S", t)
		case 'F':
			piece = strftime("%Y-%m-%d", t)
		case 'R':
			piece = strftime("%H:%M", t)
		case 'r':
			piece = strftime("%I:%M:%S %p", t)
		case 's':
			piece = fmt.Sprintf("%d", timeOf(t, time.Local).Unix())
		case 'n':
			piece = "\n"
		case 't':
			piece = "\t"
		case '%':
			piece = "%"
		default:
			piece = "%" + string(format[i])
			if flag != 0 {
				piece = "%" + string(flag) + string(format[i])
			}
			b.WriteString(piece)
			continue
		}
		switch flag {
		case '-':
			trimmed := strings.TrimLeft(piece, "0 ")
			if trimmed == "" {
				trimmed = "0"
			}
			piece = trimmed
		case '_':
			trimmed := strings.TrimLeft(piece, "0")
			if trimmed == "" {
				trimmed = "0"
			}
			piece = strings.Repeat(" ", len(piece)-len(trimmed)) + trimmed
		case '^':
			piece = strings.ToUpper(piece)
		case '0':
			piece = strings.ReplaceAll(piece, " ", "0")
		case '#':
			switch format[i] {
			case 'p', 'P':
				piece = strings.ToLower(piece)
			case 'a', 'A', 'b', 'B', 'h':
				piece = strings.ToUpper(piece)
			}
		}
		if width > len(piece) {
			numeric := strings.ContainsRune("YyCmdeHkIlMSjwuUWVGgs", rune(format[i]))
			pad := " "
			if numeric && flag != '_' {
				pad = "0"
			}
			if numeric && flag == '-' {
				pad = ""
			}
			if piece != "" && (piece[0] == '-' || piece[0] == '+') && pad == "0" {
				piece = piece[:1] + strings.Repeat(pad, width-len(piece)) + piece[1:]
			} else {
				piece = strings.Repeat(pad, width-len(piece)) + piece
			}
		}
		b.WriteString(piece)
	}
	return b.String()
}

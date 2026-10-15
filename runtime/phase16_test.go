package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOSErrorAttributes(t *testing.T) {
	e := NewError("OSError", 2, "gone", "fn")
	if e.Type != "FileNotFoundError" || e.Attrs["errno"] != int64(2) || e.Attrs["filename"] != "fn" || len(e.Args) != 2 {
		t.Fatalf("OSError(2, 'gone', 'fn') = %+v", e)
	}
	if got := e.str(); got != "[Errno 2] gone: 'fn'" {
		t.Errorf("str = %q", got)
	}
	if plain := NewError("OSError", "plain"); plain.Attrs != nil || plain.str() != "plain" {
		t.Errorf("OSError('plain') = %+v", plain)
	}
	two := sysError(os.ErrNotExist, "")
	_ = two
}

func TestSysErrorClasses(t *testing.T) {
	_, err := os.Stat("/nonexistent/x")
	e := sysError(err, "/nonexistent/x")
	if e.Type != "FileNotFoundError" || e.str() != "[Errno 2] No such file or directory: '/nonexistent/x'" {
		t.Errorf("%s: %s", e.Type, e.str())
	}
	e = sysError(err, "a", "b")
	if e.str() != "[Errno 2] No such file or directory: 'a' -> 'b'" {
		t.Errorf("two names: %s", e.str())
	}
}

func TestStrftime(t *testing.T) {
	tm := tm{year: 2024, mon: 3, mday: 5, hour: 13, min: 4, sec: 7, wday: 1, yday: 65, zone: "UTC"}
	tests := map[string]string{
		"%Y-%m-%d %H:%M:%S": "2024-03-05 13:04:07", "%a %A %b %B": "Tue Tuesday Mar March", "%j %U %W": "065 09 10",
		"%-d/%-m %_d %^a": "5/3  5 TUE", "%5Y|%I %p %Z": "02024|01 PM UTC", "%c": "Tue Mar  5 13:04:07 2024", "%%": "%",
	}
	for f, want := range tests {
		if got := strftime(f, tm); got != want {
			t.Errorf("strftime(%q) = %q, want %q", f, got, want)
		}
	}
}

func TestSurrogateEscape(t *testing.T) {
	s := DecodeBytes("caf\xe9", "utf-8", "surrogateescape")
	if string(EncodeStr(s, "utf-8", "surrogateescape")) != "caf\xe9" {
		t.Errorf("round trip = %q", s)
	}
	if got := string(EncodeStr("caf\u00e9", "cp437", "strict")); got != "caf\x82" {
		t.Errorf("cp437 = %q", got)
	}
	if got := DecodeBytes("\x80", "cp1252", "strict"); got != "\u20ac" {
		t.Errorf("cp1252 = %q", got)
	}
}

func TestFileTextAndBinary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	f := OpenFile([]any{path, "w", int64(-1), "utf-8", nil, "\r\n"}, nil)
	f.DynMethod("write", []any{"a\nb\n"}, nil)
	f.closeFile()
	data, _ := os.ReadFile(path)
	if string(data) != "a\r\nb\r\n" {
		t.Fatalf("newline translation wrote %q", data)
	}
	r := OpenFile([]any{path}, nil)
	line, _ := r.DynMethod("readline", nil, nil)
	if line != "a\n" {
		t.Errorf("readline = %q", line)
	}
	pos, _ := r.DynMethod("tell", nil, nil)
	if pos != int64(3) {
		t.Errorf("tell = %v", pos)
	}
	r.closeFile()
	if msg := catchMsg(func() { r.DynMethod("read", nil, nil) }); msg != "ValueError: I/O operation on closed file." {
		t.Errorf("closed read = %q", msg)
	}
	b := OpenFile([]any{path, "rb", int64(0)}, nil)
	if b.PyClassName() != "FileIO" {
		t.Errorf("unbuffered binary file is %s", b.PyClassName())
	}
	b.closeFile()
}

func catchMsg(f func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*PyError); ok {
				msg = e.Type + ": " + e.baseStr()
			}
		}
	}()
	f()
	return ""
}

func TestPosixNatives(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x")
	fd := callNative("posix.open", p, int64(os.O_CREATE|os.O_WRONLY), int64(0o644)).(int64)
	if n := callNative("posix.write", fd, Bytes("hello")); n != int64(5) {
		t.Errorf("write = %v", n)
	}
	callNative("posix.close", fd)
	st := callNative("posix.stat", p, true).(Seq)
	if st.At(6) != int64(5) {
		t.Errorf("size = %v", st.At(6))
	}
	if msg := catchMsg(func() { callNative("posix.stat", p+"2", true) }); !strings.HasPrefix(msg, "FileNotFoundError: [Errno 2]") {
		t.Errorf("stat error = %q", msg)
	}
}

func TestTimeNatives(t *testing.T) {
	tup := callNative("time.gmtime", int64(0)).(Seq)
	if tup.At(0) != int64(1970) || tup.At(9) != "GMT" {
		t.Errorf("gmtime(0) = %v", tup)
	}
	if got := callNative("time.mktime", NewVarTuple[any](int64(2024), int64(1), int64(1), int64(0), int64(0), int64(0), int64(0), int64(1), int64(-1))); got == nil {
		t.Error("mktime")
	}
	loc := loadZone("America/New_York")
	if loc == nil {
		t.Fatal("zone database missing")
	}
	off, dst, abbr := zoneAt(loc, time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC))
	if off != -4*3600 || dst != 3600 || abbr != "EDT" {
		t.Errorf("New_York in July = %d %d %s", off, dst, abbr)
	}
}

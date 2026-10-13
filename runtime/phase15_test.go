package runtime

import (
	"fmt"
	"testing"
)

func callNative(name string, args ...any) any {
	return NativeCall(append([]any{name}, args...), nil)
}

func raises(f func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*PyError); ok {
				msg = e.Type + ": " + e.baseStr()
			} else {
				msg = fmt.Sprint(r)
			}
		}
	}()
	f()
	return ""
}

func TestBinasciiNatives(t *testing.T) {
	tests := []struct {
		name string
		args []any
		want any
	}{
		{"binascii.crc32", []any{Bytes("hello")}, int64(907060870)},
		{"binascii.crc_hqx", []any{Bytes("abc"), int64(0)}, int64(0x9dd6)},
		{"binascii.hexlify", []any{Bytes("\x01\xab")}, Bytes("01ab")},
		{"binascii.unhexlify", []any{Bytes("01ab")}, Bytes("\x01\xab")},
		{"binascii.b2a_base64", []any{Bytes("hi")}, Bytes("aGk=\n")},
		{"binascii.a2b_base64", []any{Bytes("aGk=")}, Bytes("hi")},
	}
	for _, tc := range tests {
		if got := callNative(tc.name, tc.args...); got != tc.want {
			t.Errorf("%s(%v) = %#v, want %#v", tc.name, tc.args, got, tc.want)
		}
	}
	if msg := raises(func() { callNative("binascii.unhexlify", Bytes("abc")) }); msg != "binascii_pm_Error: Odd-length string" {
		t.Errorf("odd-length error = %q", msg)
	}
	if msg := raises(func() { callNative("binascii.unhexlify", Bytes("zz")) }); msg != "binascii_pm_Error: Non-hexadecimal digit found" {
		t.Errorf("bad digit error = %q", msg)
	}
}

func TestZlibRoundTrip(t *testing.T) {
	data := Bytes("the quick brown fox jumps over the lazy dog, again and again and again")
	for _, wbits := range []int64{15, -15, 31} {
		packed := callNative("zlib.compress", data, int64(6), wbits).(Bytes)
		if len(packed) == 0 {
			t.Fatalf("wbits %d: empty output", wbits)
		}
		if got := callNative("zlib.decompress", packed, wbits); got != data {
			t.Errorf("wbits %d: round trip = %q", wbits, got)
		}
	}
	if got := callNative("zlib.adler32", Bytes("Wikipedia")); got != int64(0x11E60398) {
		t.Errorf("adler32 = %v", got)
	}
	if msg := raises(func() { callNative("zlib.decompress", Bytes("not zlib data")) }); msg != "zlib_pm_error: Error -3 while decompressing data: incorrect header check" {
		t.Errorf("header error = %q", msg)
	}
	truncated := callNative("zlib.compress", data).(Bytes)
	if msg := raises(func() { callNative("zlib.decompress", truncated[:len(truncated)-6]) }); msg != "zlib_pm_error: Error -5 while decompressing data: incomplete or truncated stream" {
		t.Errorf("truncated error = %q", msg)
	}
}

func TestBlake2AndPbkdf2(t *testing.T) {
	h := callNative("hashlib.pbkdf2_hmac", "sha256", Bytes("password"), Bytes("salt"), int64(1), int64(32))
	want := "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b"
	if got := fmt.Sprintf("%x", string(h.(Bytes))); got != want {
		t.Errorf("pbkdf2 = %s, want %s", got, want)
	}
}

func TestPickleRegistry(t *testing.T) {
	RegisterGlobal("mymod", "thing", any(func() int64 { return 1 }))
	if v := callNative("pickle.find_global", "mymod", "thing"); v == nil {
		t.Error("registered global not found")
	}
	if got := callNative("pickle.find_global", "builtins", "int"); got != BuiltinClassOf("int") {
		t.Errorf("builtins.int = %v", got)
	}
	if msg := raises(func() { callNative("pickle.find_global", "nosuchmod", "x") }); msg != "ModuleNotFoundError: No module named 'nosuchmod'" {
		t.Errorf("unknown module = %q", msg)
	}
	if msg := raises(func() { callNative("pickle.find_global", "mymod", "other") }); msg != "AttributeError: Can't get attribute 'other' on <module 'mymod'>" {
		t.Errorf("unknown name = %q", msg)
	}
	if callNative("pickle.is_class", BuiltinClassOf("int")) != true || callNative("pickle.is_class", int64(3)) != false {
		t.Error("is_class")
	}
}

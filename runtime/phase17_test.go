package runtime

import "testing"

func TestSocketpairRoundTrip(t *testing.T) {
	pair := callNative("_socket.socketpair", int64(1), int64(1), int64(0)).(Seq)
	a, b := pair.At(0), pair.At(1)
	defer callNative("_socket.close", a)
	defer callNative("_socket.close", b)
	if n := callNative("_socket.send", a, Bytes("ping")); n != int64(4) {
		t.Fatalf("send = %v", n)
	}
	if got := callNative("_socket.recv", b, int64(10)); Repr(got) != "b'ping'" {
		t.Errorf("recv = %s", Repr(got))
	}
}

func TestSocketRecvTimeout(t *testing.T) {
	pair := callNative("_socket.socketpair", int64(1), int64(1), int64(0)).(Seq)
	a, b := pair.At(0), pair.At(1)
	defer callNative("_socket.close", a)
	defer callNative("_socket.close", b)
	defer func() {
		r := recover()
		e, ok := r.(*PyError)
		if !ok || e.Type != "TimeoutError" {
			t.Errorf("recover = %v", r)
		}
	}()
	callNative("_socket.set_nonblock", a, true)
	callNative("_socket.recv", a, int64(1), int64(0), 0.05)
}

func TestInetConversions(t *testing.T) {
	packed := callNative("_socket.inet_aton", "127.0.0.1")
	if Repr(packed) != `b'\x7f\x00\x00\x01'` {
		t.Errorf("inet_aton = %s", Repr(packed))
	}
	if got := callNative("_socket.inet_ntop", int64(2), packed); got != "127.0.0.1" {
		t.Errorf("inet_ntop = %v", got)
	}
}

func TestSetOperatorsAcrossItemTypes(t *testing.T) {
	a := NewSet[int64]()
	a.Add(1)
	a.Add(2)
	a.Add(3)
	b := NewSet[any]()
	b.Add(int64(2))
	got := DynBinary("Sub", a, b)
	if Repr(got) != "{1, 3}" {
		t.Errorf("set - set = %s", Repr(got))
	}
}

package runtime

import (
	"testing"
	"time"
)

func TestBaseTypeNameStripsModuleAndOuterClass(t *testing.T) {
	for in, want := range map[string]string{
		"Message":                      "Message",
		"mailbox_pm_Message":           "Message",
		"email__message_pm_Message":    "Message",
		"Outer_nested_Inner":           "Inner",
		"pkg__mod_pm_Outer_nested_Inn": "Inn",
	} {
		if got := BaseTypeName(in); got != want {
			t.Errorf("BaseTypeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDynShallowKeepsTheItems(t *testing.T) {
	inner := NewList[string]("a")
	outer := NewList[*List[string]](inner)
	got := DynShallow(outer).(*List[any])
	if got.Get(0) != any(inner) {
		t.Errorf("the inner list was copied")
	}
	d := NewDict[string, int64]()
	d.Set("k", 1)
	if _, ok := DynShallow(d).(*Dict[any, any]); !ok {
		t.Errorf("a dict should stay a dict")
	}
}

func TestCmathErrorsAreCPythons(t *testing.T) {
	cases := []struct {
		want string
		f    func()
	}{
		{"ValueError: math domain error", func() { callNative("cmath.log", complex(0, 0)) }},
		{"OverflowError: math range error", func() { callNative("cmath.exp", complex(1000, 0)) }},
		{"ValueError: math domain error", func() { callNative("cmath.atanh", float64(1)) }},
	}
	for _, c := range cases {
		if got := catch(c.f); got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
	if got := callNative("cmath.sqrt", float64(-1)); got != complex(0, 1) {
		t.Errorf("sqrt(-1) = %v", got)
	}
}

func TestCertificateTimesAreOpenSSLs(t *testing.T) {
	for in, want := range map[string]string{
		"2024-01-01T00:00:00Z": "Jan  1 00:00:00 2024 GMT",
		"2124-12-31T23:59:59Z": "Dec 31 23:59:59 2124 GMT",
	} {
		tm, err := time.Parse(time.RFC3339, in)
		if err != nil {
			t.Fatal(err)
		}
		if got := asn1Time(tm); got != want {
			t.Errorf("asn1Time(%s) = %q, want %q", in, got, want)
		}
	}
}

package codegen

import (
	"strings"
	"testing"
)

// mro builds a class graph from name -> bases (all earlier in the list) and
// returns the linearization of the last class as names.
func mro(t *testing.T, classes [][]string) (string, error) {
	t.Helper()
	byName := map[string]*classInfo{}
	object := &classInfo{name: "object", builtin: true}
	object.mro = []*classInfo{object}
	var last *classInfo
	for _, spec := range classes {
		c := &classInfo{name: spec[0], members: map[string]*member{}}
		for _, b := range spec[1:] {
			c.bases = append(c.bases, byName[b])
		}
		if len(c.bases) == 0 {
			c.bases = []*classInfo{object}
		}
		m, err := c3(c, 1)
		if err != nil {
			return "", err
		}
		c.mro = m
		byName[c.name] = c
		last = c
	}
	var names []string
	for _, k := range last.mro {
		names = append(names, k.name)
	}
	return strings.Join(names, " "), nil
}

func TestC3(t *testing.T) {
	// The expected orders come from CPython's C3 implementation.
	tests := []struct {
		name    string
		classes [][]string
		want    string
	}{
		{"single", [][]string{{"A"}, {"B", "A"}, {"C", "B"}}, "C B A object"},
		{"diamond", [][]string{{"A"}, {"B", "A"}, {"C", "A"}, {"D", "B", "C"}}, "D B C A object"},
		{"diamond reversed", [][]string{{"A"}, {"B", "A"}, {"C", "A"}, {"D", "C", "B"}}, "D C B A object"},
		{"wikipedia", [][]string{{"O"}, {"A", "O"}, {"B", "O"}, {"C", "O"}, {"D", "O"}, {"E", "O"},
			{"K1", "A", "B", "C"}, {"K2", "D", "B", "E"}, {"K3", "D", "A"}, {"Z", "K1", "K2", "K3"}}, "Z K1 K2 K3 D A B C E O object"},
		{"mixins", [][]string{{"M1"}, {"M2"}, {"A", "M1"}, {"B", "M2", "A"}}, "B M2 A M1 object"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := mro(t, tc.classes)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
	t.Run("inconsistent", func(t *testing.T) {
		_, err := mro(t, [][]string{{"A"}, {"B", "A"}, {"C", "A", "B"}})
		if err == nil || !strings.Contains(err.Error(), "consistent method resolution order") {
			t.Errorf("got %v", err)
		}
	})
}

func TestGenerateClassOutput(t *testing.T) {
	const point = "class Point:\n    def __init__(self, x: int) -> None:\n        self.x = x\n\n    def get(self) -> int:\n        return self.x\n\n\n"
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"interface, struct and constructor", point + "p: Point = Point(1)\n", []string{
			"type Point interface {", "type _pyc_Point struct {", "func _pyNew_Point(x int64) Point {", "M_get() int64",
			"Get_x() int64", "Set_x(v int64)", "var p Point", "p = _pyNew_Point(1)"}, nil},
		{"methods are compiled per class", point + "class Sub(Point):\n    pass\n", []string{
			"func (self *_pyc_Point) M_get__Point() int64", "func (self *_pyc_Sub) M_get__Point() int64",
			"func (self *_pyc_Sub) M___init____Point(x int64)", "Is_Point()", "Is_Sub()"}, nil},
		{"methods guard the recursion depth", point, []string{"M_get__Point() int64 {\nruntime.Enter()\ndefer runtime.Leave()"}, nil},
		{"attribute access goes through accessors", point + "p: Point = Point(1)\nprint(p.x)\np.x = 2\n", []string{"p.Get_x()", "p.Set_x(2)"}, nil},
		{"special methods return implemented", "class V:\n    def __add__(self, o: 'V') -> 'V':\n        return NotImplemented\n", []string{
			"M___add__(o V) (V, bool)", "return *new(V), false"}, nil},
		{"operators try both operands", "class V:\n    def __add__(self, o: 'V') -> 'V':\n        return o\n\n    def __radd__(self, o: int) -> 'V':\n        return self\n\n\nv: V = V()\nw: V = 1 + v\n", []string{
			"_r.M___radd__(_l)", "runtime.UnsupportedOperand(\"+\", \"int\", \"V\")"}, nil},
		{"zero-sized structs get a field", "class E:\n    pass\n", []string{"_pyid byte"}, nil},
		{"isinstance is nominal", "class A:\n    pass\n\n\nclass B(A):\n    pass\n\n\na: A = A()\nprint(isinstance(a, B))\n", []string{"runtime.InstanceOf[B](a)", "Is_A()", "Is_B()"}, nil},
		{"isinstance narrows in the branch", "class A:\n    pass\n\n\nclass B(A):\n    def f(self) -> int:\n        return 1\n\n\ndef g(a: A) -> int:\n    if isinstance(a, B):\n        return a.f()\n    return 0\n", []string{"var a B = a.(B)"}, nil},
		{"class attributes are package variables", "class C:\n    n: int = 3\n", []string{"var _pyca_C_n int64", "_pyca_C_n = 3"}, nil},
		{"a class attribute set through an instance is shadowed", "class C:\n    n: int = 3\n\n    def set(self) -> None:\n        self.n = 4\n", []string{"f_n ", "s_n ", "self.s_n = true"}, nil},
		{"exception classes register their hierarchy", "class E(ValueError):\n    pass\n", []string{`runtime.RegisterClass("E", "E", "ValueError", "Exception", "BaseException")`, "runtime.ExcBase"}, nil},
		{"raise of a user exception", "class E(Exception):\n    pass\n\n\nraise E('x')\n", []string{"panic(_pyNew_E(runtime.NewVarTuple[any](string(\"x\"))).PyExc())"}, nil},
		{"handlers bind the class", "class E(Exception):\n    pass\n\n\ntry:\n    pass\nexcept E as e:\n    print(e)\n", []string{"var e E = _py1_exc.Obj.(E)"}, nil},
		{"dataclass generates init and eq", "from dataclasses import dataclass\n\n\n@dataclass\nclass D:\n    a: int\n    b: str = 'x'\n", []string{
			"M___init____D(a int64, b string)", "M___eq____D(other any) (bool, bool)", "M___repr____D() string", "var _pyd_D___init___0 string"}, nil},
		{"frozen dataclass setters panic", "from dataclasses import dataclass\n\n\n@dataclass(frozen=True)\nclass D:\n    a: int\n", []string{`Set_a(v int64) { panic(runtime.FrozenInstance("a")) }`}, nil},
		{"function named main", "def main() -> None:\n    print(1)\n\n\nmain()\n", []string{"func _pyf_main()", "_pyf_main()"}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := generate(t, tc.src)
			if err != nil {
				t.Fatal(err)
			}
			got = strings.ReplaceAll(got, "\t", "")
			for _, w := range tc.want {
				if !strings.Contains(got, w) {
					t.Errorf("output should contain %q\n%s", w, got)
				}
			}
			for _, w := range tc.notWant {
				if strings.Contains(got, w) {
					t.Errorf("output should not contain %q\n%s", w, got)
				}
			}
		})
	}
}

func TestClassErrors(t *testing.T) {
	tests := []struct {
		name, src, want string
	}{
		{"nested class", "class A:\n    class B:\n        pass\n", "class defined inside a class at line 2"},
		{"base defined later", "class B(A):\n    pass\n\n\nclass A:\n    pass\n", "base class 'A' used before its definition at line 1"},
		{"unknown base", "class B(Foo):\n    pass\n", "unknown base class 'Foo' at line 1"},
		{"metaclass", "class B(metaclass=type):\n    pass\n", "metaclasses arrive in phase 9"},
		{"method without self", "class A:\n    def f() -> None:\n        pass\n", "method 'f' without a self parameter at line 2"},
		{"unknown attribute", "class A:\n    pass\n\n\na: A = A()\nprint(a.x)\n", "attribute 'x' of class 'A'"},
		{"constructor arguments", "class A:\n    pass\n\n\na: A = A(1)\n", "A() takes no arguments at line 5"},
		{"method arity counts self", "class A:\n    def f(self, x: int) -> int:\n        return x\n\n\nA().f(1, 2)\n", "A.f() takes 2 positional arguments but 3 were given"},
		{"override with another signature", "class A:\n    def f(self, x: int) -> int:\n        return x\n\n\nclass B(A):\n    def f(self, x: str) -> int:\n        return 0\n", "'B.f' overrides 'A.f' with a different signature"},
		{"hash of a class with __eq__", "class A:\n    def __eq__(self, o: object) -> bool:\n        return True\n\n\nd: dict[A, int] = {}\n", "dict key type A (not hashable)"},
		{"class and function share a name", "class f:\n    pass\n\n\ndef f() -> None:\n    pass\n", "'f' is both a function and a class"},
		{"raise a plain class", "class A:\n    pass\n\n\nraise A()\n", "exceptions must derive from BaseException ('A' does not)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := generate(t, tc.src)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %v, want an error containing %q", err, tc.want)
			}
		})
	}
}

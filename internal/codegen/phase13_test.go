package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase13Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"an async def is a coroutine", "async def f(x):\n    return x\n", []string{"runtime.NewGen[any, any, any]", `.AsCoro("f")`}, nil},
		{"the result of an async def is a generator of objects", "async def f() -> int:\n    return 1\n\n\ndef g():\n    return f()\n", []string{"runtime.NewGen[any, any, int64]", `.AsCoro("f")`}, nil},
		{"await of a coroutine delegates to it", "async def f():\n    return 1\n\n\nasync def g():\n    return await f()\n", []string{"runtime.YieldFromGen(_pyy, f())"}, []string{"AwaitObj"}},
		{"await of an object asks for __await__", "async def g(x):\n    return await x\n", []string{"runtime.AwaitObj(_pyy, x)"}, nil},
		{"async for awaits __anext__", "async def g(xs):\n    async for x in xs:\n        print(x)\n", []string{"__aiter__", "__anext__", "StopAsyncIteration"}, nil},
		{"async with awaits __aenter__ and __aexit__", "async def g(m):\n    async with m as v:\n        print(v)\n", []string{"__aenter__", "__aexit__"}, nil},
		{"an async generator wraps its yields", "async def g():\n    yield 1\n", []string{"runtime.NewAsyncGen(", "runtime.AGYield(_pyy, "}, nil},
		{"an async comprehension is a loop in the coroutine", "async def g(xs):\n    return [x async for x in xs]\n", []string{"__anext__"}, nil},
		{"the signature of an async def says it is async", "async def f(a, b=1):\n    return a\n\n\ng = f\n", []string{"Async: true"}, nil},
		{"a loop yields the interpreter lock when threads are used", "x = _native(\"thread.get_ident\")\nwhile x > 0:\n    x -= 1\n", []string{"runtime.Yield()"}, nil},
		{"a loop of a program without threads does not", "x = 5\nwhile x > 0:\n    x -= 1\n", nil, []string{"runtime.Yield()"}},
		{"a string formatted with an instance", "class A:\n    pass\n\n\nprint('%s' % A())\n", []string{`runtime.DynBinary("Mod"`}, []string{"UnsupportedOperand"}},
		{"star arguments with a keyword", "def f(*args, key=0):\n    return key\n\n\nxs = [1, 2]\nprint(f(*xs, key=3))\n", []string{"key"}, nil},
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
					t.Errorf("missing %q in:\n%s", w, got)
				}
			}
			for _, w := range tc.notWant {
				if strings.Contains(got, w) {
					t.Errorf("unexpected %q in:\n%s", w, got)
				}
			}
		})
	}
}

func TestGeneratePhase13Rejections(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"await outside async", "def f(x):\n    return await x\n", "unsupported: 'await' outside async function at line 2"},
		{"async generator expression", "async def g(xs):\n    return (x async for x in xs)\n", "unsupported: async generator expression at line 2"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := generate(t, tc.src)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want an error with %q, got %v", tc.want, err)
			}
		})
	}
}

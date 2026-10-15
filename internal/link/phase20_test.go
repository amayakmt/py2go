package link

import "testing"

func TestModuleUsedAsAValue(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import util\n\n\ndef f(m):\n    return hasattr(m, 'X')\n\n\nprint(f(util))\n", "util.py": "X = 1\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	if !got["_native"] || !got["util_pm_X"] {
		t.Errorf("a module used as a value should become _native(\"module.object\") with its names registered: %v", got)
	}
}

func TestTypingQualifiedNames(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import typing\nimport dataclasses\n\n\n@dataclasses.dataclass\nclass P:\n    x: typing.List[int]\n\n\ndef f(a: typing.Optional[int]) -> typing.Dict[str, int]:\n    return {}\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"List", "Optional", "Dict", "dataclass"} {
		if !got[want] {
			t.Errorf("%s should replace typing.%s", want, want)
		}
	}
	if got["typing"] || got["dataclasses"] {
		t.Errorf("the module names should be gone")
	}
}

package link

import (
	"testing"

	"github.com/amayakmt/py2go/internal/pyast"
)

// keywordArgs collects the names of the keywords of every call in the module.
func keywordArgs(mod *pyast.Module) map[string]bool {
	out := map[string]bool{}
	var visit func(n pyast.Node)
	visit = func(n pyast.Node) {
		if c, ok := n.(*pyast.Call); ok {
			for _, k := range c.Keywords {
				out[k.Arg] = true
			}
		}
		walkChildren(n, visit)
	}
	for _, st := range mod.Body {
		visit(st)
	}
	return out
}

func TestUnittestMainGetsTheTestClasses(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import unittest\n\n\nclass A(unittest.TestCase):\n    def test_a(self):\n        pass\n\n\nclass B(A):\n    pass\n\n\nclass C:\n    pass\n\n\nunittest.main()\n"})
	if err != nil {
		t.Fatal(err)
	}
	if !keywordArgs(mod)["_classes"] {
		t.Errorf("unittest.main() did not get _classes")
	}
}

func TestUnittestDecoratorsBecomeClassAttributes(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import unittest\n\n\nclass A(unittest.TestCase):\n    @unittest.skip('r')\n    def test_a(self):\n        pass\n\n    @unittest.skipIf(True, 'why')\n    def test_b(self):\n        pass\n\n    @unittest.expectedFailure\n    def test_c(self):\n        pass\n\n\n@unittest.skipUnless(False, 'class')\nclass D(unittest.TestCase):\n    pass\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"_unittest_skip_test_a", "_unittest_skip_test_b", "_unittest_xfail_test_c", "_unittest_skip_class"} {
		if !got[want] {
			t.Errorf("%s was not generated", want)
		}
	}
}

func TestLinkJoinsTheDeveloperModules(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import unittest\nimport unittest.mock\nimport difflib\nimport pprint\nimport inspect\nimport token\n\nprint(unittest.TestCase, unittest.mock.Mock, difflib.unified_diff, pprint.pformat, inspect.signature, token.NAME)\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"unittest__case_pm_TestCase", "unittest__mock_pm_Mock", "difflib_pm_unified_diff", "pprint_pm_pformat", "inspect_pm_signature", "token_pm_NAME"} {
		if !got[want] {
			t.Errorf("%s is not in the joined module", want)
		}
	}
}

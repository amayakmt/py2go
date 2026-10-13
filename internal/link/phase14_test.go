package link

import "testing"

func TestLinkJoinsTheNumericModules(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import decimal\nimport fractions\nimport statistics\nimport struct\nimport unicodedata\nimport array\nimport codecs\nimport glob\nimport shlex\nimport getopt\nimport secrets\nimport colorsys\nimport fnmatch\n\nprint(decimal.Decimal(1), fractions.Fraction(1, 3), statistics.mean([1]), struct.calcsize('i'), unicodedata.category('a'), array.array('i'), codecs.lookup('utf-8'), glob.glob('x'), shlex.split('a'), getopt.getopt([], ''), secrets.token_hex(1), colorsys.rgb_to_hsv(0, 0, 0), fnmatch.fnmatch('a', 'a'))\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"decimal_pm_Decimal", "fractions_pm_Fraction", "statistics_pm_mean", "struct_pm_calcsize", "unicodedata_pm_category", "array_pm_array", "codecs_pm_lookup", "glob_pm_glob", "shlex_pm_split", "getopt_pm_getopt", "secrets_pm_token_hex", "colorsys_pm_rgb_to_hsv", "fnmatch_pm_fnmatch"} {
		if !got[want] {
			t.Errorf("%s is not in the joined module", want)
		}
	}
}

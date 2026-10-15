package link

import "testing"

func TestLinkJoinsTheSystemModules(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import datetime\nimport calendar\nimport subprocess\nimport pathlib\nimport tempfile\nimport shutil\nimport logging\nimport logging.handlers\nimport logging.config\nimport zoneinfo\nimport tomllib\nimport zipfile\nimport tarfile\nimport shelve\nimport dbm\nimport wave\nimport mimetypes\nimport netrc\nimport filecmp\nimport fileinput\nimport errno\nimport stat\nimport pwd\nimport grp\nimport platform\nimport getpass\nimport timeit\nimport warnings\nimport resource\nimport select\nimport fcntl\n\nprint(datetime.date(2020, 1, 1), calendar.isleap(2020), subprocess.PIPE, pathlib.Path('.'), tempfile.gettempdir(), shutil.which('sh'), logging.INFO, logging.handlers.MemoryHandler, logging.config.dictConfig, zoneinfo.ZoneInfo, tomllib.loads(''), zipfile.is_zipfile('x'), tarfile.is_tarfile('x'), shelve.open, dbm.open, wave.open, mimetypes.guess_type('a.txt'), netrc.netrc, filecmp.cmp, fileinput.input, errno.ENOENT, stat.S_ISDIR(0), pwd.getpwall(), grp.getgrall(), platform.system(), getpass.getuser(), timeit.default_timer, warnings.warn, resource.getpagesize(), select.select, fcntl.LOCK_EX)\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"datetime_pm_date", "calendar_pm_isleap", "subprocess_pm_PIPE", "pathlib_pm_Path", "tempfile_pm_gettempdir", "shutil_pm_which", "logging_pm_INFO", "logging__handlers_pm_MemoryHandler", "logging__config_pm_dictConfig", "zoneinfo_pm_ZoneInfo", "tomllib___parser_pm_loads", "zipfile_pm_is_zipfile", "tarfile_pm_is_tarfile", "shelve_pm_open", "dbm_pm_open", "wave_pm_open", "mimetypes_pm_guess_type", "netrc_pm_netrc", "filecmp_pm_cmp", "fileinput_pm_input", "errno_pm_ENOENT", "stat_pm_S_ISDIR", "pwd_pm_getpwall", "platform_pm_system"} {
		if !got[want] {
			t.Errorf("%s is not in the joined module", want)
		}
	}
}

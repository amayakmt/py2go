"""Unittest main program (reduced: no argparse, no module objects)."""
import sys
from . import loader, runner
from .signals import installHandler
__unittest = True
MAIN_EXAMPLES = ''


def _obj(x: object) -> object:
    return x


def _convert_name(name):
    return name


class TestProgram(object):
    """A command-line program that runs a set of tests; this is primarily
       for making test modules conveniently executable.
    """

    def __init__(self, module='__main__', defaultTest=None, argv=None, testRunner=None, testLoader=loader.defaultTestLoader, exit=True, verbosity=1, failfast=None, catchbreak=None, buffer=None, warnings=None, *, tb_locals=False, durations=None, _classes=None):
        self.module = _obj(module)
        self.exit = exit
        self.failfast = failfast
        self.catchbreak = catchbreak
        self.verbosity = verbosity
        self.buffer = buffer
        self.tb_locals = tb_locals
        self.durations = _obj(durations)
        self.warnings = _obj(warnings)
        self.defaultTest = _obj(defaultTest)
        self.testRunner = _obj(testRunner)
        self.testLoader = testLoader
        self.progName = 'python -m unittest'
        self.test = _obj(None)
        self.result = _obj(None)
        self._classes = _obj(_classes)
        self.testNames = _obj(None)
        if argv is None:
            argv = sys.argv
        self.parseArgs(argv)
        self.runTests()

    def parseArgs(self, argv):
        names = []
        for arg in argv[1:]:
            if arg == '-v' or arg == '--verbose':
                self.verbosity = 2
            elif arg == '-q' or arg == '--quiet':
                self.verbosity = 0
            elif arg == '-f' or arg == '--failfast':
                self.failfast = True
            elif arg == '-b' or arg == '--buffer':
                self.buffer = True
            elif arg == '-c' or arg == '--catch':
                self.catchbreak = True
            elif arg.startswith('-'):
                pass
            else:
                names.append(arg)
        if len(names) > 0:
            self.testNames = names
            self.test = self.testLoader.loadTestsFromNames(names, None, self._classes)
        elif self.defaultTest is not None:
            if isinstance(self.defaultTest, str):
                self.testNames = [self.defaultTest]
            else:
                self.testNames = list(self.defaultTest)
            self.test = self.testLoader.loadTestsFromNames(self.testNames, None, self._classes)
        else:
            self.testNames = None
            self.test = self.testLoader.loadTestsFromModule(self.module, _classes=self._classes)

    def runTests(self):
        if self.catchbreak:
            installHandler()
        if self.testRunner is None:
            self.testRunner = runner.TextTestRunner
        if isinstance(self.testRunner, type):
            try:
                testRunner = self.testRunner(verbosity=self.verbosity, failfast=self.failfast, buffer=self.buffer, warnings=self.warnings, tb_locals=self.tb_locals, durations=self.durations)
            except TypeError:
                testRunner = self.testRunner()
        else:
            testRunner = self.testRunner
        self.result = testRunner.run(self.test)
        if self.exit:
            sys.exit(not self.result.wasSuccessful())


main = TestProgram

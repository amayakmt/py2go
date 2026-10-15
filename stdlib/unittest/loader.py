"""Loading unittests (reduced: no discovery, no module objects).

The linker gives `loadTestsFromModule` and `unittest.main` the test case
classes of the entry module as `_classes`, since a running program has no
module objects to scan."""
import re
import sys
from fnmatch import fnmatchcase
from . import case, suite, util
__unittest = True


def _obj(x: object) -> object:
    return x


class _FailedTest(case.TestCase):

    def __init__(self, method_name, exception):
        self._exception = _obj(exception)
        case.TestCase.__init__(self, 'runTest')
        self._failed_name = method_name

    def runTest(self):
        raise self._exception

    def id(self):
        return 'unittest.loader._FailedTest.' + self._failed_name

    def __str__(self):
        return '%s (unittest.loader._FailedTest.%s)' % (self._failed_name, self._failed_name)


def _make_failed_test(methodname, exception, suiteClass, message):
    test = _FailedTest(methodname, exception)
    return (suiteClass((test,)), message)


class TestLoader(object):
    """
    This class is responsible for loading tests according to various criteria
    and returning them wrapped in a TestSuite
    """

    def __init__(self):
        self.testMethodPrefix = 'test'
        self.sortTestMethodsUsing = True
        self.testNamePatterns = _obj(None)
        self.suiteClass = suite.TestSuite
        self.errors = []

    def loadTestsFromTestCase(self, testCaseClass):
        """Return a suite of all test cases contained in testCaseClass"""
        if issubclass(testCaseClass, suite.TestSuite):
            raise TypeError('Test cases should not be derived from TestSuite. Maybe you meant to derive from TestCase?')
        if testCaseClass is case.TestCase or testCaseClass is case.FunctionTestCase:
            testCaseNames = []
        else:
            testCaseNames = self.getTestCaseNames(testCaseClass)
            if not testCaseNames and hasattr(testCaseClass, 'runTest'):
                testCaseNames = ['runTest']
        loaded_suite = self.suiteClass([testCaseClass(n) for n in testCaseNames])
        return loaded_suite

    def loadTestsFromModule(self, module=None, *, pattern=None, _classes=None):
        """Return a suite of all test cases contained in the given module"""
        tests = []
        if _classes is not None:
            for obj in _classes:
                if obj is not case.TestCase and obj is not case.FunctionTestCase:
                    tests.append(self.loadTestsFromTestCase(obj))
        return self.suiteClass(tests)

    def loadTestsFromName(self, name, module=None, _classes=None):
        """Return a suite of all test cases given a string specifier,
        naming a test case class or a method of one of the classes passed
        as `_classes`."""
        parts = name.split('.')
        if parts[0] == '__main__' or (_classes is not None and len(parts) > 0 and parts[0] not in [c.__name__ for c in _classes]):
            parts = parts[1:]
        if _classes is not None and len(parts) > 0:
            for cls in _classes:
                if cls.__name__ == parts[0]:
                    if len(parts) == 1:
                        return self.loadTestsFromTestCase(cls)
                    if len(parts) == 2 and hasattr(cls, parts[1]):
                        return self.suiteClass([cls(parts[1])])
        error_case, error_message = _make_failed_test(name, ImportError("No module named '%s'" % name), self.suiteClass, 'Failed to import test module: %s' % name)
        self.errors.append(error_message)
        return error_case

    def loadTestsFromNames(self, names, module=None, _classes=None):
        """Return a suite of all test cases found using the given sequence
        of string specifiers. See 'loadTestsFromName()'.
        """
        suites = [self.loadTestsFromName(name, module, _classes) for name in names]
        return self.suiteClass(suites)

    def getTestCaseNames(self, testCaseClass):
        """Return a sorted sequence of method names found within testCaseClass
        """
        testFnNames = []
        for attrname in dir(testCaseClass):
            if not attrname.startswith(self.testMethodPrefix):
                continue
            testFunc = getattr(testCaseClass, attrname)
            if not callable(testFunc):
                continue
            fullName = '%s.%s.%s' % (testCaseClass.__module__, testCaseClass.__name__, attrname)
            if self.testNamePatterns is None or any((fnmatchcase(fullName, pattern) for pattern in self.testNamePatterns)):
                testFnNames.append(attrname)
        if self.sortTestMethodsUsing:
            testFnNames.sort()
        return testFnNames

    def discover(self, start_dir, pattern='test*.py', top_level_dir=None):
        """Test discovery needs the import system at run time."""
        raise NotImplementedError('test discovery is not supported')


defaultTestLoader = TestLoader()


def getTestCaseNames(testCaseClass, prefix, sortUsing=True, testNamePatterns=None):
    loader = TestLoader()
    loader.sortTestMethodsUsing = sortUsing
    loader.testMethodPrefix = prefix
    loader.testNamePatterns = testNamePatterns
    return loader.getTestCaseNames(testCaseClass)


def findTestCases(module, prefix='test', sortUsing=True, suiteClass=suite.TestSuite, _classes=None):
    loader = TestLoader()
    loader.sortTestMethodsUsing = sortUsing
    loader.testMethodPrefix = prefix
    loader.suiteClass = suiteClass
    return loader.loadTestsFromModule(module, _classes=_classes)

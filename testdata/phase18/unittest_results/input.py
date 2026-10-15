import io
import re
import unittest

log = []


class Fixtures(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        log.append("setUpClass")

    @classmethod
    def tearDownClass(cls):
        log.append("tearDownClass")

    def setUp(self):
        log.append("setUp " + self._testMethodName)

    def tearDown(self):
        log.append("tearDown " + self._testMethodName)

    def test_pass(self):
        self.addCleanup(log.append, "cleanup")
        self.assertEqual(1 + 1, 2)

    def test_fail(self):
        self.assertEqual([1, 2, 3], [1, 2, 4])

    def test_error(self):
        raise KeyError("boom")

    @unittest.skip("not today")
    def test_skipped(self):
        self.fail("should not run")

    @unittest.skipIf(True, "cond")
    def test_skip_if(self):
        self.fail("no")

    @unittest.skipUnless(True, "cond")
    def test_skip_unless_ok(self):
        pass

    @unittest.expectedFailure
    def test_xfail(self):
        self.assertTrue(False)

    @unittest.expectedFailure
    def test_xpass(self):
        pass

    def test_skiptest(self):
        self.skipTest("runtime skip")

    def test_subtests(self):
        for i in range(3):
            with self.subTest(i=i):
                self.assertNotEqual(i, 1)


class Assertions(unittest.TestCase):
    def check(self, fn):
        try:
            fn()
        except AssertionError as e:
            return str(e)
        return "no failure"

    def test_messages(self):
        msgs = [
            self.check(lambda: self.assertEqual(1, 2)),
            self.check(lambda: self.assertEqual("a", "b", "custom")),
            self.check(lambda: self.assertNotEqual(3, 3)),
            self.check(lambda: self.assertTrue(0)),
            self.check(lambda: self.assertFalse(1)),
            self.check(lambda: self.assertIs(1, 2)),
            self.check(lambda: self.assertIsNone(3)),
            self.check(lambda: self.assertIn(5, [1, 2])),
            self.check(lambda: self.assertNotIn(1, [1, 2])),
            self.check(lambda: self.assertIsInstance(1, str)),
            self.check(lambda: self.assertGreater(1, 2)),
            self.check(lambda: self.assertLessEqual(3, 2)),
            self.check(lambda: self.assertAlmostEqual(1.0, 1.5)),
            self.check(lambda: self.assertListEqual([1, 2], [1, 3])),
            self.check(lambda: self.assertDictEqual({"a": 1}, {"a": 2})),
            self.check(lambda: self.assertSetEqual({1, 2}, {2, 3})),
            self.check(lambda: self.assertTupleEqual((1,), (1, 2))),
            self.check(lambda: self.assertMultiLineEqual("a\nb", "a\nc")),
            self.check(lambda: self.assertRegex("hello", "xyz")),
            self.check(lambda: self.assertCountEqual([1, 2, 2], [1, 2])),
            self.check(lambda: self.assertRaises(ValueError, int, "5")),
            self.check(lambda: self.fail("explicit")),
        ]
        for m in msgs:
            print(m)
            print("--")

    def test_raises(self):
        with self.assertRaises(ValueError) as cm:
            int("x")
        print(type(cm.exception).__name__, cm.exception)
        with self.assertRaisesRegex(KeyError, "ke+y"):
            raise KeyError("keeey")
        self.assertRaises(ZeroDivisionError, lambda: 1 / 0)
        self.assertAlmostEqual(1.00000001, 1.0)
        self.assertAlmostEqual(1.0, 1.1, delta=0.2)
        self.assertSequenceEqual([1, 2], (1, 2))
        self.assertCountEqual("abc", "cab")


def run(case):
    stream = io.StringIO()
    suite = unittest.defaultTestLoader.loadTestsFromTestCase(case)
    result = unittest.TextTestRunner(stream=stream, verbosity=2).run(suite)
    out = re.sub(r"in \d+\.\d+s", "in Xs", stream.getvalue())
    return result, out


result, out = run(Fixtures)
print(out.split("======")[0])
print(out.splitlines()[-1])
print(result.testsRun, len(result.failures), len(result.errors), len(result.skipped), len(result.expectedFailures), len(result.unexpectedSuccesses))
for test, msg in result.failures:
    print("FAIL", test.id(), msg.strip().splitlines()[-1])
for test, msg in result.errors:
    print("ERROR", test.id(), msg.strip().splitlines()[-1])
for test, reason in result.skipped:
    print("SKIP", test.id(), reason)
print(log)
result, out = run(Assertions)
print(out.splitlines()[-1], result.wasSuccessful())

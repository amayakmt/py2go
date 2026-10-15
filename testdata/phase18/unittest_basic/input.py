import io
import re
import unittest


class MathTests(unittest.TestCase):
    def setUp(self):
        self.x = 10

    def test_add(self):
        self.assertEqual(self.x + 1, 11)

    def test_true(self):
        self.assertTrue(self.x > 1)
        self.assertFalse(self.x < 1)
        self.assertIn(3, [1, 2, 3])

    def test_raises(self):
        with self.assertRaises(ZeroDivisionError):
            1 / 0


def run(case):
    stream = io.StringIO()
    suite = unittest.defaultTestLoader.loadTestsFromTestCase(case)
    result = unittest.TextTestRunner(stream=stream, verbosity=2).run(suite)
    out = re.sub(r"in \d+\.\d+s", "in Xs", stream.getvalue())
    print(out)
    return result


r = run(MathTests)
print(r.testsRun, r.wasSuccessful(), len(r.failures), len(r.errors))

import unittest


class A(unittest.TestCase):
    def test_one(self):
        print("A.test_one")

    def test_two(self):
        print("A.test_two")
        self.assertEqual(2, 2)


class B(A):
    def test_three(self):
        print("B.test_three")


class NotATest:
    def test_x(self):
        print("never")


if __name__ == "__main__":
    prog = unittest.main(verbosity=0, exit=False)
    print("ran", prog.result.testsRun, prog.result.wasSuccessful())
    suite = unittest.TestLoader().loadTestsFromTestCase(B)
    print(suite.countTestCases(), [str(t) for t in suite])
    unittest.main(argv=["x", "A.test_one"], verbosity=0)

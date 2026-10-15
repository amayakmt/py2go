import unittest


class T(unittest.TestCase):
    def test_ok(self):
        print("ok")

    def test_bad(self):
        self.assertEqual(1, 2)


unittest.main()

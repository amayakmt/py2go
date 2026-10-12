class Version:
    def __init__(self, major, minor):
        self.major = major
        self.minor = minor

    def _key(self):
        return (self.major, self.minor)

    def __eq__(self, other):
        return self._key() == other._key()

    def __lt__(self, other):
        return self._key() < other._key()

    def __le__(self, other):
        return self._key() <= other._key()

    def __hash__(self):
        return hash(self._key())

    def __repr__(self):
        return "v%d.%d" % self._key()

vs = [Version(1, 2), Version(0, 9), Version(1, 0), Version(1, 2)]
print(sorted(vs), min(vs), max(vs))
print(Version(1, 2) == Version(1, 2), Version(1, 0) < Version(1, 2), Version(2, 0) <= Version(1, 9))
print(hash(Version(1, 2)) == hash(Version(1, 2)), hash(Version(1, 2)) == hash(Version(2, 1)))
print(sorted(vs, reverse=True)[0], [v for v in vs if v > Version(1, 0)] if False else "n/a")
try:
    print(Version(1, 0) > Version(2, 0))
except TypeError as e:
    print("TypeError", e)

class Version:
    def __init__(self, major: int, minor: int) -> None:
        self.major = major
        self.minor = minor

    def __lt__(self, other: "Version") -> bool:
        if self.major != other.major:
            return self.major < other.major
        return self.minor < other.minor

    def __le__(self, other: "Version") -> bool:
        return self < other or self == other

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, Version):
            return NotImplemented
        return self.major == other.major and self.minor == other.minor

    def __repr__(self) -> str:
        return f"v{self.major}.{self.minor}"


a: Version = Version(1, 2)
b: Version = Version(1, 10)
c: Version = Version(2, 0)
print(a < b, b < a, a <= a, c > b, a >= b)
vs: list[Version] = [c, a, b]
print(sorted(vs), min(vs), max(vs))
vs.sort(reverse=True)
print(vs)

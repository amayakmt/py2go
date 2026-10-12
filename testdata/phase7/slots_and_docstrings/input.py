class Tiny:
    """A tiny class.

    With a longer docstring.
    """

    __slots__ = ("a", "b")

    def __init__(self, a: int, b: int) -> None:
        self.a = a
        self.b = b

    def total(self) -> int:
        """Add the two."""
        return self.a + self.b


t: Tiny = Tiny(1, 2)
print(t.total(), Tiny.__name__, Tiny.__doc__)
print(t.__class__.__name__, type(t).__name__)

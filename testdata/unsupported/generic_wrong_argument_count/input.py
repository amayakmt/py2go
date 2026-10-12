class Pair[A, B]:
    def __init__(self, a: A, b: B) -> None:
        self.a = a
        self.b = b


p: Pair[int] = Pair(1, 2)

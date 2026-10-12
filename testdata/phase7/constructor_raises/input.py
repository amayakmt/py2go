class Positive:
    def __init__(self, n: int) -> None:
        if n <= 0:
            raise ValueError("must be positive, got " + str(n))
        self.n = n


made: list[Positive] = []
for n in [3, 0, -1, 5]:
    try:
        made.append(Positive(n))
    except ValueError as e:
        print("skip:", e)
print([p.n for p in made])

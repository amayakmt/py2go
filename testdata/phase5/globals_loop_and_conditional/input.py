results: list[int] = []
for n in range(4):
    sq: int = n * n
    results.append(sq)
print(results, sq, n)


def last_square() -> int:
    return sq


print(last_square())

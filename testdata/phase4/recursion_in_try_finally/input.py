def walk(n: int, trail: list[int]) -> int:
    try:
        if n == 0:
            return 0
        return walk(n - 1, trail) + 1
    finally:
        trail.append(n)


trail: list[int] = []
print(walk(5, trail), trail)
big: list[int] = []
print(walk(300, big), len(big))

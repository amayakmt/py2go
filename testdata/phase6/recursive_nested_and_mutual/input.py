def is_even(n: int) -> bool:
    def even(k: int) -> bool:
        if k == 0:
            return True
        return odd(k - 1)

    def odd(k: int) -> bool:
        if k == 0:
            return False
        return even(k - 1)
    return even(n)


print(is_even(10), is_even(7))


def hanoi(n: int, a: str, b: str, c: str, moves: list[str]) -> None:
    if n == 0:
        return
    hanoi(n - 1, a, c, b, moves)
    moves.append(a + c)
    hanoi(n - 1, b, a, c, moves)


ms: list[str] = []
hanoi(3, "A", "B", "C", ms)
print(ms, len(ms))

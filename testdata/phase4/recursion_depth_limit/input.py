def depth(n: int, seen: list[int]) -> None:
    seen[0] = n
    depth(n + 1, seen)


seen: list[int] = [0]
try:
    depth(1, seen)
except RecursionError:
    print("reached", seen[0])

def count_down(n: int) -> int:
    if n == 0:
        return 0
    return 1 + count_down(n - 1)

print(count_down(900))

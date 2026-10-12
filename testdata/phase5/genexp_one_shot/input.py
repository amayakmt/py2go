pulled: list[int] = []


def src(n: int, log: list[int]) -> int:
    log.append(n)
    return n


print(list(src(i, pulled) for i in range(3)))
print(pulled)
print(sum(src(i, pulled) for i in range(3)), pulled)

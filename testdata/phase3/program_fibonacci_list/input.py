def fibs(n: int) -> list[int]:
    out: list[int] = [0, 1]
    while len(out) < n:
        out.append(out[-1] + out[-2])
    return out[:n]


print(fibs(15))
print(fibs(1), fibs(2))

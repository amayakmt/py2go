data: list[int] = [3, 8, 12, 5]
if (n := len(data)) > 3:
    print("many", n)
total: int = 0
i: int = 0
while (v := data[i]) < 10:
    total += v
    i += 1
print(total, v, i)
print([y for x in data if (y := x * 2) > 10])
squares: list[int] = [(k := d * d) + 1 for d in data]
print(squares, k)


def first_big(xs: list[int]) -> int:
    if (m := max(xs)) > 10:
        return m
    return -1


print(first_big(data), first_big([1, 2]))
print((a := 5) + a, a)

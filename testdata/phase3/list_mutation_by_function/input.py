def add_one(xs: list[int]) -> None:
    for i in range(len(xs)):
        xs[i] += 1


def push(xs: list[int], v: int) -> None:
    xs.append(v)


def rebind(xs: list[int]) -> list[int]:
    xs = [0]
    return xs


data: list[int] = [1, 2, 3]
add_one(data)
push(data, 10)
print(data)
other: list[int] = rebind(data)
print(data, other)

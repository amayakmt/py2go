def bubble_sort(xs: list[int]) -> None:
    n: int = len(xs)
    for i in range(n):
        swapped: bool = False
        for j in range(n - 1 - i):
            if xs[j] > xs[j + 1]:
                xs[j], xs[j + 1] = xs[j + 1], xs[j]
                swapped = True
        if not swapped:
            break


data: list[int] = [64, 34, 25, 12, 22, 11, 90, -5, 0]
bubble_sort(data)
print(data)

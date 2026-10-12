def search(xs: list[int], target: int) -> int:
    lo: int = 0
    hi: int = len(xs) - 1
    while lo <= hi:
        mid: int = (lo + hi) // 2
        if xs[mid] == target:
            return mid
        elif xs[mid] < target:
            lo = mid + 1
        else:
            hi = mid - 1
    return -1


data: list[int] = [1, 3, 5, 7, 9, 11, 13]
for t in [7, 1, 13, 4, 14, 0]:
    print(t, search(data, t))

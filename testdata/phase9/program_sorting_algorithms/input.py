def bubble(xs):
    xs = list(xs)
    for i in range(len(xs)):
        for j in range(len(xs) - 1 - i):
            if xs[j] > xs[j + 1]:
                xs[j], xs[j + 1] = xs[j + 1], xs[j]
    return xs

def insertion(xs):
    xs = list(xs)
    for i in range(1, len(xs)):
        key = xs[i]
        j = i - 1
        while j >= 0 and xs[j] > key:
            xs[j + 1] = xs[j]
            j -= 1
        xs[j + 1] = key
    return xs

def merge_sort(xs):
    if len(xs) <= 1:
        return list(xs)
    mid = len(xs) // 2
    left, right = merge_sort(xs[:mid]), merge_sort(xs[mid:])
    out = []
    while left and right:
        out.append(left.pop(0) if left[0] <= right[0] else right.pop(0))
    return out + left + right

def quick(xs):
    if len(xs) <= 1:
        return list(xs)
    pivot, rest = xs[0], xs[1:]
    return quick([x for x in rest if x < pivot]) + [pivot] + quick([x for x in rest if x >= pivot])

def bsearch(xs, target):
    lo, hi = 0, len(xs) - 1
    while lo <= hi:
        mid = (lo + hi) // 2
        if xs[mid] == target:
            return mid
        if xs[mid] < target:
            lo = mid + 1
        else:
            hi = mid - 1
    return -1

data = [5, 2, 9, 1, 5, 6, -3, 0, 2.5]
for f in (bubble, insertion, merge_sort, quick, sorted):
    print(f.__name__, f(data))
print(bubble("hello"), quick(["pear", "apple", "fig"]), merge_sort([]), insertion((3, 1, 2)))
s = sorted(data)
print([bsearch(s, x) for x in (-3, 9, 4, 2.5)])

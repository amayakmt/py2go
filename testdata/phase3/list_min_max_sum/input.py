xs: list[int] = [4, -2, 9, 0]
print(min(xs), max(xs), sum(xs))
fs: list[float] = [0.1, 0.2, 0.3]
print(sum(fs), min(fs), max(fs))
tenth: list[float] = [0.1] * 10
print(sum(tenth))
ss: list[str] = ["pear", "apple", "fig"]
print(min(ss), max(ss))
big: list[float] = [1e100, 1.0, -1e100]
print(sum(big))
empty: list[int] = []
print(sum(empty), len(empty))

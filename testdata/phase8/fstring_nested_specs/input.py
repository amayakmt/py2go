width: int = 10
prec: int = 3
value: float = 3.14159265
name: str = "py"
count: int = 42
print(f"[{value:{width}.{prec}f}]")
print(f"[{name:>{width}}] [{name:<{width}}] [{name:^{width}}]")
print(f"[{count:0{width}d}] [{count:{width}}]")
fill: str = "*"
print(f"[{name:{fill}^{width}}]")
print(f"{value=} {count=}")
print(f"{value:.{prec}f} {width * 2:>{width}}")
items: list[str] = ["a", "bb"]
print(f"{items!r:>{width}}")
print(f"{'nested quotes' + "inside":>{width + 15}}")
print(f"{f'{count}':>6}|{(lambda v: v + 1)(count)}|{count:{'>' + '8'}}")

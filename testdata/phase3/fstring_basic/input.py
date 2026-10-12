name: str = "world"
n: int = 42
x: float = 2.5
ok: bool = True
print(f"hello {name}")
print(f"{n} {x} {ok}")
print(f"{n + 1} {n * 2} {n // 5} {x * 2}")
print(f"{name!r} {name!s} {n!r}")
print(f"{{literal}} {n}")
print(f"{name}{name}")
print(f"")
print(f"no fields")
print(f"{name.upper()} has {len(name)} letters")
print(f"{n=} {x=} {name=}")
print(f"{[1, 2]} {(1, 'a')}")

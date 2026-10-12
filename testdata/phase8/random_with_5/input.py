def f(x: int, s: str, xs: list[int]) -> str:
    out: list[str] = []
    match x:
        case v if v > 5:
            out.append("guard " + str(v))
        case 0:
            out.append("lit")
        case 10:
            out.append("lit")
        case _:
            out.append("wild")
    match s:
        case "stop" | "stop":
            out.append("s-hit")
        case _:
            out.append("s-miss")
    match xs:
        case [] if True:
            out.append(f"empty")
        case [a]:
            out.append(f"one {a}")
        case [a, *m, z]:
            out.append(f"ends {a} {z} {len(m)}")
        case [a, b]:
            out.append(f"two {a} {b}")
        case _:
            out.append("other")
    return ",".join(out)


print(f(7, "stop", []))
print(f(2, "go", [2, 1, 2]))
print(f(5, "stop", [3, 4, 4, 3]))
print(f(10, "a", [2]))

from typing import Optional


def f(x: Optional[int]) -> str:
    match x:
        case None:
            return "none"
        case 0:
            return "zero"
        case n:
            return "value " + str(n)


print(f(None), f(0), f(5))


def g(x: Optional[list[int]]) -> int:
    match x:
        case None:
            return -1
        case [a, *_]:
            return a
        case _:
            return 0


print(g(None), g([4, 5]), g([]))
cmd: str = "move 3"
match cmd.split():
    case ["move", amount]:
        print("moving", int(amount))
    case _:
        print("?")
events: list[tuple[str, int, int]] = [("click", 10, 20), ("key", 0, 0)]
for event in events:
    match event:
        case ("click", px, py):
            print("click at", px, py)
        case (name, _, _):
            print("other", name)

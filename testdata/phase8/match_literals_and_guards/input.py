def classify(n: int) -> str:
    match n:
        case 0:
            return "zero"
        case 1 | 2 | 3:
            return "small"
        case -1:
            return "minus one"
        case x if x < 0:
            return "negative " + str(x)
        case x if x % 2 == 0:
            return "even " + str(x)
        case _:
            return "odd"


print([classify(n) for n in [0, 2, -1, -7, 10, 11]])


def command(s: str) -> str:
    match s:
        case "go" | "run":
            return "move"
        case "stop":
            return "halt"
        case other:
            return "unknown " + other


print(command("go"), command("stop"), command("dance"))
flag: bool = True
match flag:
    case True:
        print("yes")
    case False:
        print("no")
f: float = 2.5
match f:
    case 2.5:
        print("two and a half")
    case _:
        print("other")

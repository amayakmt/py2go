def tokens(src: str) -> list[str]:
    out: list[str] = []
    for part in src.split():
        match part:
            case "+" | "-" | "*" | "/":
                out.append("op:" + part)
            case "(" | ")":
                out.append("paren")
            case _ if part in ["1", "2", "3"]:
                out.append("num:" + part)
            case other:
                out.append("word:" + other)
    return out


print(tokens("1 + x ( 2 * y ) 3"))


def evaluate(op: str, a: int, b: int) -> int:
    match op:
        case "+":
            return a + b
        case "-":
            return a - b
        case "*":
            return a * b
        case "/":
            if b == 0:
                raise ZeroDivisionError("divide by zero in evaluate")
            return a // b
        case _:
            raise ValueError("unknown operator " + op)


for o in ["+", "-", "*", "/", "%"]:
    try:
        print(o, evaluate(o, 12, 4))
    except ValueError as e:
        print("error:", e)
try:
    evaluate("/", 1, 0)
except ZeroDivisionError as e:
    print(e)

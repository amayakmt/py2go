def parse(pair: tuple[str, list[int]]) -> str:
    match pair:
        case ("sum", [*nums]) if (t := sum(nums)) > 10:
            return "big sum " + str(t)
        case ("sum", [*nums]):
            return "sum " + str(sum(nums))
        case ("max", [first, *rest]):
            return "max " + str(max([first] + rest))
        case (cmd, []):
            return cmd + " of nothing"
        case _:
            return "?"


inputs: list[tuple[str, list[int]]] = [("sum", [5, 6]), ("sum", [1, 2]), ("max", [4, 9, 2]), ("min", []), ("x", [1])]
for item in inputs:
    print(parse(item))

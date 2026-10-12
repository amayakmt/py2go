class E1(Exception):
    pass


class E2(Exception):
    def __init__(self, code: int) -> None:
        Exception.__init__(self, "code " + str(code), code)
        self.code = code


try:
    raise E1("a", 1, 2.5)
except E1 as e:
    print(e.args, e.args[0], len(e.args))
try:
    raise E2(7)
except E2 as e:
    print(e.args, e, e.code)
try:
    raise ValueError("v", 3)
except ValueError as e:
    print(e.args, str(e))
d: dict[str, int] = {}
try:
    print(d["k"])
except KeyError as e:
    print(e.args)

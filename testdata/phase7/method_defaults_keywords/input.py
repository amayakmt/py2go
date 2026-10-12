class Logger:
    def __init__(self, prefix: str = "log", level: int = 1) -> None:
        self.prefix = prefix
        self.level = level
        self.lines: list[str] = []

    def write(self, msg: str, *, level: int = 0, tag: str = "") -> str:
        line: str = self.prefix + ":" + str(level + self.level) + ":" + tag + ":" + msg
        self.lines.append(line)
        return line

    def many(self, *msgs: str, sep: str = ",") -> str:
        return sep.join(msgs)

    def info(self, **fields: int) -> int:
        return sum(fields.values())


a: Logger = Logger()
b: Logger = Logger("b", level=5)
c: Logger = Logger(level=2, prefix="c")
print(a.write("x"), b.write("y", level=3), c.write("z", tag="t", level=1))
print(a.many("p", "q", "r"), a.many("p", sep="-"), a.many())
print(a.info(), a.info(x=1, y=2))
print(len(a.lines), b.lines)

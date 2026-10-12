class Instr:
    def run(self, stack: list[int]) -> None:
        pass

    def name(self) -> str:
        return "nop"


class Push(Instr):
    def __init__(self, v: int) -> None:
        self.v = v

    def run(self, stack: list[int]) -> None:
        stack.append(self.v)

    def name(self) -> str:
        return f"push {self.v}"


class BinOp(Instr):
    def __init__(self, op: str) -> None:
        self.op = op

    def run(self, stack: list[int]) -> None:
        b: int = stack.pop()
        a: int = stack.pop()
        if self.op == "+":
            stack.append(a + b)
        elif self.op == "-":
            stack.append(a - b)
        elif self.op == "*":
            stack.append(a * b)
        else:
            raise ValueError("bad op " + self.op)

    def name(self) -> str:
        return "op " + self.op


class Dup(Instr):
    def run(self, stack: list[int]) -> None:
        stack.append(stack[-1])

    def name(self) -> str:
        return "dup"


class Machine:
    def __init__(self, prog: list[Instr]) -> None:
        self.prog = prog
        self.stack: list[int] = []
        self.trace: list[str] = []

    def run(self) -> int:
        for ins in self.prog:
            self.trace.append(ins.name())
            ins.run(self.stack)
        return self.stack[-1]


m: Machine = Machine([Push(3), Push(4), BinOp("+"), Dup(), BinOp("*"), Push(1), BinOp("-")])
print(m.run(), m.trace)
bad: Machine = Machine([Push(1), Push(2), BinOp("/")])
try:
    bad.run()
except ValueError as e:
    print("error:", e, bad.stack)
try:
    Machine([BinOp("+")]).run()
except IndexError as e:
    print("IndexError:", e)

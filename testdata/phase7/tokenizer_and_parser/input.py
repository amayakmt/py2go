class Token:
    def __init__(self, kind: str, text: str) -> None:
        self.kind = kind
        self.text = text

    def __repr__(self) -> str:
        return f"{self.kind}({self.text})"


class Lexer:
    def __init__(self, src: str) -> None:
        self.src = src
        self.pos = 0

    def tokens(self) -> list[Token]:
        out: list[Token] = []
        while self.pos < len(self.src):
            ch: str = self.src[self.pos]
            if ch == " ":
                self.pos += 1
            elif ch in "0123456789":
                start: int = self.pos
                while self.pos < len(self.src) and self.src[self.pos] in "0123456789":
                    self.pos += 1
                out.append(Token("num", self.src[start:self.pos]))
            elif ch in "+-*/()":
                out.append(Token("op", ch))
                self.pos += 1
            else:
                raise ValueError("bad character " + repr(ch))
        return out


class Parser:
    def __init__(self, toks: list[Token]) -> None:
        self.toks = toks
        self.i = 0

    def expr(self) -> int:
        v: int = self.term()
        while self.i < len(self.toks) and self.toks[self.i].text in ["+", "-"]:
            op: str = self.toks[self.i].text
            self.i += 1
            r: int = self.term()
            if op == "+":
                v = v + r
            else:
                v = v - r
        return v

    def term(self) -> int:
        v: int = self.atom()
        while self.i < len(self.toks) and self.toks[self.i].text in ["*", "/"]:
            op: str = self.toks[self.i].text
            self.i += 1
            r: int = self.atom()
            if op == "*":
                v = v * r
            else:
                v = v // r
        return v

    def atom(self) -> int:
        t: Token = self.toks[self.i]
        self.i += 1
        if t.kind == "num":
            return int(t.text)
        if t.text == "(":
            v: int = self.expr()
            self.i += 1
            return v
        raise ValueError("unexpected " + t.text)


for src in ["1 + 2 * 3", "(1 + 2) * 3", "10 / 3 - 1"]:
    print(src, "=", Parser(Lexer(src).tokens()).expr())
print(Lexer("12+3").tokens())
try:
    Lexer("1 $ 2").tokens()
except ValueError as e:
    print("ValueError:", e)

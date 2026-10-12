def tokenize(src):
    tokens = []
    i = 0
    while i < len(src):
        c = src[i]
        if c.isspace():
            i += 1
        elif c.isdigit() or c == ".":
            j = i
            while j < len(src) and (src[j].isdigit() or src[j] == "."):
                j += 1
            tokens.append(float(src[i:j]) if "." in src[i:j] else int(src[i:j]))
            i = j
        elif c in "+-*/()^":
            tokens.append(c)
            i += 1
        else:
            raise SyntaxError("bad character " + repr(c))
    return tokens

class Parser:
    def __init__(self, tokens):
        self.tokens = tokens
        self.pos = 0

    def peek(self):
        return self.tokens[self.pos] if self.pos < len(self.tokens) else None

    def take(self):
        tok = self.peek()
        self.pos += 1
        return tok

    def expr(self):
        value = self.term()
        while self.peek() in ("+", "-"):
            op = self.take()
            rhs = self.term()
            value = value + rhs if op == "+" else value - rhs
        return value

    def term(self):
        value = self.power()
        while self.peek() in ("*", "/"):
            op = self.take()
            rhs = self.power()
            value = value * rhs if op == "*" else value / rhs
        return value

    def power(self):
        base = self.atom()
        if self.peek() == "^":
            self.take()
            return base ** self.power()
        return base

    def atom(self):
        tok = self.take()
        if tok == "(":
            value = self.expr()
            if self.take() != ")":
                raise SyntaxError("expected )")
            return value
        if tok == "-":
            return -self.atom()
        if isinstance(tok, (int, float)):
            return tok
        raise SyntaxError("unexpected " + repr(tok))

def calc(src):
    try:
        return Parser(tokenize(src)).expr()
    except (SyntaxError, ZeroDivisionError, IndexError) as e:
        return type(e).__name__ + ": " + str(e)

for src in ("1 + 2 * 3", "(1 + 2) * 3", "2 ^ 3 ^ 2", "10 / 4", "10 / 0", "-(2 + 3) * 2", "1.5 * 4", "2 +", "3 $ 4", "((2))"):
    print(src, "=", calc(src))

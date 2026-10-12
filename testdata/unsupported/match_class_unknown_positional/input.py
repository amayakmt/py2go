class P:
    def __init__(self, a: int) -> None:
        self.a = a


p: P = P(1)
match p:
    case P(1):
        pass

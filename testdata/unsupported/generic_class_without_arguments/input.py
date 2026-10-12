class Box[T]:
    def __init__(self, v: T) -> None:
        self.v = v


b: Box = Box(1)

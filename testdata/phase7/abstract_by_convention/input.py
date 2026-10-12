class Shape:
    def area(self) -> float:
        raise NotImplementedError("subclass must implement")

    def describe(self) -> str:
        return type(self).__name__ + " area=" + str(self.area())


class Rect(Shape):
    def __init__(self, w: float, h: float) -> None:
        self.w = w
        self.h = h

    def area(self) -> float:
        return self.w * self.h


class Blob(Shape):
    pass


for s in [Rect(2.0, 3.5), Blob()]:
    try:
        print(s.describe())
    except NotImplementedError as e:
        print("NotImplementedError:", e)

class Tracker:
    def __init__(self, name: str) -> None:
        self.name = name

    def __enter__(self) -> "Tracker":
        print("enter", self.name)
        return self

    def __exit__(self, exc_type: object, exc: object, tb: object) -> None:
        print("exit", self.name, exc_type, exc)

    def hello(self) -> str:
        return "hello from " + self.name


with Tracker("a") as t:
    print(t.hello())
with Tracker("b"):
    print("body")
with Tracker("c") as c1, Tracker("d") as d1:
    print(c1.name, d1.name)
print("after")

class Timer:
    count: int = 0

    def __init__(self, label: str) -> None:
        self.label = label
        self.elapsed = 0

    def __enter__(self) -> tuple[str, int]:
        Timer.count += 1
        return (self.label, Timer.count)

    def __exit__(self, t: object, e: object, tb: object) -> None:
        self.elapsed = Timer.count * 10
        print("done", self.label, self.elapsed)


t: Timer = Timer("first")
with t as (label, n):
    print(label, n)
with Timer("a") as ta, Timer("b") as tb:
    print(ta, tb)
print(Timer.count)


def work(flag: bool) -> str:
    with Timer("inner") as (lab, k):
        if flag:
            return lab + str(k)
    return "fell through"


print(work(True), work(False))

from typing import Callable


class Stream[T]:
    def __init__(self, items: list[T]) -> None:
        self.items = items

    def filter(self, pred: Callable[[T], bool]) -> "Stream[T]":
        return Stream([x for x in self.items if pred(x)])

    def to_list(self) -> list[T]:
        return self.items

    def __len__(self) -> int:
        return len(self.items)


def map_stream[T, R](s: Stream[T], f: Callable[[T], R]) -> Stream[R]:
    return Stream([f(x) for x in s.to_list()])


nums: Stream[int] = Stream([1, 2, 3, 4])
evens: Stream[int] = nums.filter(lambda x: x % 2 == 0)
words: Stream[str] = map_stream(evens, lambda n: "n" + str(n))
print(evens.to_list(), words.to_list(), len(words))
lens: Stream[int] = map_stream(words, len)
print(lens.to_list())

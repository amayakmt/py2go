import dataclasses
import typing
import typing as t


@dataclasses.dataclass
class P:
    x: int
    y: int = 0
    tags: typing.List[str] = dataclasses.field(default_factory=list)


def f(a: typing.Optional[int], b: t.Dict[str, int]) -> typing.List[int]:
    first = 0 if a is None else a
    return [first] + list(b.values())


def g(cb: typing.Callable[[int], int], items: t.Iterable[int]) -> typing.Tuple[int, ...]:
    return tuple(cb(i) for i in items)


p = P(1)
print(p, f(None, {"k": 2}), f(5, {}))
print(g(lambda n: n * n, [1, 2, 3]))

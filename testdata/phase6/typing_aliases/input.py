from typing import List, Dict, Set, Tuple, Callable


def f(xs: List[int], d: Dict[str, int], s: Set[int], t: Tuple[int, str]) -> List[str]:
    return [str(len(xs)), str(len(d)), str(len(s)), t[1]]


print(f([1], {"a": 1}, {1, 2}, (1, "z")))
cb: Callable[[List[int]], int] = lambda xs: len(xs)
print(cb([1, 2, 3]))
m: Dict[str, List[int]] = {"a": [1]}
m["b"] = [2, 3]
print(m)

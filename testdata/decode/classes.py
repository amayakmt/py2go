from dataclasses import dataclass


@dataclass
class P(object):
    x: int = 0

    def get(self) -> int:
        return self.x


class Q(P, metaclass=type):
    pass

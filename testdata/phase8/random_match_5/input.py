from typing import Optional

log: list[str] = []


class M:
    def __init__(self, name: str, swallow: bool) -> None:
        self.name = name
        self.swallow = swallow

    def __enter__(self) -> 'M':
        log.append('enter ' + self.name)
        return self

    def __exit__(self, t: object, e: object, tb: object) -> bool:
        log.append('exit ' + self.name + ' ' + str(t))
        return self.swallow


def run(mode: int) -> int:
    for i in range(3):
        try:
            with M('m0', True) as c0, M('m1', False) as c1:
                log.append('body ' + str(i) + ' ' + str(mode))
                if mode == 0 and i == 1:
                    continue
                if mode == 1 and i == 1:
                    break
                if mode == 2 and i == 2:
                    return i * 10
                if mode == 3:
                    raise ValueError('boom ' + str(i))
        except ValueError as e:
            log.append('caught ' + str(e))
            return -1
    return 99


print(run(0), log)
log.clear()
print(run(1), log)
log.clear()
print(run(2), log)
log.clear()
print(run(3), log)
log.clear()

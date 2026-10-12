class Stop(Exception):
    pass


log: list[str] = []


def work(n: int) -> int:
    try:
        if n == 0:
            raise Stop("zero")
        log.append("ok " + str(n))
        return n
    finally:
        log.append("done " + str(n))


try:
    work(1)
    work(0)
except Stop as e:
    log.append("caught " + str(e))
    try:
        raise
    except Stop:
        log.append("again")
print(log)

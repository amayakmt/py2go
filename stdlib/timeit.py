"""timeit: measure the execution time of small code snippets (callables only)."""
import time as _time
import gc

__all__ = ["Timer", "timeit", "repeat", "default_timer"]

default_timer = _time.perf_counter


class Timer:
    def __init__(self, stmt="pass", setup="pass", timer=default_timer, globals=None):
        self.timer = timer
        if isinstance(stmt, str) or isinstance(setup, str):
            if (isinstance(stmt, str) and stmt != "pass") or (isinstance(setup, str) and setup != "pass"):
                raise NotImplementedError("timeit needs callables: a code string cannot be run")
        if not callable(stmt):
            stmt = _noop
        if not callable(setup):
            setup = _noop
        self._stmt = stmt
        self._setup = setup

    def print_exc(self, file=None):
        pass

    def timeit(self, number=1000000):
        self._setup()
        stmt = self._stmt
        timer = self.timer
        t0 = timer()
        for i in range(number):
            stmt()
        t1 = timer()
        return t1 - t0

    def repeat(self, repeat=5, number=1000000):
        r = []
        for i in range(repeat):
            t = self.timeit(number)
            r.append(t)
        return r

    def autorange(self, callback=None):
        i = 1
        while True:
            for j in 1, 2, 5:
                number = i * j
                time_taken = self.timeit(number)
                if callback:
                    callback(number, time_taken)
                if time_taken >= 0.2:
                    return (number, time_taken)
            i *= 10


def _noop():
    pass


def timeit(stmt="pass", setup="pass", timer=default_timer, number=1000000, globals=None):
    return Timer(stmt, setup, timer, globals).timeit(number)


def repeat(stmt="pass", setup="pass", timer=default_timer, repeat=5, number=1000000, globals=None):
    return Timer(stmt, setup, timer, globals).repeat(repeat, number)

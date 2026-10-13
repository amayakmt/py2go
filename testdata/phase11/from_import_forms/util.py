__all__ = ["add", "Counter", "VERSION"]
VERSION = "1.2"
_private = 5
hidden = "not exported"

def add(a, b):
    return a + b

def sub(a, b):
    return a - b

class Counter:
    def __init__(self):
        self.n = 0
    def inc(self):
        self.n += 1
        return self.n

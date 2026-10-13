from registry import register

class Square:
    def __init__(self, s):
        self.s = s
    def area(self):
        return self.s ** 2

class Rect:
    def __init__(self, w, h):
        self.w, self.h = w, h
    def area(self):
        return self.w * self.h

register("sq", lambda s: Square(s))
register("rect", lambda w, h: Rect(w, h))

print("load circle")
import math_like
class Circle:
    def __init__(self, r):
        self.r = r
    def area(self):
        return math_like.PI * self.r ** 2

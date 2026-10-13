PI = 3.14159
counter = 0

def area(r):
    return PI * r * r

def bump():
    global counter
    counter += 1
    return counter

class Shape:
    def __init__(self, name):
        self.name = name
    def describe(self):
        return f"shape {self.name} in {__name__}"

class Animal:
    sound = "..."
    def __init__(self, name):
        self.name = name
    def speak(self):
        return f"{self.name} says {self.sound}"
    def __repr__(self):
        return f"{type(self).__name__}({self.name!r})"

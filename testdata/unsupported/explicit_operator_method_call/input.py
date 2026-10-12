class V:
    def __add__(self, o: "V") -> "V":
        return o


a: V = V()
b: V = a.__add__(a)

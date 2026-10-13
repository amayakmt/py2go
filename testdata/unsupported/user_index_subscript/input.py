class I:
    def __index__(self):
        return 1
print([10, 20][I()])

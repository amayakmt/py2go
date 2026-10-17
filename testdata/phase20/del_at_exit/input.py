class Resource:
    def __init__(self, name):
        self.name = name
        print("open", name)

    def __del__(self):
        print("close", self.name)


r = Resource("a")
print("working")

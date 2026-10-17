class Plugin:
    registry = {}

    def __init_subclass__(cls, name=None, **kwargs):
        super().__init_subclass__(**kwargs)
        Plugin.registry[name or cls.__name__.lower()] = cls

    def run(self):
        return "base"


class Alpha(Plugin):
    def run(self):
        return "alpha"


class Beta(Plugin, name="b"):
    def run(self):
        return "beta"


class Gamma(Beta):
    pass


print(sorted(Plugin.registry))
print([Plugin.registry[k]().run() for k in sorted(Plugin.registry)])


class Counted:
    count = 0

    def __init_subclass__(cls):
        Counted.count += 1


class X(Counted):
    pass


class Y(Counted):
    pass


print(Counted.count)

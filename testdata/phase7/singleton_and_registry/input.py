class Registry:
    instance_count: int = 0
    entries: dict[str, int] = {}

    def __init__(self) -> None:
        Registry.instance_count += 1

    @classmethod
    def register(cls, name: str, value: int) -> None:
        cls.entries[name] = value

    @staticmethod
    def lookup(name: str) -> int:
        return Registry.entries[name]


Registry.register("a", 1)
Registry.register("b", 2)
r1: Registry = Registry()
r2: Registry = Registry()
print(Registry.lookup("a") + Registry.lookup("b"), Registry.instance_count)
try:
    Registry.lookup("zzz")
except KeyError as e:
    print("KeyError:", e)

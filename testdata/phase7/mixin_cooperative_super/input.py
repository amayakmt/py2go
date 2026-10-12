class Entity:
    def __init__(self) -> None:
        self.log: list[str] = []
        self.log.append("Entity")


class Timestamped(Entity):
    def __init__(self) -> None:
        super().__init__()
        self.log.append("Timestamped")


class Named(Entity):
    def __init__(self) -> None:
        super().__init__()
        self.log.append("Named")


class User(Timestamped, Named):
    def __init__(self) -> None:
        super().__init__()
        self.log.append("User")


u: User = User()
print(u.log)
print(isinstance(u, Named), isinstance(u, Timestamped), isinstance(u, Entity))

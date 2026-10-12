class Task:
    def __init__(self, name: str, prio: int, cost: float) -> None:
        self.name = name
        self.prio = prio
        self.cost = cost

    def __lt__(self, other: "Task") -> bool:
        return (self.prio, self.name) < (other.prio, other.name)

    def __repr__(self) -> str:
        return self.name


tasks: list[Task] = [Task("c", 2, 1.5), Task("a", 2, 0.5), Task("b", 1, 9.0), Task("d", 3, 2.0)]
print(sorted(tasks))
print(sorted(tasks, key=lambda t: t.cost), sorted(tasks, key=lambda t: -t.prio))
print(sorted(tasks, reverse=True), min(tasks), max(tasks))
tasks.sort()
print(tasks)
tasks.sort(key=lambda t: t.name, reverse=True)
print(tasks)

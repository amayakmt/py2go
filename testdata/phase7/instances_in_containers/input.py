class Employee:
    def __init__(self, name: str, dept: str, salary: int) -> None:
        self.name = name
        self.dept = dept
        self.salary = salary

    def __repr__(self) -> str:
        return f"{self.name}({self.salary})"


staff: list[Employee] = [
    Employee("ann", "eng", 120),
    Employee("bob", "ops", 90),
    Employee("cy", "eng", 100),
    Employee("di", "ops", 95),
]
by_dept: dict[str, list[Employee]] = {}
for e in staff:
    if e.dept not in by_dept:
        by_dept[e.dept] = []
    by_dept[e.dept].append(e)
print(by_dept)
print(sorted(staff, key=lambda e: e.salary))
print(max(staff, key=lambda e: e.salary), min(staff, key=lambda e: e.salary))
totals: dict[str, int] = {d: sum(e.salary for e in es) for d, es in by_dept.items()}
print(totals)
names: list[str] = [e.name.upper() for e in staff if e.salary > 95]
print(names)
pairs: list[tuple[str, Employee]] = [(e.name, e) for e in staff[:2]]
print(pairs)

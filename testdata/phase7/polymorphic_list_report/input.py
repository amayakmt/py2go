class Employee:
    def __init__(self, name: str, base: int) -> None:
        self.name = name
        self.base = base

    def pay(self) -> int:
        return self.base

    def title(self) -> str:
        return "Employee"

    def line(self) -> str:
        return f"{self.title():<10}{self.name:>6}{self.pay():>8}"


class Manager(Employee):
    def __init__(self, name: str, base: int, reports: int) -> None:
        super().__init__(name, base)
        self.reports = reports

    def pay(self) -> int:
        return super().pay() + 100 * self.reports

    def title(self) -> str:
        return "Manager"


class Intern(Employee):
    def pay(self) -> int:
        return self.base // 2

    def title(self) -> str:
        return "Intern"


staff: list[Employee] = [Employee("ann", 500), Manager("bob", 800, 3), Intern("cy", 400)]
for e in staff:
    print(e.line())
print(sum(e.pay() for e in staff))
best: Employee = max(staff, key=lambda e: e.pay())
print(best.name, [e.title() for e in sorted(staff, key=lambda e: -e.pay())])

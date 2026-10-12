class Student:
    def __init__(self, name: str, grades: list[int]) -> None:
        self.name = name
        self.grades = grades

    @property
    def average(self) -> float:
        if not self.grades:
            return 0.0
        return sum(self.grades) / len(self.grades)

    @property
    def letter(self) -> str:
        avg: float = self.average
        if avg >= 90:
            return "A"
        if avg >= 80:
            return "B"
        if avg >= 70:
            return "C"
        return "F"

    def __repr__(self) -> str:
        return f"{self.name}:{self.average:.1f}{self.letter}"


class Course:
    def __init__(self, title: str) -> None:
        self.title = title
        self.students: list[Student] = []

    def enroll(self, s: Student) -> "Course":
        self.students.append(s)
        return self

    def ranking(self) -> list[Student]:
        return sorted(self.students, key=lambda s: s.average, reverse=True)

    def histogram(self) -> dict[str, int]:
        h: dict[str, int] = {}
        for s in self.students:
            h[s.letter] = h.get(s.letter, 0) + 1
        return h


c: Course = Course("py").enroll(Student("a", [95, 88])).enroll(Student("b", [72, 65, 80])).enroll(Student("c", [])).enroll(Student("d", [85, 85, 85]))
print(c.ranking())
print(c.histogram())
print([s.name for s in c.students if s.letter in ["A", "B"]])

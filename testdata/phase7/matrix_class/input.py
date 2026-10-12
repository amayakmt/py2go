class Matrix:
    def __init__(self, rows: list[list[int]]) -> None:
        self.rows = rows

    def __add__(self, o: "Matrix") -> "Matrix":
        return Matrix([[a + b for a, b in zip(r1, r2)] for r1, r2 in zip(self.rows, o.rows)])

    def __mul__(self, o: "Matrix") -> "Matrix":
        n: int = len(self.rows)
        m: int = len(o.rows[0])
        k: int = len(o.rows)
        out: list[list[int]] = []
        for i in range(n):
            row: list[int] = []
            for j in range(m):
                s: int = 0
                for t in range(k):
                    s += self.rows[i][t] * o.rows[t][j]
                row.append(s)
            out.append(row)
        return Matrix(out)

    def transpose(self) -> "Matrix":
        return Matrix([[self.rows[i][j] for i in range(len(self.rows))] for j in range(len(self.rows[0]))])

    def __eq__(self, o: object) -> bool:
        return isinstance(o, Matrix) and self.rows == o.rows

    def __getitem__(self, ij: tuple[int, int]) -> int:
        return self.rows[ij[0]][ij[1]]

    def __repr__(self) -> str:
        return "Matrix(" + str(self.rows) + ")"


a: Matrix = Matrix([[1, 2], [3, 4]])
b: Matrix = Matrix([[0, 1], [1, 0]])
print(a + b, a * b, a.transpose())
print(a == Matrix([[1, 2], [3, 4]]), a == b, a[1, 0], a[0, 1])

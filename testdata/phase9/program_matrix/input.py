def zeros(r, c):
    return [[0] * c for _ in range(r)]

def identity(n):
    m = zeros(n, n)
    for i in range(n):
        m[i][i] = 1
    return m

def matmul(a, b):
    out = zeros(len(a), len(b[0]))
    for i in range(len(a)):
        for j in range(len(b[0])):
            out[i][j] = sum(a[i][k] * b[k][j] for k in range(len(b)))
    return out

def transpose(m):
    return [list(r) for r in zip(*m)]

def power(m, n):
    result = identity(len(m))
    while n:
        if n & 1:
            result = matmul(result, m)
        m = matmul(m, m)
        n >>= 1
    return result

def det(m):
    if len(m) == 1:
        return m[0][0]
    return sum((-1) ** j * m[0][j] * det([row[:j] + row[j + 1:] for row in m[1:]]) for j in range(len(m)))

a = [[1, 2], [3, 4]]
b = [[0, 1], [1, 0]]
print(matmul(a, b), transpose(a), transpose([[1, 2, 3]]))
print(power([[1, 1], [1, 0]], 10), power(a, 0))
print(det(a), det([[2, 0, 1], [1, 3, 2], [1, 1, 1]]), det(identity(4)))
print(matmul([[0.5, 0.5]], [[2], [4]]), [[x * 1.5 for x in row] for row in a])

def det(m):
    n = len(m)
    m = [row[:] for row in m]
    sign, prev = 1, 1
    for k in range(n - 1):
        if m[k][k] == 0:
            for i in range(k + 1, n):
                if m[i][k] != 0:
                    m[k], m[i] = m[i], m[k]
                    sign = -sign
                    break
            else:
                return 0
        for i in range(k + 1, n):
            for j in range(k + 1, n):
                m[i][j] = (m[i][j] * m[k][k] - m[i][k] * m[k][j]) // prev
        prev = m[k][k]
    return sign * m[n - 1][n - 1]

def matmul(a, b):
    return [[sum(x * y for x, y in zip(r, c)) for c in zip(*b)] for r in a]

def matpow(m, e):
    r = [[int(i == j) for j in range(len(m))] for i in range(len(m))]
    while e:
        if e & 1:
            r = matmul(r, m)
        m = matmul(m, m)
        e >>= 1
    return r

print(det([[2, 0], [0, 3]]), det([[1, 2], [3, 4]]), det([[0, 1], [1, 0]]), det([[1, 2, 3], [4, 5, 6], [7, 8, 9]]), det([[2, -1, 0], [-1, 2, -1], [0, -1, 2]]))
h = [[1000003 ** (i + j) % (10 ** 12 + 39) + 1 for j in range(5)] for i in range(5)]
print(det(h) % 10 ** 9, det([[10 ** 20 + i * j for j in range(4)] for i in range(4)]))
fibm = matpow([[1, 1], [1, 0]], 500)
print(fibm[0][1], len(str(fibm[0][0])), matpow([[2, 1], [1, 1]], 64)[0])
vander = [[(i + 1) ** j for j in range(6)] for i in range(6)]
print(det(vander), det([[2 ** (i * j) for j in range(5)] for i in range(5)]) != 0)

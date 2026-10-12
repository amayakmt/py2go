xs: list[int] = [1, 2, 3]
for x in xs:
    if x < 3:
        xs.append(x + 10)
print(xs)
ys: list[int] = [1, 2, 3, 4]
for i in range(len(ys)):
    ys[i] *= 2
print(ys)
zs: list[int] = [5, 6, 7, 8]
for z in zs:
    zs.remove(z)
print(zs)
ws: list[int] = [0]
count: int = 0
for w in ws:
    count += 1
    if count < 5:
        ws.append(w)
print(ws, count)

a = b = c = []
a.append(1)
print(a, b, c, a is c)
x = y = 5
x += 1
print(x, y)
p, q = r = (1, 2)
print(p, q, r)
i = j = 0
for i in range(3):
    j += i
print(i, j)
d = {}
d["k"] = e = [1]
e.append(2)
print(d, e)

lst = [3, 1, 2]
lst            # an expression statement
lst[0]
len(lst)
lst.sort() or 0
-lst[0]
"string"
42
print(lst)
try:
    lst[10]
except IndexError as exc:
    print("IndexError", exc)
try:
    undefined_name_zzz
except NameError as exc:
    print("NameError", exc)

x = 1
print(x, type(x).__name__)
x = "now a string"
print(x, type(x).__name__)
x = [1, 2, 3]
x.append(4)
print(x, type(x).__name__)
x = None
print(x, type(x).__name__)
x = 2.5
print(x * 2, type(x).__name__)

def changing(n):
    v = n
    for i in range(3):
        if i == 1:
            v = str(v) + "!"
        else:
            v = [v]
    return v

print(changing(1), changing("s"))

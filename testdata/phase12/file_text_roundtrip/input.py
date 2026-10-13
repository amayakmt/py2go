import os
path = "/tmp/py2go_phase12_text.txt"
with open(path, "w") as f:
    print(f.write("alpha\n"), f.write("beta\ngamma\n"), file=None)
    f.writelines(["delta\n", "épsilon\n"])
    print("printed line", file=f)
with open(path) as f:
    text = f.read()
print(repr(text), len(text))
with open(path) as f:
    print([line.rstrip("\n") for line in f])
with open(path) as f:
    print(f.readline(), f.readline(5), f.readline(), f.readlines()[:2], f.readline() == "")
with open(path, "a") as f:
    f.write("appended")
with open(path) as f:
    lines = f.readlines()
print(len(lines), lines[-1], sum(len(l) for l in lines))
with open(path) as f:
    for n, line in enumerate(f, 1):
        if n == 2:
            print(n, repr(line))
f = open(path)
print(f.name == path, f.mode, f.closed, f.encoding)
f.close()
print(f.closed)
f.close()
try:
    f.read()
except ValueError as e:
    print(e)
with open(path, "w+") as f:
    f.write("0123456789")
    f.seek(3)
    print(f.read(4), f.tell())
    f.seek(0)
    print(f.read(2))
os.remove(path)
print(os.path.exists(path))

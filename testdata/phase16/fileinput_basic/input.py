import fileinput
import os

d = "/tmp/py2go_fileinput_test"
os.makedirs(d, exist_ok=True)
a, b = os.path.join(d, "a.txt"), os.path.join(d, "b.txt")
open(a, "w").write("a1\na2\n")
open(b, "w").write("b1\nb2\nb3")
for line in fileinput.input([a, b]):
    print(fileinput.filename().replace(d, "D"), fileinput.lineno(), fileinput.filelineno(), fileinput.isfirstline(), repr(line))
fileinput.close()
with fileinput.FileInput(files=[a, b]) as fi:
    for line in fi:
        if fi.filelineno() == 2:
            fi.nextfile()
        print(fi.lineno(), repr(line))
with fileinput.input(a, openhook=fileinput.hook_encoded("utf-8")) as fi:
    print(fi.readline(), fi.readline(), repr(fi.readline()))
print(list(fileinput.input(b, mode="rb")))
try:
    fileinput.input(a)
    fileinput.input(a)
except RuntimeError as e:
    print(e)
fileinput.close()
os.remove(a)
os.remove(b)
os.rmdir(d)

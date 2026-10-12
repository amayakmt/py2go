parts: list[str] = []
for i in range(1, 6):
    parts.append(str(i * i))
print(",".join(parts))
line: str = ""
for i in range(3):
    line += f"[{i}]"
print(line)
csv: str = "name,age\nada,36\nbob,25"
for row in csv.split("\n")[1:]:
    cells: list[str] = row.split(",")
    print(cells, cells[0].upper(), int(cells[1]) + 1)
acc: str = ""
for ch in "hello":
    acc = ch + acc
print(acc)

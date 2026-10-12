out: list[str] = []
for i in range(4):
    for j in range(4):
        try:
            try:
                if j == 1:
                    continue
                if j == 3:
                    break
                if i == 2:
                    raise ValueError("i2")
                out.append(str(i) + str(j))
            finally:
                out.append("f")
        except ValueError:
            out.append("v")
            break
    out.append("|")
print(" ".join(out))

def find_pair(limit: int, out: list[str]) -> str:
    for a in range(limit):
        for b in range(limit):
            try:
                try:
                    if a * b == 6:
                        return str(a) + "x" + str(b)
                finally:
                    out.append("i")
            finally:
                out.append("o")
    return "none"

print(find_pair(5, out), len(out))

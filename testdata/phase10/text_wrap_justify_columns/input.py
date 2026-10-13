def wrap(text, width):
    lines, cur = [], []
    for w in text.split():
        if sum(map(len, cur)) + len(cur) + len(w) > width and cur:
            lines.append(" ".join(cur))
            cur = []
        cur.append(w)
    if cur:
        lines.append(" ".join(cur))
    return lines

def justify(line, width):
    words = line.split()
    if len(words) == 1:
        return line.ljust(width)
    gaps = len(words) - 1
    total = width - sum(map(len, words))
    q, r = divmod(total, gaps)
    return "".join(w + " " * (q + (1 if i < r else 0)) for i, w in enumerate(words[:-1])) + words[-1]

text = "In 1066 the Normans crossed the Channel, and ünïcode wörds like Gëmütlichkeit and 日本語 appear in the middle of English prose, which is hard to wrap."
for width in (20, 33, 50):
    ls = wrap(text, width)
    print(width, len(ls))
    for l in ls:
        print("|" + justify(l, width) + "|", len(l))
    print("\n".join(l.center(width, ".") for l in ls[:2]))
rows = [("name", "qty", "price"), ("widget", 4, 3.5), ("gadget", 12, 10.25), ("thingamajig", 100, 0.5)]
w = [max(len(str(r[i])) for r in rows) for i in range(3)]
for r in rows:
    print(f"{str(r[0]):<{w[0]}} | {str(r[1]):>{w[1]}} | {str(r[2]):>{w[2]}}")
print("-" * (sum(w) + 6), "%-*s|%*s" % (8, "ab", 6, "cd"), "{:-^30}".format(" total "), "{:>{}}|{:<{}}".format(1, 5, 2, 5))

WIDTH = 12
SEP = " | "
def row(*cells):
    return SEP.join(str(c).ljust(WIDTH) for c in cells)
def money(x):
    return f"${x:,.2f}"
def banner(s):
    return s.center(WIDTH * 3, "=")

def run(n):
    if n > 2:
        raise RuntimeError(f"too big: {n}")
    return n

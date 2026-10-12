def make_counter(start=0):
    count = start
    def inc(step=1):
        nonlocal count
        count += step
        return count
    def get():
        return count
    return inc, get

inc, get = make_counter(10)
print(inc(), inc(5), get())
inc2, get2 = make_counter()
print(inc2(), get(), get2())

def make_multipliers():
    return [lambda x, n=n: x * n for n in range(4)]

print([f(10) for f in make_multipliers()])
late = [lambda: n for n in range(3)]
print([f() for f in late])

total = 0
def add_to_total(n):
    global total
    total += n
    return total

print(add_to_total(5), add_to_total(10), total)
x = "global"
def shadow():
    x = "local"
    def inner():
        return x
    return inner()
print(shadow(), x)

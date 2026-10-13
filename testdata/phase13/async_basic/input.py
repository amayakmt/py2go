import asyncio


async def add(a, b):
    return a + b


async def fib(n):
    if n < 2:
        return n
    return await fib(n - 1) + await fib(n - 2)


async def greet(name):
    await asyncio.sleep(0)
    return "hello " + name


class Fut:
    def __init__(self):
        self.done = False
        self.result = None

    def __await__(self):
        if not self.done:
            yield self
        return self.result


def drive(coro):
    try:
        x = coro.send(None)
        while True:
            if isinstance(x, Fut):
                x.done = True
                x.result = 42
            x = coro.send(None)
    except StopIteration as e:
        return e.value


async def with_fut():
    v = await Fut()
    return v + 1


print(drive(with_fut()))
c = add(1, 2)
print(type(c).__name__, asyncio.iscoroutine(c), asyncio.iscoroutinefunction(add))
print(drive(c))


async def main():
    print(await add(2, 3))
    print(await fib(15))
    print(await greet("bob"))
    results = []
    for i in range(3):
        results.append(await add(i, i))
    print(results)
    print([await add(i, 1) for i in range(4)])
    return "done"


print(asyncio.run(main()))
coro = add(5, 6)
try:
    coro.send(1)
except TypeError as e:
    print("TypeError", e)
coro2 = add(1, 1)
coro2.close()
try:
    coro2.send(None)
except RuntimeError as e:
    print("RuntimeError", e)
try:
    asyncio.run(add)
except ValueError as e:
    print("ValueError", str(e)[:30])

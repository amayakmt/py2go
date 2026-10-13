import asyncio


async def countdown(n):
    while n > 0:
        yield n
        await asyncio.sleep(0)
        n -= 1


async def pairs():
    for i in range(3):
        try:
            yield i
        finally:
            print("finally", i)


class Resource:
    def __init__(self, name):
        self.name = name

    async def __aenter__(self):
        print("enter", self.name)
        await asyncio.sleep(0)
        return self

    async def __aexit__(self, exc_type, exc, tb):
        print("exit", self.name, exc_type.__name__ if exc_type else None)
        return exc_type is KeyError


class Ticker:
    def __init__(self, n):
        self.i = 0
        self.n = n

    def __aiter__(self):
        return self

    async def __anext__(self):
        if self.i >= self.n:
            raise StopAsyncIteration
        self.i += 1
        await asyncio.sleep(0)
        return self.i * 10


async def main():
    out = []
    async for v in countdown(3):
        out.append(v)
    print(out)
    async for v in Ticker(3):
        print("tick", v)
    g1 = pairs()
    async for v in g1:
        if v == 1:
            break
    else:
        print("no break")
    await g1.aclose()
    async for v in pairs():
        pass
    else:
        print("exhausted")
    async with Resource("r1") as r:
        print("body", r.name)
    async with Resource("r2"), Resource("r3"):
        print("both")
    async with Resource("r4"):
        raise KeyError("swallowed")
    try:
        async with Resource("r5"):
            raise ValueError("propagates")
    except ValueError as e:
        print("caught", e)
    g = countdown(2)
    print(await g.__anext__(), await g.__anext__())
    try:
        await g.__anext__()
    except StopAsyncIteration:
        print("stop")
    g2 = countdown(5)
    print(await g2.__anext__())
    await g2.aclose()
    try:
        await g2.__anext__()
    except StopAsyncIteration:
        print("closed")
    print(type(g2).__name__)
    print([x async for x in countdown(3)])
    print([x * 2 async for x in countdown(4) if x % 2 == 0])


asyncio.run(main())

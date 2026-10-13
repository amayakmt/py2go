import asyncio

log = []


async def worker(name, delay):
    log.append("start " + name)
    await asyncio.sleep(delay)
    log.append("end " + name)
    return name.upper()


async def failing(delay):
    await asyncio.sleep(delay)
    raise ValueError("boom " + str(delay))


async def main():
    r = await asyncio.gather(worker("a", 0.03), worker("b", 0.01), worker("c", 0.02))
    print(r)
    print(log)
    r = await asyncio.gather(worker("d", 0.01), failing(0.02), return_exceptions=True)
    print(r[0], type(r[1]).__name__, r[1])
    try:
        await asyncio.gather(worker("e", 0.05), failing(0.01))
    except ValueError as e:
        print("caught", e)
    print(await asyncio.gather())
    t1 = asyncio.create_task(worker("f", 0.01))
    t2 = asyncio.create_task(worker("g", 0.01))
    print(await asyncio.gather(t1, t2, t1))
    done, pending = await asyncio.wait([asyncio.create_task(worker("h", 0.01)), asyncio.create_task(worker("i", 0.2))], return_when=asyncio.FIRST_COMPLETED)
    print(len(done), len(pending))
    for p in pending:
        p.cancel()
    await asyncio.sleep(0.01)
    res = []
    for coro in asyncio.as_completed([worker("j", 0.03), worker("k", 0.01), worker("l", 0.02)]):
        res.append(await coro)
    print(res)


asyncio.run(main())

import asyncio


async def slow(n, result="ok"):
    await asyncio.sleep(n)
    return result


async def main():
    print(await asyncio.wait_for(slow(0.01), 1))
    try:
        await asyncio.wait_for(slow(1), 0.02)
    except TimeoutError:
        print("wait_for timeout")
    try:
        async with asyncio.timeout(0.02):
            await slow(1)
    except TimeoutError as e:
        print("timeout ctx", type(e).__name__)
    async with asyncio.timeout(1) as cm:
        await slow(0.01)
    print(cm.expired())
    try:
        await asyncio.wait_for(slow(1), 0)
    except asyncio.TimeoutError:
        print("zero timeout")
    print(await asyncio.wait_for(slow(0.01, "x"), None))
    t = asyncio.create_task(slow(1))
    try:
        await asyncio.wait_for(t, 0.01)
    except TimeoutError:
        print("task cancelled:", t.cancelled())
    d, p = await asyncio.wait([asyncio.create_task(slow(0.2)), asyncio.create_task(slow(0.01))], timeout=0.05)
    print(len(d), len(p))
    for x in p:
        x.cancel()
    await asyncio.sleep(0)
    loop = asyncio.get_running_loop()
    hits = []
    loop.call_later(0.02, hits.append, "later")
    loop.call_soon(hits.append, "soon")
    h = loop.call_later(0.01, hits.append, "cancelled")
    h.cancel()
    await asyncio.sleep(0.05)
    print(hits)
    print(loop.time() > 0, loop.is_running(), loop.is_closed())


asyncio.run(main())

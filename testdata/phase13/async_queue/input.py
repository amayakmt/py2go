import asyncio


async def producer(q, n):
    for i in range(n):
        await q.put(i)
        print("put", i)
    await q.put(None)


async def consumer(q, name):
    total = 0
    while True:
        item = await q.get()
        if item is None:
            q.task_done()
            break
        total += item
        print(name, "got", item)
        q.task_done()
    return total


async def main():
    q = asyncio.Queue(maxsize=2)
    p = asyncio.create_task(producer(q, 5))
    c = asyncio.create_task(consumer(q, "c1"))
    await p
    await q.join()
    print(await c)
    print(q.empty(), q.qsize(), q.full(), q.maxsize)
    pq = asyncio.PriorityQueue()
    for x in [5, 1, 3]:
        pq.put_nowait(x)
    print([pq.get_nowait() for _ in range(3)])
    lq = asyncio.LifoQueue()
    for x in [1, 2, 3]:
        lq.put_nowait(x)
    print([lq.get_nowait() for _ in range(3)])
    try:
        lq.get_nowait()
    except asyncio.QueueEmpty:
        print("empty")
    q2 = asyncio.Queue(1)
    q2.put_nowait(1)
    try:
        q2.put_nowait(2)
    except asyncio.QueueFull:
        print("full")
    try:
        q2.task_done()
        q2.task_done()
    except ValueError as e:
        print("ValueError", e)


asyncio.run(main())

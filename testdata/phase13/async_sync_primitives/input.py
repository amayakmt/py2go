import asyncio

events = []


async def locker(lock, name, hold):
    async with lock:
        events.append(name + " in")
        await asyncio.sleep(hold)
        events.append(name + " out")


async def waiter(ev, name):
    await ev.wait()
    events.append(name + " released")


async def sem_user(sem, name, active):
    async with sem:
        active.append(name)
        events.append("active " + str(len(active)))
        await asyncio.sleep(0.01)
        active.remove(name)


async def cond_consumer(cond, box, name):
    async with cond:
        await cond.wait_for(lambda: len(box) > 0)
        events.append(name + " took " + str(box.pop()))


async def main():
    lock = asyncio.Lock()
    print(lock.locked())
    await asyncio.gather(locker(lock, "a", 0.01), locker(lock, "b", 0.01), locker(lock, "c", 0))
    print(events)
    events.clear()
    ev = asyncio.Event()
    tasks = [asyncio.create_task(waiter(ev, n)) for n in "xyz"]
    await asyncio.sleep(0.01)
    print(ev.is_set(), events)
    ev.set()
    await asyncio.gather(*tasks)
    print(ev.is_set(), events)
    ev.clear()
    print(ev.is_set())
    events.clear()
    sem = asyncio.Semaphore(2)
    active = []
    await asyncio.gather(*[sem_user(sem, "u" + str(i), active) for i in range(5)])
    print(events)
    events.clear()
    cond = asyncio.Condition()
    box = []
    cs = [asyncio.create_task(cond_consumer(cond, box, "c" + str(i))) for i in range(2)]
    await asyncio.sleep(0.01)
    async with cond:
        box.append(1)
        box.append(2)
        cond.notify_all()
    await asyncio.gather(*cs)
    print(events)
    b = asyncio.BoundedSemaphore(1)
    await b.acquire()
    b.release()
    try:
        b.release()
    except ValueError as e:
        print("ValueError", e)
    try:
        lock.release()
    except RuntimeError as e:
        print("RuntimeError", e)
    bar = asyncio.Barrier(3)
    order = []

    async def party(i):
        await asyncio.sleep(0.01 * i)
        idx = await bar.wait()
        order.append((i, idx))

    await asyncio.gather(party(0), party(1), party(2))
    print(sorted(order))


asyncio.run(main())

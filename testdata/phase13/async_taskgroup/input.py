import asyncio


async def work(name, delay, fail=False):
    await asyncio.sleep(delay)
    if fail:
        raise ValueError(name)
    print("finished", name)
    return name


async def main():
    async with asyncio.TaskGroup() as tg:
        t1 = tg.create_task(work("a", 0.01))
        t2 = tg.create_task(work("b", 0.02))
    print(t1.result(), t2.result())
    try:
        async with asyncio.TaskGroup() as tg:
            tg.create_task(work("c", 0.01, True))
            tg.create_task(work("d", 1))
    except* ValueError as eg:
        print("group", len(eg.exceptions), eg.exceptions[0])
    try:
        async with asyncio.TaskGroup() as tg:
            tg.create_task(work("e", 0.01, True))
            tg.create_task(work("f", 0.01, True))
    except* ValueError as eg:
        print(sorted(str(e) for e in eg.exceptions))
    try:
        async with asyncio.TaskGroup() as tg:
            tg.create_task(work("g", 1))
            raise KeyError("body")
    except* KeyError as eg:
        print("body error", eg.exceptions[0])


asyncio.run(main())

import asyncio


async def sleeper(name):
    try:
        print(name, "sleeping")
        await asyncio.sleep(10)
    except asyncio.CancelledError:
        print(name, "cancelled")
        raise
    finally:
        print(name, "cleanup")


async def stubborn():
    try:
        await asyncio.sleep(10)
    except asyncio.CancelledError:
        print("ignoring cancel")
    await asyncio.sleep(0.01)
    return "survived"


async def main():
    t = asyncio.create_task(sleeper("s1"), name="first")
    print(t.get_name(), t.done(), t.cancelled())
    await asyncio.sleep(0.01)
    t.cancel()
    try:
        await t
    except asyncio.CancelledError:
        print("main saw cancel", t.cancelled(), t.done())
    t2 = asyncio.create_task(stubborn())
    await asyncio.sleep(0.01)
    t2.cancel()
    print(await t2)
    t3 = asyncio.create_task(sleeper("s3"))
    await asyncio.sleep(0)
    print(t3.cancel("because"), t3.cancelling())
    try:
        await t3
    except asyncio.CancelledError as e:
        print("msg", e.args)
    print(t3.cancel())
    f = asyncio.get_running_loop().create_future()
    f.cancel()
    print(f.cancelled(), f.done())
    try:
        f.result()
    except asyncio.CancelledError:
        print("future cancelled")
    cur = asyncio.current_task()
    print(cur.get_name().startswith("Task-"), len(asyncio.all_tasks()))
    shielded = asyncio.shield(asyncio.sleep(0.02, "shielded"))
    print(await shielded)


asyncio.run(main())

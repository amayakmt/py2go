import asyncio


async def boom():
    raise RuntimeError("inside")


async def nested():
    try:
        await boom()
    except RuntimeError as e:
        print("nested caught", e)
        raise KeyError("converted") from e


async def main():
    try:
        await nested()
    except KeyError as e:
        print(type(e).__name__, e, type(e.__cause__).__name__)
    f = asyncio.get_running_loop().create_future()
    try:
        f.result()
    except asyncio.InvalidStateError as e:
        print("InvalidStateError", e)
    f.set_result(5)
    print(f.result(), f.done(), f.exception())
    try:
        f.set_result(6)
    except asyncio.InvalidStateError:
        print("already set")
    g = asyncio.get_running_loop().create_future()
    g.set_exception(ValueError("bad"))
    try:
        await g
    except ValueError as e:
        print("awaited", e)
    t = asyncio.create_task(boom())
    await asyncio.sleep(0)
    await asyncio.sleep(0)
    print(t.done(), repr(t.exception()))
    try:
        await t
    except RuntimeError:
        print("task raised")
    try:
        await 5
    except TypeError as e:
        print("TypeError", e)
    done = asyncio.get_running_loop().create_future()
    done.add_done_callback(lambda fut: print("callback", fut.result()))
    done.set_result("cb")
    await asyncio.sleep(0)
    try:
        asyncio.run(boom())
    except RuntimeError as e:
        print("nested run", e)


asyncio.run(main())
print("after")
asyncio.run(boom())

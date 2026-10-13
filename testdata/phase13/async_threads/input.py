import asyncio
import contextvars
import threading
import time

var = contextvars.ContextVar("var", default="dflt")


async def child(n):
    seen = var.get()
    var.set("child" + str(n))
    await asyncio.sleep(0)
    return seen, var.get()


def blocking(x):
    time.sleep(0.02)
    return x * 2, threading.current_thread().name.startswith("asyncio")


async def main():
    var.set("main")
    print(await asyncio.gather(child(1), child(2)), var.get())
    print(await asyncio.to_thread(blocking, 21))
    loop = asyncio.get_running_loop()
    print(await loop.run_in_executor(None, blocking, 5))
    res = []

    def other():
        time.sleep(0.01)
        loop.call_soon_threadsafe(res.append, "from thread")

    th = threading.Thread(target=other)
    th.start()
    await asyncio.sleep(0.05)
    th.join()
    print(res)
    fut = asyncio.run_coroutine_threadsafe(child(3), loop)
    await asyncio.sleep(0.01)
    print(fut.result(1))
    results = await asyncio.gather(*[asyncio.to_thread(blocking, i) for i in range(3)])
    print([r[0] for r in results])
    t = asyncio.create_task(child(4), context=contextvars.copy_context())
    print(await t)


asyncio.run(main())

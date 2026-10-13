import asyncio
import contextvars

request_id = contextvars.ContextVar("request_id", default="-")
log = []


def note(msg):
    log.append(request_id.get() + ": " + msg)


class Job:
    def __init__(self, n, cost):
        self.n = n
        self.cost = cost


class Pool:
    def __init__(self, size):
        self.size = size
        self.queue = asyncio.Queue()
        self.results = {}
        self.failed = []

    async def worker(self, wid):
        while True:
            job = await self.queue.get()
            request_id.set("job" + str(job.n))
            try:
                note("start on w" + str(wid))
                await asyncio.sleep(job.cost)
                if job.n % 4 == 3:
                    raise RuntimeError("job " + str(job.n) + " exploded")
                self.results[job.n] = job.n * 100
                note("done")
            except RuntimeError as e:
                self.failed.append(str(e))
                note("failed")
            finally:
                self.queue.task_done()

    async def run(self, jobs):
        workers = [asyncio.create_task(self.worker(i)) for i in range(self.size)]
        for j in jobs:
            self.queue.put_nowait(j)
        await self.queue.join()
        for w in workers:
            w.cancel()
        outcomes = await asyncio.gather(*workers, return_exceptions=True)
        return [type(o).__name__ for o in outcomes]


async def main():
    pool = Pool(3)
    outcomes = await pool.run([Job(i, 0.01 * ((i % 3) + 1)) for i in range(8)])
    print(outcomes)
    print(sorted(pool.results.items()))
    print(sorted(pool.failed))
    print(len(log), request_id.get())
    print(sorted(l for l in log if l.endswith("failed")))
    async with asyncio.TaskGroup() as tg:
        a = tg.create_task(asyncio.sleep(0.01, "a"))
        b = tg.create_task(asyncio.sleep(0.02, "b"))
    print(a.result() + b.result())


asyncio.run(main())

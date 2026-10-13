import asyncio
import random


class Page:
    def __init__(self, url, links, delay):
        self.url = url
        self.links = links
        self.delay = delay


SITE = {
    "/": Page("/", ["/a", "/b", "/c"], 0.01),
    "/a": Page("/a", ["/d", "/b"], 0.02),
    "/b": Page("/b", ["/d"], 0.01),
    "/c": Page("/c", [], 0.03),
    "/d": Page("/d", ["/"], 0.01),
    "/slow": Page("/slow", [], 5),
}


class FetchError(Exception):
    def __init__(self, url, reason):
        Exception.__init__(self, url + ": " + reason)
        self.url = url
        self.reason = reason


class Fetcher:
    def __init__(self, limit):
        self.sem = asyncio.Semaphore(limit)
        self.calls = 0
        self.max_active = 0
        self.active = 0

    async def fetch(self, url):
        async with self.sem:
            self.calls += 1
            self.active += 1
            self.max_active = max(self.max_active, self.active)
            try:
                page = SITE.get(url)
                if page is None:
                    raise FetchError(url, "404")
                await asyncio.sleep(page.delay)
                return page
            finally:
                self.active -= 1


class Crawler:
    def __init__(self, fetcher):
        self.fetcher = fetcher
        self.seen = set()
        self.errors = []
        self.order = []

    async def crawl(self, url):
        if url in self.seen:
            return
        self.seen.add(url)
        try:
            page = await asyncio.wait_for(self.fetcher.fetch(url), 0.5)
        except FetchError as e:
            self.errors.append(str(e))
            return
        except asyncio.TimeoutError:
            self.errors.append(url + ": timeout")
            return
        self.order.append(url)
        await asyncio.gather(*[self.crawl(link) for link in page.links])


async def main():
    f = Fetcher(2)
    c = Crawler(f)
    await c.crawl("/")
    print(sorted(c.order), f.calls, f.max_active <= 2)
    c2 = Crawler(Fetcher(1))
    await asyncio.gather(c2.crawl("/missing"), c2.crawl("/slow"), c2.crawl("/c"))
    print(sorted(c2.errors), c2.order)
    results = await asyncio.gather(*[Fetcher(3).fetch(u) for u in ["/b", "/zzz", "/c"]], return_exceptions=True)
    print([r.url if isinstance(r, Page) else (type(r).__name__, r.reason) for r in results])


asyncio.run(main())

async def fetch(n):
    async with lock() as conn:
        async for row in rows(conn):
            await handle(row)
    return [x async for x in rows(n)]

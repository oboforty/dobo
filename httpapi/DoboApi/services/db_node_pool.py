import asyncio
from contextlib import asynccontextmanager
from typing import AsyncGenerator

from dbobo import ServerNodeAsync, Table, int32, float32
from DoboApi.settings import settings
from dbobo.node import NodeCommandWrapper


class OboDBNodePool:
    def __init__(self):
        size = settings.DB.pool_size
        self._pool = asyncio.Queue(maxsize=size)
        self.size = size

    async def init(self):
        for i in range(self.size):
            db = ServerNodeAsync(
                host=f"{settings.DB.host}:{settings.DB.port}",
                tls_key=settings.DB.tls_key,
                tls_cert=settings.DB.tls_cert,
            )

            # TODO: should client join immediately or defer until 1st call?
            await db.connect()
            # TODO: log?
            await self._pool.put(db)

    @asynccontextmanager
    async def acquire(self):
        node = await self._pool.get()
        try:
            yield node
        finally:
            await self._pool.put(node)


node_pool: OboDBNodePool = None


async def get_node() -> AsyncGenerator[NodeCommandWrapper, None]:
    async with node_pool.acquire() as node:
        yield NodeCommandWrapper(node)


async def initialize():
    global node_pool
    node_pool = OboDBNodePool()
    await node_pool.init()

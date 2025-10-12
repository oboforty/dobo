import asyncio
from contextlib import asynccontextmanager
from typing import AsyncGenerator

from dbobo import ServerNodeAsync
from DoboApi.settings import settings
from dbobo.node import NodeCommandWrapper


class OboDBNodePool:
    def __init__(self):
        size = settings.DBClient.pool_size
        self._pool = asyncio.Queue(maxsize=size)
        self.size = size

    async def init(self):
        for i in range(self.size):
            db = ServerNodeAsync(
                host=f"{settings.DB.host}:{settings.DB.port}",
                tls_key=settings.DBClient.tls_key,
                tls_cert=settings.DBClient.tls_cert,
            )

            # TODO: should client join immediately or defer until 1st call?
            await db.connect()
            # TODO: log?
            await self._pool.put(db)

    @asynccontextmanager
    async def acquire(self) -> AsyncGenerator[ServerNodeAsync, None]:
        node = await self._pool.get()
        try:
            yield node
        finally:
            await self._pool.put(node)


node_pool: OboDBNodePool = OboDBNodePool()


async def get_node() -> AsyncGenerator[NodeCommandWrapper, None]:
    async with node_pool.acquire() as node:
        yield NodeCommandWrapper(node)


async def initialize_pool():
    await node_pool.init()

from time import time

from DoboApi.services.db import get_node
from DoboApi.settings import settings

table_metadata_cache: dict[str, tuple[dict, float]] = {}


async def get_table_metadata(table: str) -> dict:
    now = time()

    try:
        cfg, cached_time = table_metadata_cache[table]
        if now - cached_time <= settings.DBClient.metadata_cache_ttl:
            return cfg
    except KeyError:
        pass

    # Retrieve single TableInfo from Node & cache
    node = await anext(get_node())
    cfg = await node.get_table(table)
    table_metadata_cache[cfg['name']] = cfg, now

    return cfg


async def initialize_cache():
    node = await anext(get_node())
    tables: list[dict] = await node.list_tables()

    # TODO: log?
    now = time()

    for cfg in tables:
        table_metadata_cache[cfg['name']] = cfg, now

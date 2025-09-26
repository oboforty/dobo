from time import time

from DoboApi.settings import settings

table_metadata_cache: dict[str, tuple[dict, float]] = {}


async def get_table_metadata(table: str) -> dict:
    try:
        cfg, cached_time = table_metadata_cache[table]

        now = time()
        if now - cached_time <= settings.DBClient.metadata_cache_ttl:
            return cfg
    except KeyError:
        pass

    return None

    # TODO: Retrieve single TableInfo from Node & cache
    raise NotImplementedError("my cat looking at me like not giving it food every 2hrs is famine")


def initialize_cache(tables: list[dict]):
    # TODO: log?
    now = time()

    for table in tables:
        table_metadata_cache[table['name']] = table, now

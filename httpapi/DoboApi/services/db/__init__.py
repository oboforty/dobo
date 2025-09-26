from .node_pool import get_node, NodeCommandWrapper, node_pool, initialize_pool
from .table_cache import initialize_cache, get_table_metadata


async def initialize():
    await initialize_pool()

    node = await anext(get_node())
    tables = await node.list_tables()
    initialize_cache(tables)


__all__ = [
    'get_node',
    'get_table_metadata',
    'NodeCommandWrapper',
    'initialize',
]
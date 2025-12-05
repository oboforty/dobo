from fastapi.exceptions import RequestValidationError
from starlette.requests import Request

from .node_pool import get_node, NodeCommandWrapper, node_pool, initialize_pool
from .table_cache import initialize_cache, get_table_metadata


async def initialize():
    await initialize_pool()
    await initialize_cache()


async def get_current_table(request: Request) -> dict:
    table_name = request.path_params.get('table')
    if not table_name:
        table_name = request.query_params.get("table")

    if not table_name and request.method != "GET":
        try:
            body = await request.json()
            table_name = body.get("table")
        except (ValueError, RequestValidationError):
            pass

    if table_name:
        return await get_table_metadata(table_name)

    return None


__all__ = [
    'get_node',
    'get_table_metadata',
    'get_current_table',
    'NodeCommandWrapper',
    'initialize',
]
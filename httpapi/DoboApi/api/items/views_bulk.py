import random

from fastapi import Depends, APIRouter
from starlette.requests import Request

from DoboApi.api.mw import inject_db_conn

bulk_router = APIRouter(
    prefix='/tables/{table}/items',
    tags=["items bulk"],
    dependencies=[Depends(inject_db_conn)],
)


@bulk_router.get("/")
async def query_items(table: str, request: Request):
    """
    Query items
    """

    return {
        "table": table,
        "items": []
    }


@bulk_router.put("/")
async def upsert_items(table: str, request: Request):
    """
    Bulk Upsert items
    """

    return {
        "table": table,
        "items": []
    }


@bulk_router.patch("/")
async def update_items(table: str, request: Request):
    """
    Bulk Update items
    """

    return {
        "table": table,
        "items": []
    }


@bulk_router.delete("/")
async def remove_items(table: str, request: Request):
    """
    Bulk Remove items OR truncate
    """

    return {
        "table": table,
        "items": []
    }

import random

from fastapi import Depends, APIRouter
from starlette.requests import Request

from DoboApi.api.mw import inject_db_conn

items_router = APIRouter(
    prefix='/tables/{table}/items/{pkey}',
    tags=["items"],
    dependencies=[Depends(inject_db_conn)],
)


@items_router.get("/")
async def get_item(pkey, table: str):

    return {
        'pkey': pkey,
        'table': table
    }


@items_router.put("/")
async def put_item(pkey, table: str):

    return {
        'pkey': pkey,
        'table': table
    }


@items_router.patch("/")
async def update_item(pkey, table: str):

    return {
        'pkey': pkey,
        'table': table
    }


@items_router.delete("/")
async def remove_item(pkey, table: str):

    return {
        'pkey': pkey,
        'table': table
    }

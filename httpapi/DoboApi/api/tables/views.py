from fastapi import Depends, APIRouter, Body
from starlette.requests import Request

from DoboApi.api.mw import inject_current_table

tables_router = APIRouter(
    prefix='/tables',
    tags=['tables'],
    dependencies=[Depends(inject_current_table)],
)


@tables_router.get("/")
async def list_tables(request: Request):
    """
    LIST all Tables
    """
    return {
        "tables": []
    }


@tables_router.post("/")
async def create_table(table):
    """
    Create Table
    """

    return {
        'table': table,
    }


@tables_router.get("/{table}")
async def get_table(table):
    """
    Get Table information
    """
    # with checkout_entity(wid) as entity:
    #     entity.items['gold'] += 1

    # return entity.view
    return {
        'table': table,
    }


@tables_router.patch("/{table}")
async def edit_table_settings(table):
    """
    """
    return {
        'table': table,
    }


@tables_router.delete("/{table}")
async def drop_table(table):
    """
    """
    return {
        'table': table,
    }

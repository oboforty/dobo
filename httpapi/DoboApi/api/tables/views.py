from typing import Annotated
from contextvars import ContextVar

from fastapi import Depends, APIRouter, Body, HTTPException
from starlette.requests import Request

from DoboApi.api.mw import inject_db_conn

from dbobo import Table, NodeCommandWrapper


tables_router = APIRouter(
    prefix='/tables',
    tags=['tables'],
    dependencies=[Depends(inject_db_conn)],
)


@tables_router.get("/")
async def list_tables(request: Request):
    """
    LIST all Tables
    """
    db: NodeCommandWrapper = request.state.db
    tables = await db.list_tables()

    return {
        "tables": tables
    }


@tables_router.post("/")
async def create_table():
    """
    Create Table
    """

    return {
        'table': table,
    }


@tables_router.get("/{table}")
async def get_table(request: Request):
    """
    Get Table information
    """
    table: dict = request.state.current_table

    if table is None:
        raise HTTPException(status_code=404, detail="Table not found")

    # TODO: ITT: describe table? with more info flag?

    return table


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

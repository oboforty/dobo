import asyncio
import time
from typing import Annotated

from fastapi import Request, FastAPI, Depends
from fastapi.exceptions import RequestValidationError

from DoboApi.services.db import get_node, get_table_metadata, get_current_table, NodeCommandWrapper
from dbobo import Table


async def inject_db_conn(
    request: Request,
    node: Annotated[NodeCommandWrapper, Depends(get_node)],
    table: Annotated[dict, Depends(get_table_metadata)] = None
):
    """
    Injects current table name & node resource (connected from DB pool) to the request.

    NOTE: you can also connect a table to an endpoint function as such:
          async def my_request(
            request: Request,
            db: Annotated[NodeCommandWrapper, Depends(get_node)]
          ):
              pass
    """
    request.state.node = node
    request.state.current_table = table


def setup_middleware(app: FastAPI):
    @app.middleware("http")
    async def measure_request_time(request: Request, call_next):
        start_time = time.perf_counter()
        response = await call_next(request)
        process_time = time.perf_counter() - start_time
        response.headers["X-Process-Time"] = str(process_time)
        return response

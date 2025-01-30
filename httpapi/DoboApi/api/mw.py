import time

from fastapi import Request, FastAPI
from fastapi.exceptions import RequestValidationError


async def inject_current_table(request: Request):
    # TODO: load table metadata & inject?
    table_name = request.path_params.get('table')
    if not table_name:
        table_name = request.query_params.get("table")

    # table_name = None
    #
    # match request.method:
    #     case "GET":
    #     case _:
    #         try:
    #             body = await request.json()
    #             table_name = body.get("table")
    #         except (ValueError, RequestValidationError):
    #             pass
    #
    # request.state.table = table_name
    pass


def setup_middleware(app: FastAPI):
    @app.middleware("http")
    async def measure_request_time(request: Request, call_next):
        start_time = time.perf_counter()
        response = await call_next(request)
        process_time = time.perf_counter() - start_time
        response.headers["X-Process-Time"] = str(process_time)
        return response

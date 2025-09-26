from typing import Awaitable, Callable

from fastapi import FastAPI

from DoboApi.services.db import initialize
from DoboApi.settings import settings


def register_startup_event(
    app: FastAPI,
) -> Callable[[], Awaitable[None]]:
    """
    Actions to run on application startup.

    This function uses fastAPI app to store data
    in the state, such as db_engine.

    :param app: the fastAPI application.
    :return: function that actually performs actions.
    """

    @app.on_event("startup")
    async def _startup() -> None:
        app.middleware_stack = app.build_middleware_stack()

        await initialize()

    return _startup

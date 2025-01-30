from typing import Awaitable, Callable

from fastapi import FastAPI
from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine

from DoboApi.settings import settings

db_engine = None
db_session_factory = None


def get_db_factory():
    return db_session_factory


def setup_db() -> None:  # pragma: no cover
    """
    Creates connection to the database.

    This function creates SQLAlchemy engine instance,
    session_factory for creating sessions
    and stores them in the application's state property.

    :param app: fastAPI application.
    """
    engine = create_async_engine(str(settings.DB.db_url), echo=settings.db_echo)
    session_factory = async_sessionmaker(
        engine,
        expire_on_commit=False,
    )

    global db_engine, db_session_factory
    db_engine = engine
    db_session_factory = session_factory
    # app.state.db_engine = engine
    # app.state.db_session_factory = session_factory


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

    return _startup


def register_shutdown_event(
    app: FastAPI,
) -> Callable[[], Awaitable[None]]:  # pragma: no cover
    """
    Actions to run on application's shutdown.

    :param app: fastAPI application.
    :return: function that actually performs actions.
    """

    @app.on_event("shutdown")
    async def _shutdown() -> None:  # noqa: WPS430
        global db_engine
        await db_engine.dispose()

        pass  # noqa: WPS420

    return _shutdown


def new_db_session():
    engine = create_async_engine(str(settings.DB.db_url), echo=settings.db_echo)

    session_factory = async_sessionmaker(
        engine,
        expire_on_commit=False,
    )

    return session_factory()

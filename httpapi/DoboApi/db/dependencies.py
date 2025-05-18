from typing import AsyncGenerator

from sqlalchemy.ext.asyncio import AsyncSession
from starlette.requests import Request

from DoboApi.lifetime import get_db_factory


async def get_db_session(request: Request) -> AsyncGenerator[AsyncSession, None]:
    """
    Create and get database session.

    :param request: current request.
    :yield: database session.
    """
    db_session_factory = get_db_factory()
    session: AsyncSession = db_session_factory()

    try:  # noqa: WPS501
        yield session
    finally:
        await session.commit()
        await session.close()

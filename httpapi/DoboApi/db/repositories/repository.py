import uuid
from typing import Generic, TypeVar, Any, Iterable, List
from collections.abc import Callable

from fastapi import Depends

from sqlalchemy import BinaryExpression, select, SQLColumnExpression
from sqlalchemy.ext.asyncio import AsyncSession

from DoboApi.db import base
from DoboApi.db.dependencies import get_db_session


Model = TypeVar("Model", bound=base.Base)


class Repository(Generic[Model]):
    """Repository for performing database queries."""

    def __init__(self, model: type[Model], session: AsyncSession) -> None:
        self.model = model
        self.session = session

    async def create(self, **data) -> Model:
        instance = self.model(**data)
        self.session.add(instance)
        await self.session.commit()
        await self.session.refresh(instance)
        return instance

    async def get(self, pk: int | str | uuid.UUID | tuple) -> Model | None:
        return await self.session.get(self.model, pk)

    async def list(self) -> Iterable[Model]:
        q = select(self.model)

        return await self.session.scalars(q)

    async def filter(
        self,
        *expressions: BinaryExpression,
        col: SQLColumnExpression = None,
    ) -> List[Model]:
        q = select(col or self.model)
        if expressions:
            q = q.where(*expressions)
        return list(await self.session.scalars(q))

    async def find(
        self,
        *expressions: BinaryExpression,
        col: SQLColumnExpression = None,
    ) -> Model:
        q = select(col or self.model)
        if expressions:
            q = q.where(*expressions)
        return await self.session.scalar(q)

    async def delete(self, model: Model):
        await self.session.delete(model)


def get_repository(
    model: type[base.Base],
) -> Callable[[AsyncSession], Repository]:
    def func(dbsession: AsyncSession = Depends(get_db_session)):
        return Repository(model, dbsession)

    return func

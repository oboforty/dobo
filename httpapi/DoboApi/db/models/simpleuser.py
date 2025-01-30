from DoboApi.db.base import Base

import uuid

from sqlalchemy import ForeignKey, func
from sqlalchemy.dialects.postgresql import JSON
from sqlalchemy.orm import Mapped, mapped_column, relationship, validates
from sqlalchemy.sql.sqltypes import String, TIMESTAMP, UUID, Boolean


class SimpleUser(Base):
    __tablename__ = "users"

    """
    """
    uid: Mapped[str] = mapped_column(UUID(), primary_key=True)
    username: Mapped[str] = mapped_column(String(length=32))
    hashed_password: Mapped[str] = mapped_column(String(length=32))
    admin: Mapped[bool] = mapped_column(Boolean(), default=False)
    #
    # _is_new: bool = True

    def __init__(self, /, username: str, hashed_password: str):
        uid = str(uuid.uuid4())
        super().__init__(uid=uid, username=username, hashed_password=hashed_password)

    # @validates("uid")
    # def validate_write_once(self, key, value):
    #     if not self._is_new:
    #         raise ValueError(f"{key} is immutable")
    #     return value

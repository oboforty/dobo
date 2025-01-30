from dataclasses import dataclass
from typing import Literal


@dataclass
class User:
    id: str
    username: str
    token_type: Literal["access", "refresh", "fingerprint"]
    is_active: bool
    is_verified: bool
    admin: bool = False
    is_superuser: bool = False
    points: int = 0

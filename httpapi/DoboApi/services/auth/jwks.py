from functools import lru_cache

import ujson as json
import os
import random
from typing import override, Any, TypedDict

from fastapi import Depends
import jwt
from jwt.types import JWKDict
from starlette.responses import JSONResponse

from DoboApi.services.utils import Singleton
from DoboApi.settings import settings


class JWKPublicKey(TypedDict):
    # Header
    kty: str
    alg: str
    kid: str
    use: str

    # Key
    e: str
    n: str


class JWKSet(TypedDict):
    keys: list[JWKPublicKey | JWKDict]


class LocalJWKClient(jwt.PyJWKClient):
    """
    A JWKS client that fetches the data from locally,
    instead of a public OIDC JWKS endpoint.
    This way we save an HTTP call when fetching pub key at auth verify
    """

    def __init__(self,):
        # TODO: settings.oauth endpoint
        super().__init__(uri='127.0.0.1', cache_keys=True)

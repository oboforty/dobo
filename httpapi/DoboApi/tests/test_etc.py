import random
import string
from urllib.parse import urlparse, parse_qs

import httpx
import pytest
from fastapi import FastAPI
from httpx import AsyncClient
from sqlalchemy.ext.asyncio import AsyncSession

from DoboApi.db import Oauth2AuthCode
from DoboApi.tests.utils import setup_tests, fixtures
from DoboApi.tests.utils.setup_tests import get_user_header


async def test_spike(
    fastapi_app: FastAPI,
    client: AsyncClient,
    dbsession: AsyncSession
) -> None:
    """
    Complete auth flow (without user login)
    """
    pass

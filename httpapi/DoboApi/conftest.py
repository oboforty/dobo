import asyncio
from unittest import mock
from unittest.mock import AsyncMock, MagicMock
from typing import AsyncGenerator, Any, Literal

import pytest
import pytest_asyncio
from httpx import AsyncClient, ASGITransport
from fastapi import FastAPI

from DoboApi.application import get_app
from DoboApi.services.db import node_pool, get_node, get_table_metadata
from DoboApi.settings import settings
from dbobo.node import NodeCommandWrapper


@pytest_asyncio.fixture
async def client():
    app: FastAPI = get_app()
    """Create async test client."""
    async with AsyncClient(
        transport=ASGITransport(app=app), base_url="http://test"
    ) as ac:
        yield ac


class MockedServerNodeAsync:
    def __init__(self):
        self._responses: dict[tuple[int, str | None, tuple[bytes, ...] | None | Literal["_T__ANY__"]], tuple[bytes, ...]] = {}

    async def request(
        self, /, cmd: int, *,
        table: str = None,
        payload_format: int = None,
        dynamic_payload: list[bytes] = None,
        expected_payloads: int = 1
    ) -> tuple[bytes, ...] | None:
        if dynamic_payload:
            dynamic_payload = tuple(dynamic_payload)

        reqkey = cmd, table, dynamic_payload
        if reqkey not in self._responses:
            reqkey = cmd, table, "_T__ANY__"

        if reqkey in self._responses:
            resp = self._responses[reqkey]
        else:
            raise NotImplementedError(f"Node command {cmd} for table {table} not mocked!")

        if len(resp) != expected_payloads:
            raise Exception(f"Incorrect mocking: expected {expected_payloads} but got {len(resp)}")

        return resp

    def mock_response(
        self, /, cmd: int, *, table: str = None,
        dynamic_payload: list[bytes] = None,
        response: tuple[bytes, ...] | None = None
    ):
        if dynamic_payload:
            if dynamic_payload == Any:
                dynamic_payload = "_T__ANY__"
            else:
                dynamic_payload = tuple(dynamic_payload)

        self._responses[cmd, table, dynamic_payload] = response

    def clear(self):
        self._responses = {}


@pytest_asyncio.fixture
async def db_node_fixture():
    # minimum pool size to satisfy every async request
    assert settings.DBClient.pool_size >= 2

    conn = MockedServerNodeAsync()

    for _ in range(settings.DBClient.pool_size):
        await node_pool._pool.put(conn)

    return conn

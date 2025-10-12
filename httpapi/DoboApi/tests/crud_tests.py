from typing import Any

import pytest
from httpx import AsyncClient


@pytest.mark.asyncio
async def test_get_item(client, db_node_fixture):
    # Arrange
    db_node_fixture.mock_response(
        # get table cfg
        cmd=10, table="table1",
        response=(b'{"name":"table1", "key_type": "int64"}',)
    )
    db_node_fixture.mock_response(
        # get item
        cmd=20, table="table1", dynamic_payload=Any,
        response=(b'{"my":"item", "oof": 23523}', (2).to_bytes())
    )

    # Act
    response = await client.get("/tables/table1/items/123457?format=item")

    # Assert - special JSON item response is given
    assert response.json() == dict(
        item={"my":"item", "oof": 23523},
        table="table1",
        key=123457,
    )
    assert response.status_code == 200


@pytest.mark.asyncio
async def test_index_endpoint(client):
    """Test the root endpoint."""
    response = await client.get("/")
    assert response.status_code == 200
    data = response.json()
    assert "welcome" in data
    assert "version" in data
    assert data["welcome"] == "Dobo DB - HTTP Api"
    assert data["version"] == "0.1.0"

#
# # Example of testing multiple endpoints concurrently
# @pytest.mark.asyncio
# async def test_multiple_endpoints_concurrently(client: AsyncClient):
#     """Example of testing multiple endpoints concurrently using asyncio.gather."""
#     import asyncio
#
#     # Test multiple endpoints concurrently
#     tasks = [
#         client.get("/"),
#         client.get("/tables/table1/items/123457"),
#     ]
#
#     responses = await asyncio.gather(*tasks)
#
#     # Verify all responses
#     assert all(response.status_code == 200 for response in responses)
#     assert responses[0].json()["welcome"] == "Dobo DB - HTTP Api"
#     assert responses[1].json() == {"msg": "Hello World"}
#

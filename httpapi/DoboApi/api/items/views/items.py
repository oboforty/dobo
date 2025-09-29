import json
from json import JSONDecodeError
from typing import Annotated, Literal

from fastapi import Header, HTTPException
from starlette.requests import Request
from starlette.responses import Response

from DoboApi.api.items.router import router
from DoboApi.api.serialize import convert_key_from_params, get_file_meta
from dbobo import NodeCommandWrapper, ItemNotFoundError


@router.get("/{pkey}")
async def get_item(
    pkey: str,
    request: Request,
    format: Literal["item", "download"] | None = None,
    accept: Annotated[str | None, Header()] = None,
):
    table = request.state.current_table
    node: NodeCommandWrapper = request.state.node

    # Param validation
    if not table:
        raise HTTPException(status_code=404, detail="Table not found")
    key, key_bytes = convert_key_from_params(pkey, table['key_type'])

    # Fetch item from DB
    try:
        item = await node.get_item(table['name'], key_bytes)
    except ItemNotFoundError:
        raise HTTPException(status_code=404, detail="Item not found")

    can_json = accept in ("*/*", "application/json")
    if format == "item" and can_json:
        # OboDB item json schema
        try:
            value = json.loads(item.value)
        except JSONDecodeError:
            raise HTTPException(status_code=400, detail="Invalid JSON item expected")

        return {
            'table': table['name'],
            'key': key,
            'item': value
        }

    headers = {}
    filename, content_type = get_file_meta(key, table['key_type'], accept)
    if format == "download":
        headers['Content-Disposition'] = f'attachment; filename="{filename}"'
    return Response(content=item.value, media_type=content_type, headers=headers)


@router.put("/{pkey}")
async def put_item(
    pkey,
    request: Request,
    content_type: Annotated[str | None, Header()] = None,
):
    table = request.state.current_table
    node: NodeCommandWrapper = request.state.node

    # Param validation
    if not table:
        raise HTTPException(status_code=404, detail="Table not found")
    key, key_bytes = convert_key_from_params(pkey, table['key_type'])

    data = await request.body()

    # insert item
    try:
        await node.put_item(table['name'], key_bytes, data)
    except ItemNotFoundError:
        raise HTTPException(status_code=404, detail="Item not found")

    return {
        'key': pkey,
        'table': table['name'],
        'inserted': len(data)
    }


@router.patch("/{pkey}")
async def update_item(pkey, table: str):
    raise HTTPException(status_code=405, detail="Patch not implemented yet")


@router.delete("/{pkey}")
async def delete_item(pkey, request: Request):
    table = request.state.current_table
    node: NodeCommandWrapper = request.state.node

    # Param validation
    if not table:
        raise HTTPException(status_code=404, detail="Table not found")
    key, key_bytes = convert_key_from_params(pkey, table['key_type'])

    # Remove item
    try:
        await node.delete_item(table['name'], key_bytes)
    except ItemNotFoundError:
        raise HTTPException(status_code=404, detail="Item not found")

    return {
        'pkey': pkey,
        'table': table['name']
    }

import base64
from typing import Annotated, Literal

from fastapi import Header, HTTPException
from starlette.requests import Request
from starlette.responses import PlainTextResponse, JSONResponse

from DoboApi.api.items.router import router
from DoboApi.api.serialize import convert_key_from_params, get_file_meta

from dbobo import NodeCommandWrapper, ItemNotFoundError


@router.get("/{pkey}/base64")
async def get_item(
    pkey: str,
    request: Request,
    download: Literal["true", "1"] | None = None,
    accept: Annotated[str | None, Header()] = None,
):
    table = request.state.current_table
    node: NodeCommandWrapper = request.state.node

    # Param validation
    if not table:
        raise HTTPException(status_code=404, detail="Table not found")
    key, key_bytes = convert_key_from_params(pkey, table['key_type'])
    if accept not in ('*/*', 'text/plain', 'application/json'):
        raise HTTPException(status_code=400, detail=f"Invalid accept header for base64: {accept}")

    # Fetch item from DB
    try:
        item = await node.get_item(table['name'], key_bytes)
    except ItemNotFoundError:
        raise HTTPException(status_code=404, detail="Item not found")

    value = base64.b64encode(item.value)
    headers: dict[str, str] = {
        'X-Found-In': item.found_in.name,
    }

    if accept == "text/plain":
        if download:
            filename, _ = get_file_meta(key, table['key_type'], accept)
            headers['Content-Disposition'] = f'attachment; filename="{filename}"'
        return PlainTextResponse(value, headers=headers)
    else:
        return JSONResponse({
            'table': table['name'],
            'key': key,
            'item': value
        }, headers=headers)


@router.put("/{pkey}/base64")
async def put_item(
    pkey,
    request: Request,
):
    table = request.state.current_table
    node: NodeCommandWrapper = request.state.node

    # Param validation
    if not table:
        raise HTTPException(status_code=404, detail="Table not found")
    key, key_bytes = convert_key_from_params(pkey, table['key_type'])

    # insert item
    value = await request.body()
    data = base64.b64decode(value)

    try:
        await node.put_item(table['name'], key_bytes, data)
    except ItemNotFoundError:
        raise HTTPException(status_code=404, detail="Item not found")

    return {
        'key': pkey,
        'table': table['name'],
        'inserted': len(data)
    }

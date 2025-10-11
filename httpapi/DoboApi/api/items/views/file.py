from typing import Annotated, Optional

from fastapi import Header, HTTPException, File
from starlette.requests import Request
from starlette.responses import Response

from DoboApi.api.items.router import router
from DoboApi.api.serialize import convert_key_from_params, get_file_meta

from dbobo import NodeCommandWrapper, ItemNotFoundError


@router.get("/{pkey}/file")
async def download_file(
    pkey: str,
    table: str,
    request: Request,
    user_agent: Annotated[str | None, Header()] = None,
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

    # TODO: optimize & use streaming instead?

    filename, content_type = get_file_meta(key, table['key_type'], accept)

    return Response(
        content=item.value,
        media_type=content_type,
        headers={
            'Content-Disposition': f'attachment; filename="{filename}"',
            'X-Found-In': item.found_in,
        }
    )


@router.put("/{pkey}/file")
async def upload_file(
    pkey,
    table: str,
    request: Request,
    file: Optional[bytes] = File(None),
    user_agent: Annotated[str | None, Header()] = None,
    content_type: Annotated[str | None, Header()] = None,
):
    # TODO: optimize & use UploadFile instead?

    table = request.state.current_table
    node: NodeCommandWrapper = request.state.node

    # Param validation
    if not table:
        raise HTTPException(status_code=404, detail="Table not found")
    key, key_bytes = convert_key_from_params(pkey, table['key_type'])

    # insert item
    try:
        await node.put_item(table['name'], key_bytes, file)
    except ItemNotFoundError:
        raise HTTPException(status_code=404, detail="Item not found")

    return {
        'key': pkey,
        'table': table['name'],
        'inserted': len(file)
    }

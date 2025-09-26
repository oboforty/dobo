import base64
import json
import mimetypes
import os
import random
from typing import Annotated, Literal

from fastapi import Depends, APIRouter, Header, HTTPException
from starlette.requests import Request
from starlette.responses import Response, PlainTextResponse, FileResponse

from DoboApi.api.mw import inject_db_conn
from DoboApi.api.serialize import str2bytes_with_string_type

from dbobo import Table, NodeCommandWrapper, ItemNotFoundError
from dbobo.serialize import (
    KeyTypes, ValueTypes,
    convert_to_bytes, convert_from_bytes, TypeHintTypes
)


items_router = APIRouter(
    prefix='/tables/{table}/items/{pkey}',
    tags=["items"],
    dependencies=[Depends(inject_db_conn)],
)


@items_router.get("/")
async def get_item(
    request: Request,
    pkey: str, table: str,
    format: Literal["base64", "raw", "download"] | None = None,
    user_agent: Annotated[str | None, Header()] = None,
    x_item_format: Annotated[Literal["base64", "raw", "download"] | None, Header()] = None,
    accept: Annotated[str | None, Header()] = None,
):
    table = request.state.current_table
    if not table:
        raise HTTPException(status_code=404, detail="Table not found")

    # Param validation
    key, key_bytes = str2bytes_with_string_type(pkey, table['key_type'])
    value = None
    return_raw_file = format in ("raw", "download")
    if accept in ("*/*", None):
        content_type = "application/json"
    else:
        content_type = accept

    if not format:
        format = x_item_format or None

    # Fetch item from DB
    try:
        node: NodeCommandWrapper = request.state.node
        item = await node.get_item(table['name'], key_bytes)
        value = item.value
    except ItemNotFoundError:
        raise HTTPException(status_code=404, detail="Item not found")

    if format == "base64":
        # Handle special base64 response type
        if accept not in ('*/*', 'text/plain', 'application/json'):
            raise HTTPException(status_code=400, detail=f"Invalid accept header for base64 {accept}")

        value = base64.b64encode(value)
        return_raw_file = accept == "text/plain"
    elif format == "download":
        return_raw_file = True
    elif format is None and accept in ("*/*", "application/json"):
        # json takes a special format, unless X-Item-Format specifies raw json
        value = json.loads(value)
    else:
        return_raw_file = True

    if not return_raw_file:
        # serve based on response schema
        return {
            'table': table['name'],
            'key': key,
            'item': value
        }
    else:
        # serve raw value
        headers = {
            "X-Item-Format": str(format),
        }
        if format in ("raw", "download"):
            # smart guess file from extension
            if table['key_type'] == "string" and "." in key:
                _, ext = os.path.splitext(key)

                if accept == "*/*":
                    content_type = mimetypes.guess_type(key)[0] or "application/octet-stream"
            else:
                ext = mimetypes.guess_extension(content_type) or ".dat"

            if format == "download":
                headers['Content-Disposition'] = f'attachment; filename="download{ext}"'

        return Response(
            content=value,
            media_type=content_type,
            headers=headers
        )


@items_router.put("/")
async def put_item(
    pkey,
    table: str,
    user_agent: Annotated[str | None, Header()] = None,
    x_item_format: Annotated[Literal["base64", "raw"] | None, Header()] = None,
    content_type: Annotated[str | None, Header()] = None,
):

    return {
        'pkey': pkey,
        'table': table
    }


@items_router.patch("/")
async def update_item(pkey, table: str):

    return {
        'pkey': pkey,
        'table': table
    }


@items_router.delete("/")
async def remove_item(pkey, table: str):

    return {
        'pkey': pkey,
        'table': table
    }

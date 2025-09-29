from starlette.requests import Request

from DoboApi.api.items.router import router


@router.get("/")
async def query_items(table: str, request: Request):
    """
    Query items
    """

    return {
        "table": table,
        "items": []
    }


@router.put("/")
async def upsert_items(table: str, request: Request):
    """
    Bulk Upsert items
    """

    return {
        "table": table,
        "items": []
    }


@router.patch("/")
async def update_items(table: str, request: Request):
    """
    Bulk Update items
    """

    return {
        "table": table,
        "items": []
    }


@router.delete("/")
async def remove_items(table: str, request: Request):
    """
    Bulk Remove items OR truncate
    """

    return {
        "table": table,
        "items": []
    }

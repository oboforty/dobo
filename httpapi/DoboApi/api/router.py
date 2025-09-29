from fastapi.routing import APIRouter

from . import docs
from .tables.views import tables_router
from .items.views.items import router as items_router

api_router = APIRouter()


api_router.include_router(tables_router)
api_router.include_router(items_router)

api_router.include_router(docs.router)


@api_router.get("/", include_in_schema=True)
async def index():
    # TODO: include statistics
    return {
        "welcome": "Dobo DB - HTTP Api",
        "version": "0.1.0"
    }

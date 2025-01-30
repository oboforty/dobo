from fastapi import FastAPI
from fastapi.responses import UJSONResponse

from .docs.openapi import create_openapi_schema
from .router import api_router


app = FastAPI(
    root_path="/api",
    openapi_url="/openapi.json",
    default_response_class=UJSONResponse,
    #dependencies=[Depends(auth_dependency)]
)
app.include_router(router=api_router)


def openapi_gen():
    if app.openapi_schema:
        return app.openapi_schema

    app.openapi_schema = create_openapi_schema(
        routes=app.routes,
        version="1.0.0"
    )

    return app.openapi_schema


app.openapi = openapi_gen


__all__ = ['app']

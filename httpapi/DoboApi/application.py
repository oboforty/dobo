import os
from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from DoboApi.api.docs.openapi import create_openapi_schema
from DoboApi.api.mw import setup_middleware
from DoboApi.api.router import api_router

from DoboApi.lifetime import register_startup_event

APP_ROOT = Path(__file__).parent


def get_app() -> FastAPI:
    """
    Get FastAPI application.

    This is the main constructor of an application.

    :return: application.
    """
    app = FastAPI(
        title="DoboApi",
        version="1.0.0",
        docs_url=None,
        redoc_url=None,
        openapi_url="/api/openapi.json",
    )

    def openapi_gen():
        if app.openapi_schema:
            return app.openapi_schema

        app.openapi_schema = create_openapi_schema(
            routes=app.routes,
            version="1.0.0"
        )

        return app.openapi_schema

    app.openapi = openapi_gen

    register_startup_event(app)

    app.include_router(api_router, prefix='')

    if os.environ.get('HTTPAPI_ENV') != 'prod':
        app.mount(
            "/public",
            StaticFiles(directory=APP_ROOT / "public"),
            name="public",
        )

    return app

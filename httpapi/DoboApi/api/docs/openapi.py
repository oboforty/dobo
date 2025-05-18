from fastapi.openapi.utils import get_openapi


def create_openapi_schema(**kwargs):
    openapi_schema = get_openapi(
        title="Dobo DB API",
        summary="HTTP API for Dobo Database",
        description="",

        **kwargs
    )

    return openapi_schema

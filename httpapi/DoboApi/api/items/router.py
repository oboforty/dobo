from fastapi import APIRouter, Depends

from DoboApi.api.mw import inject_db_conn


router = APIRouter(
    prefix='/tables/{table}/items',
    tags=["items"],
    dependencies=[
        Depends(inject_db_conn),
    ],
)

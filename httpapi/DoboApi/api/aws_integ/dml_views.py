import json
from typing import Annotated

from fastapi import Depends, APIRouter, HTTPException
from starlette.requests import Request

from DoboApi.api.serialize import convert_key_from_params
from DoboApi.services.aws_integ.botoauth import (
    signv4_verify, aws_integ_parse_request,
    AWSRequest
)
from DoboApi.services.db import get_table_metadata, get_node
from dbobo import NodeCommandWrapper, ItemNotFoundError

ents_router = APIRouter(
    prefix='/aws/dynamodb',
    tags=['aws'],
)


# @TODO: temporal solution- integrate with DbOBO simple login later
def secret_getter(key):
    if key == "obo":
        return "hotmail"
    elif key == "oldalas":
        return "asdage1"
    else:
        return "dbobo"


@ents_router.get("/{sid}", include_in_schema=True)
async def index(
    request: Request,
    node: Annotated[NodeCommandWrapper, Depends(get_node)]
):
    # @TODO: secret getter - design proxy for auth AWS?

    aws_request = AWSRequest(method=request.method, url=request.url, data=await request.json(), params=None, headers=request.headers)
    integ: tuple[str, str] = signv4_verify(aws_request, secret_getter=secret_getter)

    if not integ:
        return {
            "__type": "com.amazon.coral.service#UnrecognizedClientException",
            "message": "The security token included in the request is invalid."
        }, 400

    target, target_args = aws_integ_parse_request(*integ, request)

    print("@@@ TODO: ", target, target_args)

    # TODO: mapper for different services?
    #       getattr(srv, target)(**target_args)
    #       get table_name, key, value from request
    table_name = target_args.get("TableName")
    table = await get_table_metadata(table_name)


    # TODO: just copy the DDBClient from tomcru alltogether
    #       & check if it still works

    match target:
        case "GetItem":
            key, key_bytes = convert_key_from_params(target_args['Key'], table['key_type'])

            try:
                resp = await node.get_item(table_name, key_bytes)
            except ItemNotFoundError:
                raise HTTPException(status_code=404, detail="Item not found")


    # if isinstance(response, (dict, list)):
    #     return json.dumps(response, separators=(',', ":"))
    # else:
    #     return response

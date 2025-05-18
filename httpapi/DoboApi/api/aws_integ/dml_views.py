import random
import json
import logging
from time import sleep

from botocore.auth import (
    SigV4Auth, SigV4QueryAuth,
    S3SigV4Auth, S3SigV4QueryAuth, S3SigV4PostAuth
)
from botocore.awsrequest import AWSRequest
from botocore.credentials import Credentials

from fastapi import Depends, APIRouter
from starlette.requests import Request

from .aws_integ import signv4_verify, aws_integ_parse_request


ents_router = APIRouter(
    prefix='/aws/dynamodb',
    tags=['interface'],
)


@ents_router.get("/{sid}", include_in_schema=True)
async def index(request: Request):
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

    # todo: later: how to handle region parameter, -> pass to partition
    #       that is not passed to `target` method, but only to parse_request?

    # if hasattr(srv, 'aws_integ_parse_request'):
    #     target = srv.aws_integ_parse_request(target, integ[1], request, target_args)

    response = getattr(srv, target)(**target_args)

    # if hasattr(srv, 'aws_integ_parse_response'):
    #     response = srv.aws_integ_parse_response(*integ, response)

    if isinstance(response, (dict, list)):
        return json.dumps(response, separators=(',', ":"))
    else:
        return response

import typing
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession
from starlette.responses import Response

from DoboApi.db import User, Oauth2Client
from typing import TypedDict, Literal
from . import fixtures
from DoboApi.services.oidc.oauth import OIDCAuth
from DoboApi.services.oidc.jwks import LocalJWKClient, JWKServer
from ...services.webauth import WebAuth


async def get_user_header(
    user: User,
    client: Oauth2Client,
    type_token: Literal['access', 'refresh'] = 'access'
):
    auth = OIDCAuth(
        jwks_client=LocalJWKClient(
            jwks_server=JWKServer()
        ),
        jwks_server=JWKServer(),
    )
    token = auth.create_jwt(user=user, client=client, type_token=type_token)

    return {
        'Authorization': f'JWT {token}'
    }


async def get_user_cookies(
    user: User,
):
    class TestSetupMockedResponse(Response):
        def __init__(self):
            self.cookies = {}

        def set_cookie(
            self,
            key: str, value: str = "",
            max_age: typing.Optional[int] = None,
            expires: typing.Optional[typing.Union[datetime, str, int]] = None,
            path: str = "/",
            domain: typing.Optional[str] = None,
            secure: bool = False,
            httponly: bool = False,
            samesite: typing.Optional[Literal["lax", "strict", "none"]] = "lax",
        ) -> None:
            self.cookies[key] = value

    auth = WebAuth(req=None)
    token = auth.create_jwt(user=user)

    mock_resp = TestSetupMockedResponse()
    auth.set_access_cookies(token, response=mock_resp)

    return mock_resp.cookies


class TestSetup(TypedDict):
    user: User
    client: Oauth2Client
    user_headers: dict
    user_cookies: dict


async def the_usual_please(dbsession: AsyncSession) -> TestSetup:
    ts = dict()

    ts['user'] = await fixtures.create_test_user(dbsession)
    ts['client'] = await fixtures.create_client(dbsession)
    ts['user_headers'] = await get_user_header(ts['user'], ts['client'])
    ts['user_cookies'] = await get_user_cookies(ts['user'])

    assert ts['user'].is_active and ts['user'].is_verified

    return ts

from fastapi import Depends
from fastapi.security import OpenIdConnect
from starlette.exceptions import HTTPException

# from DoboApi.services.auth.oauth import OAuth
from DoboApi.services.auth.user import User
from DoboApi.settings import settings


oidc_scheme = OpenIdConnect(
    scheme_name='Dave1 OIDC',
    openIdConnectUrl=settings.OAuth.issuer,
    auto_error=False
)


async def get_current_user(
    token_header: str = Depends(oidc_scheme),
    # auth: OAuth = Depends(),
):
    return None

    if not token_header:
        raise HTTPException(
            status_code=401,
            detail="Missing auth header"
        )

    auth._get_jwt_from_headers(token_header) # noqa: bullshit return type
    claims = auth.get_raw_jwt()
    token_type = claims.get('type')

    if token_type == 'access':
        # settings.OAuth.audience
        raise NotImplementedError('dave1 access token')
        auth._verify_jwt_in_request(auth._token, token_type, 'headers')
    elif token_type == 'fingerprint':
        # TODO: check DB for claims?
        print("@@@@ kuki kaki", claims)

    return User(
        id=claims['sub'],
        username=claims['username'],
        admin=claims.get('admin', False),
        is_superuser=claims.get('sn', False),
        points=claims.get('points', 0),
        token_type=token_type,
        is_active=True,
        is_verified=True
    )

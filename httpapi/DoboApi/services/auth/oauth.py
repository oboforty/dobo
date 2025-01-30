# from functools import lru_cache
# from typing import Literal, Dict, Union, Optional
#
# from fastapi import Depends
# from fastapi_another_jwt_auth import AuthJWT
# from fastapi_another_jwt_auth.exceptions import RevokedTokenError
# from jwt.algorithms import has_crypto
#
# from DoboApi.settings import settings
# from DoboApi.services.auth.oauth_cfg import JWTAuthConfigFactory
#
#
# JWT_TOKEN = Dict[str, Union[str, int, bool]]
#
#
# class OAuth(AuthJWT):
#     # hack to get the kid of an encode request
#     # We'll have to rely on this until `fastapi-another-jwt-auth`
#     # is abandoned just like `fastapi-jwt-auth` was,
#     # in which case we'll add more elegant asymmetric key support to our own fork
#     # Determines current JWK key id being used
#     # TODO: use a different one (change strictly after encrypt)
#     current_kid: str = None
#     _token: str = None
#     _token_claims: JWT_TOKEN = None
#
#     # jwks_client: LocalJWKClient = None
#
#     def __init__(
#         self,
#         # jwks_client: LocalJWKClient = Depends(),
#     ):
#         # self.jwks_client = jwks_client
#         self.load_config(JWTAuthConfigFactory(
#             settings.OAuth.algorithm,
#             settings.OAuth.issuer,
#             settings.OAuth.audience
#         ))
#
#         super().__init__(req=None, res=None)
#
#     def _verified_token(
#         self,
#         encoded_token: str,
#         issuer: Optional[str] = None
#     ) -> Dict[str, Union[str, int, bool]]:
#         if self._token_claims:
#             return self._token_claims
#
#         # Cache the token claims so they can be reused
#         self._token_claims = super()._verified_token(encoded_token, issuer=issuer)
#         return self._token_claims
#
#     def _get_secret_key(self, alg: str, process: Literal["decode", "encode"]):
#         """
#         Fetches pub/priv key pair
#         Overridden to match OIDC issuer behaviour.
#         Fetches previously set request.state.auth_token
#         """
#         if alg != self._algorithm or not has_crypto:
#             raise ValueError("Algorithm must be RS512 present with crypto extra pack")
#         elif process != "decode":
#             raise ValueError("encode")
#
#         raise NotImplementedError("dave1 JWT/RSA OAuth not yet implemented!")
#
#         signing_key = self.jwks_client.get_signing_key_from_jwt(self._token)
#
#         # Used at verifying a JWT. Fetch existing public key from JWKS
#         return signing_key.key
#
#     def _check_token_is_revoked(self, raw_token: JWT_TOKEN):
#         if False:
#             raise RevokedTokenError(status_code=401, message="Token has been revoked")

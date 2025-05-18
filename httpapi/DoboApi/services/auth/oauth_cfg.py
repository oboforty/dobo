from typing import Literal


class JWTAuthConfigFactory:
    """
    Automatically preconfigures OIDCAuth and WebAuth
    """

    def __init__(
        self,
        alg: Literal["RS512", "RS256", "HS512", "HS256"],
        issuer: str,
        audience: str,
        loc: str = 'headers'
    ):
        self.cfg = [
            ('authjwt_token_location', {loc}),

            # JWT
            ('authjwt_header_type', 'JWT'),
            ('authjwt_algorithm', alg),
            ('authjwt_decode_algorithms', [alg]),
            ('authjwt_decode_issuer', issuer),
            ('authjwt_decode_audience', audience),
        ]

    def __call__(self):
        return self.cfg

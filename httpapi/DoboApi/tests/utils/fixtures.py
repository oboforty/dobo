from typing import Literal

from sqlalchemy.ext.asyncio import AsyncSession

from DoboApi.db import User, Oauth2Client, UserRepository
from DoboApi.services.autogen import assign_face
from DoboApi.services.users import UserManager


async def create_test_user(
    dbsession: AsyncSession,
    role: Literal["user", "admin", "sa"] = "user",
    password: str = "DoboApi",
    **kwargs
) -> User:
    user_db = UserRepository(dbsession, User)
    mgr = UserManager(user_db)

    kwargs['is_superuser'] = role == "sa"
    kwargs['admin'] = role != "user"

    return await user_db.create(dict(
        email='test@gmail.com',
        username='Davesoliders',
        hashed_password=mgr.password_helper.hash(password),
        is_active=True,
        is_verified=True,
        face=assign_face(),
        points=10000,
        **kwargs
    ))


async def create_client(
    dbsession: AsyncSession,
    name='Test Client',
    client_id='TT'
) -> Oauth2Client:
    client = Oauth2Client(
        name=name,
        client_id=client_id,
        client_secret='fgberbw4v3583v4v34f43f34fg3',
        redirect_uri='http://myclient.com/oauth2/code',
        client_uri='http://myclient.com/',
        scope="profile",
        etc={}
    )
    dbsession.add(client)
    await dbsession.commit()

    return client

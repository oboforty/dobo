import asyncio
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).parent.parent))

from DoboApi.db import SimpleUser, Repository
from DoboApi.lifetime import new_db_session


async def create_test_ent():
    async with new_db_session() as session:
        repo = Repository(SimpleUser, session)

        repo.session.add(SimpleUser(
            sid="faszarcu",
            wid="world1",
            iso='co1',
            stype='debug'
        ))
        await repo.session.commit()


async def main():
    await create_test_ent()


if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())
    loop.close()

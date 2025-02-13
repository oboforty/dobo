import asyncio
import os.path

from dbobo import ServerNodeAsync, Table, int32, float32


TABLE_NAME = "table1"
KEYPATH = "/home/rajmund_csombordi/obodb/"


async def main():
    db = ServerNodeAsync(
        host="127.0.0.1:2480",
        tls_cert=os.path.join(KEYPATH, "nodekey.crt"),
        tls_key=os.path.join(KEYPATH, "nodekey.key"),
    )

    async with db as node:
        tables = set(await node.list_tables())

        print("Tables:")
        print("\n".join(sorted(tables)))
        print("---------------")

        if TABLE_NAME not in tables:
            await node.create_table({
                "name": TABLE_NAME,
                # "key_type": "int64"
            })

        table = Table[int](node, "table1")
        # table.type_hint = int

        print("Inserting new item")
        val = b"""{"name":"Rajmund", "age": 350, "data":{"attr1": "asdasd", "adas": 123}, "teso": false}"""
        await table.put_item(123456, val)

        # val = b"""{"name":"Rajmund", "age": 350, "data":{"attr1": "asdasd", "adas": 123}, "teso": false}"""
        # await table.put_item(123123, val)

        # Should automatically convert value json bytes to dict
        item = await table.get_item(123456, dict)

        print("Retrieved item:", item.key)
        print(item.value)


if __name__ == "__main__":
    asyncio.run(main())

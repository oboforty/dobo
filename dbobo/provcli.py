import asyncio
import json
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
        item_val = {"name":"Rajmund", "age": 350, "data":{"attr1": "asdasd", "adas": 123}, "teso": False}
        item_content = json.dumps(item_val).encode("utf8")
        item_key = 123456

        await table.put_item(item_key, item_content)

        item = await table.get_item(item_key)
        assert item.key == item_key, f"Key mismatch: {item.key} != {item_key}"
        assert item.value == item_content, f"Value mismatch:\n{item.value}\n!=\n{item_val}"

        # Should automatically convert value json bytes to dict
        item = await table.get_item(item_key, dict)
        assert item.key == item_key, f"Key mismatch: {item.key} != {item_key}"
        assert item.value == item_val, f"Value mismatch:\n{item.value}\n!=\n{item_val}"

        print("Retrieved item:", item.key, type(item.value))
        print(item.value)


if __name__ == "__main__":
    asyncio.run(main())

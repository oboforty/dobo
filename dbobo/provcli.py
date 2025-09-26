import asyncio
import json
import os.path
import random

from dbobo import ServerNodeAsync, Table
from dbobo.node import ItemNotFoundError

TABLE_NAME = "table1"
KEYPATH = "/home/rajmund_csombordi/obodb/"


async def main():
    db = ServerNodeAsync(
        host="127.0.0.1:2480",
        tls_cert=os.path.join(KEYPATH, "nodekey.crt"),
        tls_key=os.path.join(KEYPATH, "nodekey.key"),
    )

    async with db as node:
        item_val = {
            "name": "Rajmund",
            "age": random.randint(1, 9999),
            "data": {
                "attr1": "asdasd",
                "adas": 123
            },
            "teso": False
        }
        item_key = 123456

        tables = await node.list_tables()
        table_names = set()

        print("Tables:")
        for table in tables:
            print(f"- {table['name']}[{table['key_type']}] \n")
            table_names.add(table['name'])
        print("---------------")

        if TABLE_NAME not in table_names:
            await node.create_table({
                "name": TABLE_NAME,
                # "key_type": "int64"
            })

        table = Table[int](node, TABLE_NAME)
        table.type_hint = int

        print("Inserting new item")
        item_content = json.dumps(item_val).encode("utf8")

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

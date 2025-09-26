from .node import NodeCommandWrapper
from .serialize import KeyTypes, ValueTypes, TypeHintTypes, convert_from_bytes, convert_to_bytes


class ItemGeneric[P: KeyTypes, V: ValueTypes]:
    key: P
    value: V
    # metadata: dict[str, Any]

    def __init__(self, key: P, value: V):
        self.key = key
        self.value = value

    def __repr__(self):
        return self.__str__()

    def __str__(self):
        return f"ItemGeneric[{P.__name__}, {V.__name__}]<{self.key} -> {self.value}>"


class Table[P: KeyTypes]:
    def __init__(self, node: NodeCommandWrapper, table_name: str):
        self.node = node
        self.table_name = table_name
        self.type_hint: TypeHintTypes = None

    async def get_item[V: ValueTypes](self, key: P, return_type: TypeHintTypes=None) -> ItemGeneric[P, V]:
        key_bytes = convert_to_bytes(key, self.type_hint)
        item = await self.node.get_item(self.table_name, key_bytes)

        value: V = convert_from_bytes(item.value, return_type)

        return ItemGeneric[P, V](
            key=key,
            value=value,
        )

    async def put_item[V: ValueTypes](self, key: P, value: V) -> V:
        key_bytes = convert_to_bytes(key, self.type_hint)
        value_bytes = convert_to_bytes(value)

        await self.node.put_item(self.table_name, key_bytes, value_bytes)

        return value


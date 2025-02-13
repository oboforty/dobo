import dataclasses
import json
import struct
from typing import Literal, TypeAliasType

from .node import NodeCommandWrapper

type KeyTypes = int | float | str | bytes
type ValueTypes = int | float | str | bytes | dict | list | set | tuple

type TypeHintTypes = Literal["float32", "int32"] | type[ValueTypes] | None

# "type" variables to mark golang specific key types
int32 = "int32"
float32 = "float32"


class ItemGeneric[P: KeyTypes, V: ValueTypes]:
    key: P
    value: V
    # metadata: dict[str, Any]

    def __init__(self, key: P, value: V):
        self.key = key
        self.value = value


class Table[P: KeyTypes]:
    def __init__(self, node: NodeCommandWrapper, table: str):
        self.node = node
        self.table = table
        self.type_hint: TypeHintTypes = None

    async def get_item[V: ValueTypes](self, key: P, return_type: TypeHintTypes=None) -> ItemGeneric[P, V]:
        key_bytes = convert_bytes(key, self.type_hint)
        item = await self.node.get_item(self.table, key_bytes)

        value_bytes: V = convert_from_bytes(item.value, return_type)

        return ItemGeneric[P, V](
            key=key,
            value=value_bytes,
        )

    async def put_item[V: ValueTypes](self, key: P, value: V) -> V:
        key_bytes = convert_bytes(key, self.type_hint)
        value_bytes = convert_bytes(value)

        await self.node.put_item(self.table, key_bytes, value_bytes)

        return value


def convert_bytes(data: ValueTypes, type_hint: TypeHintTypes = None) -> bytes:
    """
    Key Types mapping:

    ======= ======= =======
    python  obodb   golang
    ======= ======= =======
    int     int64   int64
    int32   int32   int32
    float   float64 float64
    float32 float32 float32
    string  string  string
    bytes   bytes   []byte
    ======= ======= =======

    Value Types mapping:

    ======= =======
    python  obodb
    ======= =======
    int     packed int64
    int32   packed int32
    float   packed float64
    float32 packed float32
    string  bytes
    bytes   bytes
    dict    json bytes
    set     json list
    list    json list
    tuple   json list
    ======= =======
    """

    # TODO: force format instead of json?
    #       parquet, pbuf, asn1?

    if isinstance(data, int):
        if type_hint == int32:
            return struct.pack("!I", data)
        else:
            return struct.pack("!Q", data)
        # return struct.pack("!B", data)
    elif isinstance(data, float):
        if type_hint == int32:
            return struct.pack("!f", data)
        else:
            return struct.pack("!d", data)
    elif isinstance(data, str):
        return data.encode("utf-8")
    elif isinstance(data, bytes):
        return data
    elif isinstance(data, (dict, set, list, tuple)):
        return json.dumps(data, default=custom_encoder).encode("utf-8")
    else:
        raise TypeError(f"Unsupported data type: {type(data)}")


def convert_from_bytes[V: ValueTypes](data: bytes, type_hint: TypeHintTypes = None) -> V:
    # data_type = convert_from_bytes.__annotations__

    if type_hint == int:
        return struct.pack("!Q", data)
    elif type_hint == int32:
        return struct.pack("!I", data)
    elif type_hint == float:
        return struct.unpack("!d", data)[0]
    elif type_hint == float32:
        return struct.unpack("!f", data)[0]
    elif type_hint == str:
        return data.decode("utf-8")
    elif type_hint == bytes or type_hint is None:
        return data
    elif type_hint in (dict, set, list, tuple):
        return json.loads(data)
    else:
        raise TypeError(f"Unsupported data type: {type_hint}")


def custom_encoder(obj):
    # TODO: store DS types like in DynamoDB for tuple & set?
    if isinstance(obj, (tuple, set)):
        return list(obj)
    return obj

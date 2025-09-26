import base64
import json

from dbobo.serialize import (
    TypeHintTypes,
    int32, float32,
    convert_from_bytes, convert_to_bytes,
    KeyTypes, ValueTypes,
)


def string_type_to_type(type_name: str = None) -> TypeHintTypes:
    if type_name == "string":
        return str
    elif type_name == "bytes" or type_name is None or type_name == "base64":
        return bytes
    elif type_name in ("int", "int64", "int32", "long"):
        return int
    elif type_name in ("float", "float64", "float32", "double"):
        return float
    elif type_name == "bool":
        return bool
    elif type_name == "set":
        return set
    elif type_name == "tuple":
        return tuple
    elif type_name == "list":
        return list
    elif type_name == "dict":
        return dict
    else:
        raise TypeError(f"Unsupported data type name: {type_name}")


# def type_to_string(type_hint: TypeHintTypes = None) -> str:
#     if type_hint == int:
#         return "int64"
#     elif type_hint == int32:
#         return "int32"
#     elif type_hint == float:
#         return "float64"
#     elif type_hint == float32:
#         return "float32"
#     elif type_hint == str:
#         return "string"
#     else:
#         return type_hint.__name__
#     # elif type_hint == bytes:
#     #     return "bytes"
#     # elif type_hint in (dict, set, list, tuple):
#     # else:
#     #     raise TypeError(f"Unsupported data type: {type_hint}")


def str2bytes_with_string_type(value: str, type_hint_str: str) -> tuple[KeyTypes, bytes]:
    """
    @TODO:  explain the need for this & stuff
            API <--> dbobo type conversion
    """
    type_hint: TypeHintTypes = string_type_to_type(type_hint_str)

    if type_hint_str == "base64":
        # API provides base64, but DB looks up byte value
        orig_value = value
        bytes_value: bytes = base64.b64decode(value)
    # TODO: should we bother with providing true bytes in GET request at all?
    # elif type_hint == bytes:
    #     orig_value = value.encode('utf-8')
    else:
        # primitive python type conversion from str
        orig_value = type_hint(value)
        bytes_value = convert_to_bytes(orig_value, type_hint=type_hint)

    return orig_value, bytes_value

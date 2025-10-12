import asyncio
import dataclasses
import json
import ssl
import struct
from asyncio import StreamReader, StreamWriter
from enum import Enum, IntEnum
from types import TracebackType
from typing import Literal, Self, Type, AsyncContextManager, Any, Iterable


class RequestError(Exception):
    def __init__(self, cmd, code, *args):
        self.command_code = cmd
        self.error_code = code
        super().__init__(*args)

    def __repr__(self):
        return self.__str__()

    def __str__(self):
        return f"CMD={self.command_code!r}, ERR={self.error_code!r}, {self.args[0]})"


class ItemNotFoundError(RequestError):
    pass


class ServerNodeAsync:
    def __init__(self, host: str, *, tls_cert: str, tls_key: str):
        ss = host.split(":")

        self.ssl_context = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
        self.ssl_context.load_cert_chain(certfile=tls_cert, keyfile=tls_key)
        self.ssl_context.check_hostname = False
        self.ssl_context.verify_mode = ssl.CERT_NONE # @TODO...

        self.host: str = ss[0]
        self.port: int = int(ss[1]) if len(ss) > 1 else 2480

        self.reader: StreamReader | None = None
        self.writer: StreamWriter | None = None

        self.retries = 10
        self.reconnect_delay = 1

    async def __aenter__(self) -> 'NodeCommandWrapper':
        await self.connect()

        return NodeCommandWrapper(self)

    async def __aexit__(self, exc_type: Type[BaseException], exc_value: BaseException, traceback: TracebackType) -> bool:
        await self.disconnect()

    async def connect(self) -> None:
        self.reader, self.writer = await asyncio.open_connection(self.host, self.port, ssl=self.ssl_context)

    async def disconnect(self) -> None:
        #TODO: send BYE signal?

        self.writer.close()
        await self.writer.wait_closed()

    def write_dynamic(self, payload: bytes, *, recv_kl: Literal[1, 4] = 4):
        if recv_kl == 1:
            data_length: bytes = struct.pack("!B", len(payload))
        elif recv_kl == 4:
            data_length: bytes = struct.pack("!I", len(payload))
        else:
            raise Exception(f"Unknown recv_kl: {recv_kl}")

        self.writer.write(data_length)
        self.writer.write(payload)

    async def recv_dynamic(self, *, recv_kl: Literal[1, 4] = 4) -> bytes:
        data_length_bytes = await self.reader.read(recv_kl)

        if recv_kl == 1:
            data_length: int = struct.unpack("!B", data_length_bytes)[0]
        elif recv_kl == 4:
            data_length: int = struct.unpack("!I", data_length_bytes)[0]
        else:
            raise Exception(f"Unknown recv_kl: {recv_kl}")

        return await self.reader.read(data_length)

    async def request(
        self, /, cmd: int, *,
        table: str = None,
        payload_format: int = None,
        dynamic_payload: list[bytes] = None,
        expected_payloads: int = 1
    ) -> tuple[bytes, ...] | None:
        # TODO: retries? does the aio lib implement it?

        try:
            # Request
            self.writer.write(struct.pack("!B", cmd))
            if table:
                self.write_dynamic(table.encode('ascii'), recv_kl=1)

            # TODO: refactor this into dynamic payload? as only 1 operation uses it
            if payload_format:
                self.writer.write(struct.pack("!B", payload_format))

            if dynamic_payload:
                for data in dynamic_payload:
                    self.write_dynamic(data, recv_kl=4)

            await self.writer.drain()

            # Response
            resp_cmd: int = struct.unpack("!B", await self.reader.read(1))[0]
            resp_ok: int = struct.unpack("!B", await self.reader.read(1))[0]

            if not resp_ok:
                err_code: int = struct.unpack("!B", await self.reader.read(1))[0]

                # Handle error
                # TODO: refine err reporting...
                errmsg = await self.recv_dynamic()
                if errmsg == b'??':
                    errmsg = f"unknown error"
                else:
                    errmsg = json.loads(errmsg)

                if err_code == 4:
                    raise ItemNotFoundError(resp_cmd, err_code, errmsg)
                else:
                    raise RequestError(resp_cmd, err_code, errmsg)
            else:
                resp_payloads: list[bytes] = []
                for i in range(expected_payloads):
                    resp_payloads.append(await self.recv_dynamic())
                return tuple(resp_payloads)

        except asyncio.CancelledError:
            # TODO: Add logging error
            print("Connection task cancelled.")
            pass
        except (ConnectionResetError, ConnectionRefusedError, OSError) as e:
            await asyncio.sleep(self.reconnect_delay)
        except Exception as e:
            print(f"Unexpected error: {e}")
            await asyncio.sleep(self.reconnect_delay)


class FindStatus(IntEnum):
    FOUND_STATUS_UNKNOWN = 0
    FOUND_AT_MEM = 1
    FOUND_AT_BLOOM = 2
    FOUND_AT_SS = 3
    NOT_FOUND = 4


@dataclasses.dataclass
class Item:
    key: bytes
    value: bytes
    found_in: FindStatus
    # metadata: dict[str, Any]


class NodeCommandWrapper:
    def __init__(self, conn: ServerNodeAsync):
        self.conn = conn

    async def list_tables(self) -> list[dict]:
        tables_json, = await self.conn.request(
            cmd=4,
            expected_payloads=1
        )

        return json.loads(tables_json)

    async def get_table(self, table: str) -> dict:
        table_json, = await self.conn.request(
            cmd=10,
            table=table,
            expected_payloads=1
        )
        return json.loads(table_json)

    async def create_table(self, cfg: dict):
        resp, = await self.conn.request(
            cmd=11,
            payload_format=1, # cfg type = json
            dynamic_payload=[
                json.dumps(cfg).encode('ascii')
            ],
            expected_payloads=1
        )
        print("Create table response: ", resp)

    async def get_item(self, table: str, key: bytes) -> Item:
        value, found_in = await self.conn.request(
            cmd=20,
            table=table,
            dynamic_payload=[key],
            expected_payloads=2
        )

        return Item(key=key, value=value, found_in=FindStatus.from_bytes(found_in))

    async def put_item(self, table: str, key: bytes, value: bytes):
        resp, = await self.conn.request(
            cmd=21,
            table=table,
            dynamic_payload=[key, value],
            expected_payloads=1
        )
        print("put item response: ", resp)

    async def delete_item(self, table: str, key: bytes):
        resp, = await self.conn.request(
            cmd=23,
            table=table,
            dynamic_payload=[key],
            expected_payloads=1
        )
        print("delete item response: ", resp)

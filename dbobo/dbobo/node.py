import asyncio
import dataclasses
import json
import ssl
import struct
from asyncio import StreamReader, StreamWriter
from socket import socket
from types import TracebackType
from typing import Literal, Self, Type, AsyncContextManager, Any, Iterable

CMD_ERR = 2


class ServerNodeAsync:
    def __init__(self, host: str, *, tls_cert: str, tls_key: str):
        ss = host.split(":")
        self.tls_cert = tls_cert
        self.tls_key = tls_key

        self.host: str = ss[0]
        self.port: int = int(ss[1]) if len(ss) > 1 else 2480

        self.reader: StreamReader | None = None
        self.writer: StreamWriter | None = None

    async def __aenter__(self) -> 'NodeCommandWrapper':
        await self.connect()

        return NodeCommandWrapper(self)

    async def __aexit__(self, exc_type: Type[BaseException], exc_value: BaseException, traceback: TracebackType) -> bool:
        await self.disconnect()

    async def connect(self) -> None:
        ssl_context = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
        ssl_context.load_cert_chain(certfile=self.tls_cert, keyfile=self.tls_key)
        ssl_context.check_hostname = False
        ssl_context.verify_mode = ssl.CERT_NONE

        self.reader, self.writer = await asyncio.open_connection(self.host, self.port, ssl=ssl_context)

    async def disconnect(self) -> None:
        #TODO: send BYE signal?

        self.writer.close()
        await self.writer.wait_closed()

    # async def send_command(self, cmd: int, table: str = None, payload: list[bytes] = None):
        # self.writer.write(struct.pack("!B", cmd))
        #
        # if table:
        #     await self.send_dynamic(table.encode('ascii'), recv_kl=1)
        # if payload:
        #     for data in payload:
        #         await self.send_dynamic(data, recv_kl=4)

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
    ) -> tuple[bytes, ...]:
        # TODO: wrap in exception

        # Request
        self.writer.write(struct.pack("!B", cmd))
        if table:
            self.write_dynamic(table.encode('ascii'), recv_kl=1)

        if payload_format:
            self.writer.write(struct.pack("!B", payload_format))

        if dynamic_payload:
            for data in dynamic_payload:
                self.write_dynamic(data, recv_kl=4)

        await self.writer.drain()

        # Response
        resp_cmd: int = struct.unpack("!B", await self.reader.read(1))[0]

        if resp_cmd == CMD_ERR:
            # Handle error
            # TODO: refine err reporting...
            errmsg = await self.recv_dynamic()
            errmsg = json.loads(errmsg)

            raise Exception(errmsg)
        else:
            resp_payloads: list[bytes] = []
            for i in range(expected_payloads):
                resp_payloads.append(await self.recv_dynamic())
            return tuple(resp_payloads)


@dataclasses.dataclass
class Item:
    key: bytes
    value: bytes
    # metadata: dict[str, Any]


class NodeCommandWrapper:
    def __init__(self, conn: ServerNodeAsync):
        self.conn = conn

    async def list_tables(self) -> list[str]:
        tables_json, = await self.conn.request(
            cmd=4
        )

        return json.loads(tables_json)

    async def create_table(self, cfg: dict) -> dict:
        asd, = await self.conn.request(
            cmd=11,
            payload_format=1, # cfg type = json
            dynamic_payload=[
                json.dumps(cfg).encode('ascii')
            ]
        )
        print(asd)

    async def get_item(self, table: str, key: bytes) -> Item:
        value, = await self.conn.request(
            cmd=20,
            table=table,
            dynamic_payload=[key],
            expected_payloads=1
        )

        return Item(key=key, value=value)

    async def put_item(self, table: str, key: bytes, value: bytes):
        await self.conn.request(
            cmd=21,
            table=table,
            dynamic_payload=[
                key,
                value
            ],
            expected_payloads=0
        )

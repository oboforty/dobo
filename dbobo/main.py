import asyncio
import ssl
import struct
import json

KEYPATH = "/home/rajmund_csombordi/obodb/"

CMD_OK = 1
CMD_ERR = 2
CMD_LIST_TABLES = 4
CMD_CREATE_TABLE = 11
CMD_DROP_TABLE = 13
CMD_GET_ITEM = 20
CMD_PUT_ITEM = 21
CMD_UPD_ITEM = 22
CMD_DEL_ITEM = 23


async def send_command(writer, cmd, payload=None):
    writer.write(struct.pack("!B", cmd))
    if payload:
        writer.write(struct.pack("!I", len(payload)))
        writer.write(payload.encode())
    await writer.drain()


async def read_response(reader):
    data_length_bytes = await reader.read(4)
    data_length = struct.unpack("!I", data_length_bytes)[0]
    data = await reader.read(data_length)
    return data.decode()


async def main():
    ssl_context = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
    ssl_context.load_cert_chain(certfile=KEYPATH + "nodekey.crt", keyfile=KEYPATH + "nodekey.key")
    ssl_context.check_hostname = False
    ssl_context.verify_mode = ssl.CERT_NONE

    reader, writer = await asyncio.open_connection("127.0.0.1", 2480, ssl=ssl_context)
    print(f"Connected to: {writer.get_extra_info('peername')}")

    # Send LIST_TABLES command
    await send_command(writer, CMD_LIST_TABLES)
    tables_json = await read_response(reader)
    tables = json.loads(tables_json)

    print("Tables:")
    for table in tables:
        print('->', table)
    print("---------------")

    if "table1" not in tables:
        print("Creating table table1")
        settings = json.dumps({"name": "table1"})
        await send_command(writer, CMD_CREATE_TABLE, settings)
        new_table_cfg = await read_response(reader)
        print("New table created! Config:")
        print(new_table_cfg)
    else:
        table_name = "table1"
        writer.write(struct.pack("!B", len(table_name)))
        writer.write(table_name.encode())
        await writer.drain()

    writer.close()
    await writer.wait_closed()
    print("Connection closed.")

asyncio.run(main())

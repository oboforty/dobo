import struct
import sys
import os
import io
import gzip
import time

import tabulate


def list_sum_file(filepath):
  parts = ["metadata", "statistics", "sparse_idx"]

  part = parts.pop(0)
  print("--", part.upper(), "--")

  table = []
  with open(filepath, 'rb') as fh:
    text_reader = io.TextIOWrapper(fh, encoding='utf-8')

    for line in text_reader:
      line = line.rstrip('\n')
      if line == "------":
        part = parts.pop(0)
        print("--", part.upper(), "--")
        if part == "parse_idx":
          break
        continue
      print(line)

    fb = text_reader.detach()

    while True:
      kyln = fh.read(4)
      if len(kyln) == 0:
        break

      key_len, = struct.unpack('!I', kyln)
      part_key, = struct.unpack('!I', fh.read(key_len))
      offset, = struct.unpack('!I', fh.read(4))
      table.append([part_key, offset])

  print(tabulate.tabulate(table, headers=["part key", "byte offset"], tablefmt="grid"))


def list_idx_file(filepath, limit=float('inf'), iterate=False):
  table = []
  prev_offset = 0
  i = 0

  if iterate:
    print(f"{"part key":<10}{"offset":<10}{"inter offset":<10}")

  with open(filepath, 'rb') as fh:
    while True:
      i += 1
      kyln = fh.read(4)
      if len(kyln) == 0:
        break

      key_len, = struct.unpack('!I', kyln)
      part_key, = struct.unpack('!I', fh.read(key_len))
      offset, = struct.unpack('!I', fh.read(4))
      inter_offset, = struct.unpack('!I', fh.read(4))

      if iterate:
        print(f"\r{part_key:<10}{offset:<10}{inter_offset:<10}", end="")
      else:
        if offset != prev_offset:
          print("block offset =", prev_offset)
          print(tabulate.tabulate(table, headers=["part key", "decomp offset"], tablefmt="fancy_grid"))
          print("\n")

          table = []
          prev_offset = offset

        table.append([part_key, inter_offset])

      if i > limit:
        break
  
  if iterate:
    print("\nDone")
  else:
    print(tabulate.tabulate(table, headers=["part key", "decomp offset"], tablefmt="fancy_grid"))


def list_dat_file(filepath, limit=float('inf'), iterate=False):
  i = 0
  block_id = -1
  table = []
  iter_speed = iterate

  if iterate:
    print(f"{"part key":<15}{"data":<20}{"#blc":<6}{"offset":<10}{"blclen":<10}{"inter offset":<10}")

  with open(filepath, 'rb') as fh:
    while i < limit:
      # Read block length
      block_length_bytes = fh.read(4)
      if len(block_length_bytes) == 0:
        break
      block_length = struct.unpack('!I', block_length_bytes)[0]

      # Create gzip reader for the block
      comp_reader = gzip.GzipFile(fileobj=io.BytesIO(fh.read(block_length)))
      block_id += 1
      inter_off = 0
      j = 0

      if iterate:
        # keep last item
        print("")

      try:
        while True:
          # Read key length and key
          key_len_bytes = comp_reader.read(4)
          if len(key_len_bytes) == 0:
            break

          key_len = struct.unpack('!I', key_len_bytes)[0]
          part_key = struct.unpack('!I', comp_reader.read(key_len))[0]
          
          value_len = struct.unpack('!I', comp_reader.read(4))[0]
          value = comp_reader.read(value_len)

          # inter block offset
          inter_off += key_len + value_len

          if iterate:
            vstr = value.decode('utf-8')
            block_off = fh.tell()
            print(f"\r{part_key:<15}{vstr:<20}{block_id:<6}{block_off:<10}{block_length:<10}{inter_off:<10}", end="")
            # time.sleep(iter_speed)
          else:
            table.append([part_key, value])

          if j == 0:
            # keep first item
            vstr = value.decode('utf-8')
            block_off = fh.tell()
            print(f"\r{part_key:<15}{vstr:<20}{block_id:<6}{block_off:<10}{block_length:<10}{inter_off:<10}", end="")
            print("\n  ...")

          j += 1
          i += 1
          if i >= limit:
            break
      finally:
        comp_reader.close()
      
      if i >= limit:
        break

  if iterate:
    print("\nDone")
  else:
    print(tabulate.tabulate(table, headers=["part key", "value"], tablefmt="fancy_grid"))


def find_in_dat_file(filepath, to_find):
  i = 0
  block_id = -1
  print(f"{"part key":<15}{"data":<20}{"#blc":<6}{"offset":<10}{"blclen":<10}{"inter offset"}")

  with open(filepath, 'rb') as fh:
    while True:
      # Read block length
      block_length_bytes = fh.read(4)
      if len(block_length_bytes) == 0:
        break
      block_length = struct.unpack('!I', block_length_bytes)[0]

      # Create gzip reader for the block
      comp_reader = gzip.GzipFile(fileobj=io.BytesIO(fh.read(block_length)))
      block_id += 1
      inter_off = 0
      j = 0

      try:
        while True:
          key_len_bytes = comp_reader.read(4)
          if len(key_len_bytes) == 0:
            # print last item in block
            vstr = value.decode('utf-8')
            print(f"{part_key:<15}{vstr:<20}{block_id:<6}{block_off:<10}{block_length:<10}{inter_off}")
            break

          key_len = struct.unpack('!I', key_len_bytes)[0]
          part_key = struct.unpack('!I', comp_reader.read(key_len))[0]

          value_len = struct.unpack('!I', comp_reader.read(4))[0]
          value = comp_reader.read(value_len)

          # inter block offset
          inter_off += key_len + value_len

          if j == 0 or to_find == part_key:
            # print first item in block
            vstr = value.decode('utf-8')
            block_off = fh.tell()
            if to_find == part_key:
              print(f" >{part_key:<13}", end="")
            else:
              print(f"{part_key:<15}", end="")
            print(f"{vstr:<20}{block_id:<6}{block_off:<10}{block_length:<10}{inter_off}")

          j += 1
          i += 1
      except Exception as e:
        print("@@ ", e)
      finally:
        comp_reader.close()

  print("\nDone")


if __name__ == "__main__":
  filepath = sys.argv[1]
  _, ext = os.path.splitext(filepath)

  # TODO: add cli arg parser
  iterate = None#0.035

  if ext == '.sum':
    list_sum_file(filepath)
  elif ext == '.idx':
    list_idx_file(filepath, iterate=iterate)
  elif ext == '.dat':
    try:
      search_key = sys.argv[2]
    except Exception as e:
      search_key = None
    
    if search_key:
      find_in_dat_file(filepath, to_find=int(search_key))
    else:
      list_dat_file(filepath, iterate=iterate)

import struct
import sys
import os
import io

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


def list_idx_file(filepath, limit=float('inf')):
  table = []
  prev_offset = 0
  i = 0

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

      if offset != prev_offset:
        print("block offset =", prev_offset)
        print(tabulate.tabulate(table, headers=["part key", "decomp offset"], tablefmt="fancy_grid"))
        print("\n")

        table = []
        prev_offset = offset

      table.append([part_key, inter_offset])

      if i > limit:
        break
  print(tabulate.tabulate(table, headers=["part key", "decomp offset"], tablefmt="fancy_grid"))


def list_dat_file(filepath):
  pass


if __name__ == "__main__":
  _, filepath = sys.argv
  _, ext = os.path.splitext(filepath)

  if ext == '.sum':
    list_sum_file(filepath)
  elif ext == '.idx':
    list_idx_file(filepath, limit=100)
  elif ext == '.dat':
    list_dat_file(filepath, limit=100)

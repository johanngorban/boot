#!/usr/bin/env python3
"""
Постобработка прошивки для загрузчика STM32F103.

Принимает финальный .bin, проверяет магическое число в заголовке образа,
считает CRC32/ISO-HDLC от тела (всех байт после 512-байтового заголовка)
и записывает размер тела и CRC в заголовок по фиксированным смещениям.
Файл правится на месте.

Раскладка заголовка (image_metadata_t, packed, little-endian):
    offset  0  uint32  magic
    offset  4  uint32  size   <- проставляется здесь
    offset  8  uint32  crc    <- проставляется здесь
    offset 12  uint8   version_major
    offset 13  uint8   version_minor
    offset 14  uint8   version_patch
    offset 15  ...     reserved
"""

import struct
import sys
import zlib

IMAGE_METADATA_SIZE = 512
IMAGE_MAGIC_NUMBER = 0xAAAAAAAA
FIRMWARE_SLOT_SIZE = 48 * 1024  # см. memory_map.h

OFFSET_MAGIC = 0
OFFSET_SIZE = 4
OFFSET_CRC = 8


def patch(path: str) -> None:
    with open(path, "rb") as f:
        data = bytearray(f.read())

    if len(data) < IMAGE_METADATA_SIZE:
        raise SystemExit(
            f"error: файл {path} меньше заголовка "
            f"({len(data)} < {IMAGE_METADATA_SIZE} байт)"
        )

    magic = struct.unpack_from("<I", data, OFFSET_MAGIC)[0]
    if magic != IMAGE_MAGIC_NUMBER:
        raise SystemExit(
            f"error: неверное магическое число: "
            f"0x{magic:08X}, ожидалось 0x{IMAGE_MAGIC_NUMBER:08X}"
        )

    body = data[IMAGE_METADATA_SIZE:]
    size = len(body)
    if size == 0:
        raise SystemExit("error: тело образа пустое")

    if size + IMAGE_METADATA_SIZE > FIRMWARE_SLOT_SIZE:
        # Загрузчик отвергнет такой образ в image_verify
        raise SystemExit(
            f"error: образ не помещается в слот: "
            f"{size + IMAGE_METADATA_SIZE} > {FIRMWARE_SLOT_SIZE} байт"
        )

    crc = zlib.crc32(body) & 0xFFFFFFFF

    struct.pack_into("<I", data, OFFSET_SIZE, size)
    struct.pack_into("<I", data, OFFSET_CRC, crc)

    with open(path, "wb") as f:
        f.write(data)

    print(f"{path}: size={size} байт, crc=0x{crc:08X}")


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit(f"usage: {sys.argv[0]} <firmware.bin>")
    patch(sys.argv[1])


if __name__ == "__main__":
    main()

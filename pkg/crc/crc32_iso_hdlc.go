package crc

func Crc32IsoHdlc(data []uint8) (uint32, error) {
	if data == nil {
		return 0, ErrDataIsNil
	}

	var crc uint32 = 0xFFFFFFFF
	for i := range len(data) {
		crc ^= uint32(data[i])
		for _ = range 8 {
			if (crc & 1) == 1 {
				crc = (crc >> 1) ^ 0xEDB88320
			} else {
				crc >>= 1
			}
		}
	}
	crc = (crc ^ 0xFFFFFFFF) & 0xFFFFFFFF
	return crc, nil
}

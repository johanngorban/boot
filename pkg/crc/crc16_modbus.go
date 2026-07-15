package crc

func Crc16Modbus(data []uint8) (uint16, error) {
	if data == nil {
		return 0, ErrDataIsNil
	}

	var crc uint16 = 0xFFFF
	for i := range len(data) {
		crc ^= uint16(data[i])
		for _ = range 8 {
			if (crc & 1) == 1 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	crc &= 0xFFFF
	return crc, nil
}

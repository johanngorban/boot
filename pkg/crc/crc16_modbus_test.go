package crc

import "testing"

func TestEmptyCrc16Modbus(t *testing.T) {
	testData := make([]uint8, 0)
	expectedCrc := uint16(0xFFFF)

	realCrc, err := Crc16Modbus(testData)
	if err != nil {
		t.Fatalf("crc error: %v", err)
	}

	if expectedCrc != realCrc {
		t.Fatalf("wrong crc: 0x%02X expected 0x%02X", realCrc, expectedCrc)
	}
}

func TestZeroCrc16Modbus(t *testing.T) {
	testData := []uint8{0x00}
	expectedCrc := uint16(0x40BF)

	realCrc, err := Crc16Modbus(testData)
	if err != nil {
		t.Fatalf("crc error: %v", err)
	}

	if expectedCrc != realCrc {
		t.Fatalf("wrong crc: 0x%02X expected 0x%02X", realCrc, expectedCrc)
	}
}

func TestNilDataCrc16Modbus(t *testing.T) {
	if _, err := Crc16Modbus(nil); err != ErrDataIsNil {
		t.Fatalf("wrong error: %v, expected ErrDataIsNil (%v)", err, ErrDataIsNil)
	}
}

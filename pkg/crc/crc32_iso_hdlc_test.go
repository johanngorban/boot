package crc

import (
	"encoding/hex"
	"testing"
)

func TestEmptyCrc32IsoHdlc(t *testing.T) {
	testData := make([]uint8, 0)
	expectedCrc := uint32(0x00000000)

	realCrc, err := Crc32IsoHdlc(testData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if realCrc != expectedCrc {
		t.Errorf("wrong crc: 0x%02X expected 0x%02X", realCrc, expectedCrc)
	}
}

func TestZeroCrc32IsoHdlc(t *testing.T) {
	testData := []uint8{0}
	expectedCrc := uint32(0xD202EF8D)

	realCrc, err := Crc32IsoHdlc(testData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if realCrc != expectedCrc {
		t.Errorf("wrong crc: 0x%02X, expected 0x%02X", realCrc, expectedCrc)
	}
}

func TestCorrectCrc32IsoHdlc(t *testing.T) {
	data, _ := hex.DecodeString("ab34998bde78a9890a090909c9898d")
	expected := uint32(0x57271CE2)

	real, err := Crc32IsoHdlc(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if real != expected {
		t.Errorf("wrong crc: 0x%02X, expected 0x%02X", real, expected)
	}
}

func TestVeryLongCrc32IsoHdlc(t *testing.T) {
	data, _ := hex.DecodeString("7A3F9B2C4E8D1F6A0B5C3E7F9A2D4B6C8E0F1A3B5C7D9E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C")
	expected := uint32(0xF2A3D78B)

	real, err := Crc32IsoHdlc(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if real != expected {
		t.Errorf("wrong crc: 0x%02X, expected 0x%02X", real, expected)
	}
}

func TestRandom256DigitCrc32IsoHdlc(t *testing.T) {
	data, _ := hex.DecodeString(
		"7A3F9B2C4E8D1F6A0B5C3" +
			"E7F9A2D4B6C8E0F1A3B5C" +
			"7D9E1F2A3B4C5D6E7F8A9" +
			"B0C1D2E3F4A5B6C7D8E9F" +
			"0A1B2C3D4E5F6A7B8C9D0" +
			"E1F2A3B4C5D6E7F8A9B0C" +
			"7A3F9B2C4E8D1F6A0B5C3" +
			"E7F9A2D4B6C8E0F1A3B5C" +
			"7D9E1F2A3B4C5D6E7F8A9" +
			"B0C1D2E3F4A5B6C7D8E9F" +
			"0A1B2C3D4E5F6A7B8C9D0" +
			"E1F2A3B4C5D6E7F8A9B0C")

	expected := uint32(0x880F5100)

	real, err := Crc32IsoHdlc(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if real != expected {
		t.Errorf("wrong crc: 0x%02X, expected 0x%02X", real, expected)
	}
}

func TestNilDataCrc32IsoHdlc(t *testing.T) {
	if _, err := Crc32IsoHdlc(nil); err != ErrDataIsNil {
		t.Errorf("wrong error: %v, expected ErrDataIsNil (%v)", err, ErrDataIsNil)
	}
}

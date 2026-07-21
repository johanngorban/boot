package fwp

import (
	"encoding/binary"
	"stm32-bootctl/pkg/crc"
)

const (
	fwpSof        = 0xAA
	fwpAck        = 0x06
	fwpNak        = 0x15
	fwpDataSize   = 256
	fwpHeaderSize = 5
)

type fwpPacketType byte

const (
	fwpTypeStart fwpPacketType = 0x01
	fwpTypeData                = 0x02
	fwpTypeEnd                 = 0x03
)

func buildPacket(packetType fwpPacketType, seq uint16, payload []byte) ([]byte, error) {
	if len(payload) > fwpDataSize {
		return nil, ErrPayloadOverflow
	}

	header := make([]byte, fwpHeaderSize)
	header[0] = byte(packetType)
	binary.LittleEndian.PutUint16(header[1:3], seq)
	binary.LittleEndian.PutUint16(header[3:5], uint16(len(payload)))

	body := append(header, payload...)
	crc, err := crc.Crc16Modbus(body)
	if err != nil {
		return nil, err
	}

	packet := make([]byte, len(body)+2)
	binary.LittleEndian.PutUint16(packet[len(body):len(body)+1], crc)
	return packet, nil
}

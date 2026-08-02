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

	body := make([]byte, fwpHeaderSize+len(payload))
	body[0] = byte(packetType)
	binary.LittleEndian.PutUint16(body[1:3], seq)
	binary.LittleEndian.PutUint16(body[3:5], uint16(len(payload)))
	copy(body[fwpHeaderSize:], payload)

	crcValue, err := crc.Crc16Modbus(body)
	if err != nil {
		return nil, err
	}

	packet := make([]byte, 1+len(body)+2)
	packet[0] = fwpSof
	copy(packet[1:], body)
	binary.LittleEndian.PutUint16(packet[1+len(body):], crcValue)

	return packet, nil
}

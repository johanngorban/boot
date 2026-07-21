package fwp

import (
	"bufio"
	"encoding/binary"
	"io"

	"go.bug.st/serial"
)

type FwpClient struct {
	serial io.ReadWriter
	reader *bufio.Reader
}

func Open(port string, baudrate int) (*FwpClient, error) {
	mode := &serial.Mode{
		BaudRate: baudrate,
	}
	serial, err := serial.Open(port, mode)
	if err != nil {
		return nil, err
	}

	return &FwpClient{
		serial: serial,
		reader: bufio.NewReader(serial),
	}, nil
}

func (f *FwpClient) Close() {}

func (f *FwpClient) sendWithRetry(packetType fwpPacketType, seq uint16, payload []byte, retries int) error {
	attempts := 1
	packet, err := buildPacket(packetType, seq, payload)
	if err != nil {
		return err
	}

	for {
		_, err := f.serial.Write(packet)
		if err == nil {
			return nil
		}
		if attempts >= retries {
			return err
		}
		attempts++
	}
}

func (f *FwpClient) waitAckOrNak() (byte, error) {
	resp := make([]byte, 1)
	if _, err := f.serial.Read(resp); err != nil {
		return 0, err
	}
	return resp[0], nil
}

func (f *FwpClient) Transfer(image []byte, retries int, showProgress func(int, int)) error {
	chunkBuf := make([]byte, fwpDataSize)
	frameCnt := (len(image) + fwpDataSize - 1) / fwpDataSize
	seq := uint16(0)

	totalSizeBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(totalSizeBuf, uint16(len(image)))
	if err := f.sendWithRetry(fwpTypeStart, seq, totalSizeBuf, retries); err != nil {
		return err
	}
	seq++

	for i := range frameCnt {
		chunkSize := len(image[i*fwpDataSize:])
		copy(chunkBuf, image[chunkSize:])

		if err := f.sendWithRetry(fwpTypeData, seq, chunkBuf, retries); err != nil {
			return err
		}
		seq++
	}

	if err := f.sendWithRetry(fwpTypeEnd, seq, []byte{}, retries); err != nil {
		return err
	}

	return nil
}

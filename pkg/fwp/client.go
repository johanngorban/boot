package fwp

import (
	"bufio"
	"encoding/binary"
	"fmt"
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
	packet, err := buildPacket(packetType, seq, payload)
	if err != nil {
		return err
	}

	var last byte
	for attempt := 1; attempt <= retries; attempt++ {
		if _, err := f.serial.Write(packet); err != nil {
			return err
		}

		b, err := f.waitAckOrNak()
		if err != nil {
			return err
		}
		last = b
		if b == fwpAck {
			return nil
		} else if b == fwpNak {
			continue
		} else {
			return fmt.Errorf("unexpected FWP reply: 0x%02X (type=0x%02X, seq=0x%02X)", b, packetType, seq)
		}
	}
	return fmt.Errorf("packet failed after %d retries (type=0x%02X, seq=%d, last 0x%02X)", retries, packetType, seq, last)
}

func (f *FwpClient) waitAckOrNak() (byte, error) {
	resp := make([]byte, 1)
	if _, err := f.serial.Read(resp); err != nil {
		return 0, err
	}
	return resp[0], nil
}

func (f *FwpClient) Transfer(image []byte, retries int, showProgress func(int, int)) error {
	if showProgress == nil {
		showProgress = func(a int, b int) {}
	}

	seq := uint16(0)
	totalSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(totalSizeBuf, uint32(len(image)))
	if err := f.sendWithRetry(fwpTypeStart, seq, totalSizeBuf, retries); err != nil {
		return err
	}
	seq++

	for offset := 0; offset < len(image); offset += fwpDataSize {
		end := min(len(image), offset+fwpDataSize)
		if err := f.sendWithRetry(fwpTypeData, seq, image[offset:end], retries); err != nil {
			return err
		}
		seq++
		showProgress(end, len(image))
	}

	if err := f.sendWithRetry(fwpTypeEnd, seq, []byte{}, retries); err != nil {
		return err
	}

	return nil
}

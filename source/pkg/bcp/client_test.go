package bcp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/iotest"
)

func formatBytes(data []byte) string {
	var builder strings.Builder
	for i, b := range data {
		if i > 0 {
			builder.WriteString(" ")
		}
		_, _ = builder.WriteString(fmt.Sprintf("0x%02X", b))
	}
	return builder.String()
}

func TestSend(t *testing.T) {
	buf := &bytes.Buffer{}
	c := Client{
		serial: buf,
		reader: bufio.NewReader(buf),
	}

	expectedRaw := []byte{bcpSof, 0x01, 0x04, 0xF0, 0x9B, 0x83, 0xA1, 0x6D, 0x12}
	testRequest, err := NewRequest(0x01, []byte{0xF0, 0x9B, 0x83, 0xA1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.Send(testRequest); err != nil {
		t.Fatalf("send: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), expectedRaw) {
		t.Fatalf("wrong data sent\nwant: %v\ngot: %v",
			formatBytes(expectedRaw),
			formatBytes(buf.Bytes()))
	}
}

func TestRecv(t *testing.T) {
	buf := bytes.NewBuffer([]byte{bcpSof, 0x04, 0x00, 0x02, 0x01, 0x00, 0x90, 0x75})
	c := Client{
		serial: buf,
		reader: bufio.NewReader(buf),
	}

	expectedData := []byte{0x01, 0x00}

	r, err := c.Recv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Command != VersionCommand {
		t.Fatalf("wrong command\nwant: 0x%02X\ngot: 0x%02X", uint8(VersionCommand), uint8(r.Command))
	}
	if r.Status != bcpOk {
		t.Fatalf("wrong status\nwant: 0x%02X\ngot: 0x%02X", bcpOk, r.Status)
	}
	if !bytes.Equal(r.Data, expectedData) {
		t.Fatalf("wrong data received\nwant: %v\ngot: %v",
			formatBytes(expectedData),
			formatBytes(r.Data))
	}
}

func TestRecvBadCrc(t *testing.T) {
	buf := bytes.NewBuffer([]byte{bcpSof, 0x04, 0x00, 0x02, 0x01, 0x00, 0xFF, 0xFF})
	c := Client{
		serial: buf,
		reader: bufio.NewReader(buf),
	}

	if _, err := c.Recv(); !errors.Is(err, ErrBadCrc) {
		t.Fatalf("want: %v\ngot: %v", ErrBadCrc, err)
	}
}

func TestRecvDeviceStatus(t *testing.T) {
	buf := bytes.NewBuffer([]byte{bcpSof, 0x01, 0x03, 0x00, 0xF0, 0x20})
	c := Client{
		serial: buf,
		reader: bufio.NewReader(buf),
	}

	_, err := c.Recv()
	if err == nil {
		t.Fatalf("want error for status 0x%02X, got nil", bcpBadCrc)
	}
	t.Logf("mapBcpStatus(0x%02X) -> %v", bcpBadCrc, err)
}

func TestRecvTwoFrames(t *testing.T) {
	buf := bytes.NewBuffer([]byte{
		bcpSof, 0x04, 0x00, 0x02, 0x01, 0x00, 0x90, 0x75,
		bcpSof, 0x03, 0x00, 0x00, 0xC0, 0x81,
	})
	c := Client{
		serial: buf,
		reader: bufio.NewReader(buf),
	}

	first, err := c.Recv()
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if first.Command != VersionCommand {
		t.Fatalf("first: wrong command 0x%02X", uint8(first.Command))
	}

	second, err := c.Recv()
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if second.Command != RunCommand {
		t.Fatalf("second: wrong command 0x%02X", uint8(second.Command))
	}
}

func TestRecvPartialReads(t *testing.T) {
	buf := bytes.NewBuffer([]byte{bcpSof, 0x04, 0x00, 0x02, 0x01, 0x00, 0x90, 0x75})
	c := Client{
		serial: buf,
		reader: bufio.NewReader(iotest.OneByteReader(buf)),
	}

	if _, err := c.Recv(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

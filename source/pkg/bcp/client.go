package bcp

import (
	"bufio"
	"errors"
	"io"

	"go.bug.st/serial"
)

type Client struct {
	port   serial.Port
	reader *bufio.Reader
}

func Open(portName string, baudRate int) (*Client, error) {
	mode := &serial.Mode{
		BaudRate: int(baudRate),
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}

	return &Client{
		port:   port,
		reader: bufio.NewReader(port),
	}, nil
}

func (c *Client) Close() error {
	return c.port.Close()
}

func readFrame(r *bufio.Reader) ([]byte, error) {
	// Waiting for SOF
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == bcpSof {
			break
		}
	}

	// Read header
	header := make([]byte, bcpResponseHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	// Read data of length in the header
	data := make([]byte, header[2])
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	crcBuf := make([]byte, 2)
	if _, err := io.ReadFull(r, crcBuf); err != nil {
		return nil, err
	}

	frame := []byte{bcpSof}
	frame = append(frame, header...)
	frame = append(frame, data...)
	frame = append(frame, crcBuf...)
	return frame, nil
}

func (c *Client) Send(r Request) error {
	rawRequest, err := packRequest(r)
	if err != nil {
		return err
	}

	frame := make([]byte, len(rawRequest)+1)
	frame[0] = bcpSof
	copy(frame[1:], rawRequest)

	n, err := c.port.Write(frame)
	if err != nil {
		return err
	}
	if n != len(frame) {
		return errors.New("request has been sent incomplete")
	}

	return nil
}

func (c *Client) Recv() (Response, error) {
	frame, err := readFrame(c.reader)
	if err != nil {
		return Response{}, err
	}

	r, err := unpackResponse(frame)
	if err != nil {
		return Response{}, err
	}

	if err = mapBcpStatus(r.Status); err != nil {
		return Response{}, err
	}

	return r, nil
}

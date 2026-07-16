package bcp

import "stm32-bootctl/pkg/crc"

type BcpCommand uint8

const (
	FlashCommand   BcpCommand = BcpCommand(0x01)
	VerifyCommand             = BcpCommand(0x02)
	RunCommand                = BcpCommand(0x03)
	VersionCommand            = BcpCommand(0x04)
)

const (
	bcpOk             = uint8(0x00)
	bcpUnknownCommand = uint8(0x01)
	bcpInvalidData    = uint8(0x02)
	bcpBadCrc         = uint8(0x03)
	bcpInvalidSlot    = uint8(0x04)
	bcpInternalError  = uint8(0x05)
)

const (
	bcpSof                = (0xAA)
	bcpRequestHeaderSize  = (0x02)
	bcpResponseHeaderSize = (0x03)
	bcpMaxDataLength      = (255)
	bcpResponseMaxSize    = (1 + bcpResponseHeaderSize + bcpMaxDataLength + 2)
	bcpResponseMinSize    = (1 + bcpResponseHeaderSize + 2)
	bcpRequestMaxSize     = (1 + bcpRequestHeaderSize + bcpMaxDataLength + 2)
)

type Response struct {
	Command BcpCommand
	Status  uint8
	Data    []byte
}

type Request struct {
	Command BcpCommand
	Data    []byte
}

func isCommandValid(cmd BcpCommand) bool {
	switch cmd {
	case FlashCommand, RunCommand, VerifyCommand, VersionCommand:
		return true
	}
	return false
}

func NewRequest(cmd BcpCommand, data []byte) (Request, error) {
	r := Request{}
	if err := r.SetCommand(cmd); err != nil {
		return Request{}, err
	}
	if err := r.SetData(data); err != nil {
		return Request{}, err
	}
	return r, nil
}

func (r *Request) SetCommand(cmd BcpCommand) error {
	if !isCommandValid(cmd) {
		return ErrUnknownCommand
	}
	r.Command = cmd
	return nil
}

func (r *Request) SetData(data []byte) error {
	if data == nil {
		return ErrDataIsNil
	}
	if len(data) > int(bcpMaxDataLength) {
		return ErrDataIsTooLong
	}
	r.Data = make([]byte, len(data))
	copy(r.Data, data)
	return nil
}

func packRequest(r Request) ([]byte, error) {
	arr := make([]byte, bcpRequestHeaderSize+len(r.Data)+2)

	arr[0] = uint8(r.Command)
	if len(r.Data) > bcpMaxDataLength {
		return nil, ErrDataIsTooLong
	}
	arr[1] = uint8(len(r.Data))
	copy(arr[2:], r.Data)

	crc, err := crc.Crc16Modbus(arr[:bcpRequestHeaderSize+len(r.Data)])
	if err != nil {
		return nil, err
	}

	arr[bcpRequestHeaderSize+len(r.Data)] = uint8(crc >> 8)
	arr[bcpRequestHeaderSize+len(r.Data)+1] = uint8(crc)

	return arr, nil
}

func packResponse(r Response) ([]byte, error) {
	return []byte{}, nil
}

func unpackRequest(frame []byte) (Request, error) {
	return Request{}, nil
}

func unpackResponse(frame []byte) (Response, error) {
	if (len(frame) > bcpResponseMaxSize) || (len(frame) < bcpResponseMinSize) {
		return Response{}, ErrResponseBadLength
	}

	var r Response
	r.Command = BcpCommand(frame[1])
	r.Status = frame[2]

	dataLength := int(frame[3])
	if len(frame) != bcpResponseMinSize+dataLength {
		return Response{}, ErrResponseBadLength
	}

	r.Data = make([]byte, dataLength)
	copy(r.Data, frame[4:4+dataLength])

	expectedCrc := uint16(frame[4+dataLength])<<8 | uint16(frame[5+dataLength])
	realCrc, err := crc.Crc16Modbus(frame[1 : 4+dataLength])
	if err != nil {
		return Response{}, err
	}

	if expectedCrc != realCrc {
		return Response{}, ErrBadCrc
	}

	return r, nil
}

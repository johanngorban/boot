package bcp

import "errors"

var (
	ErrUnknownCommand = errors.New("unknown BCP command")
	ErrInvalidData    = errors.New("invalid data")
	ErrBadCrc         = errors.New("bad CRC")
	ErrInvalidSlot    = errors.New("invalid slot")
	ErrInternalError  = errors.New("bootloader internal error")
	ErrUnknwonError   = errors.New("unknown error")
)

var (
	ErrDataIsNil         = errors.New("data is nil")
	ErrDataIsTooLong     = errors.New("data is more than 255 bytes length")
	ErrResponseBadLength = errors.New("response is not enough or too long")
)

func mapBcpStatus(status uint8) error {
	switch status {
	case bcpOk:
		return nil
	case bcpUnknownCommand:
		return ErrUnknownCommand
	case bcpInvalidData:
		return ErrInvalidData
	case bcpBadCrc:
		return ErrBadCrc
	case bcpInvalidSlot:
		return ErrInvalidSlot
	case bcpInternalError:
		return ErrInvalidSlot
	default:
		return ErrUnknwonError
	}
}

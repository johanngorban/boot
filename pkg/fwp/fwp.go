package fwp

const (
	fwpSof        = 0xAA
	fwpTypeStart  = 0x01
	fwpTypeData   = 0x02
	fwpTypeEnd    = 0x03
	fwpAck        = 0x06
	fwpNak        = 0x15
	fwpDataSize   = 256
	fwpHeaderSize = 5
)

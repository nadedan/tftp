package tftp

import (
	"encoding/binary"
	"fmt"
)

type (
	mode     string
	opCode   uint16
	errCode  uint16
	blockNum uint16
	option   string
)

const (
	tftpPortInit  = 69
	maxRxSize     = 1024
	opCodeBytes   = 2
	errCodeBytes  = 2
	blockNumBytes = 2
)

const (
	opRRQ   opCode = 1 // Read Request
	opWRQ   opCode = 2 // Write Request
	opDATA  opCode = 3 // Data Packet
	opACK   opCode = 4 // Acknowledgment
	opERROR opCode = 5 // Error Packet
	opOACK  opCode = 6 // Option Acknowledgment (RFC 2347)
)

const (
	errUndefined       errCode = 0 // Not defined, see error message
	errFileNotFound    errCode = 1 // File not found
	errAccessViolation errCode = 2 // Access violation
	errDiskFull        errCode = 3 // Disk full or allocation exceeded
	errIllegalOp       errCode = 4 // Illegal TFTP operation
	errUnknownTID      errCode = 5 // Unknown transfer ID
	errFileExists      errCode = 6 // File already exists
	errNoSuchUser      errCode = 7 // No such user
)

func (m mode) bytes() []byte {
	return []byte(string(m))
}

const (
	optionBlockSize option = "blksize"
	optionTimeout   option = "timeout"
)

func (o option) bytes() []byte {
	return []byte(string(o))
}

type optVal interface {
	optValBytes() []byte
}

type (
	optValBlocksize uint16
	optValTimeout   uint8
)

const blockSizeDefault optValBlocksize = 512

func (o optValBlocksize) optValBytes() []byte {
	asciiBlksize := fmt.Sprintf("%d", o)
	b := []byte(asciiBlksize)
	return b
}

func (o optValTimeout) optValBytes() []byte {
	return []byte{byte(o)}
}

type tftpUint16 interface {
	opCode | errCode | blockNum | optValBlocksize
}

func twoBytes[T tftpUint16](v T) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	return b
}

func putTwoBytes[T tftpUint16](b []byte, v T) {
	if len(b) != 2 {
		panic(fmt.Sprintf("putTwoBytes requires a buffer of 2 bytes, but it got %d", len(b)))
	}
	binary.BigEndian.PutUint16(b, uint16(v))
}

func fromTwoBytes[T tftpUint16](b []byte) T {
	if len(b) < 2 {
		panic(fmt.Sprintf("fromTwoBytes requires at least 2 bytes, but it got %d", len(b)))
	}
	return T(binary.BigEndian.Uint16(b))
}

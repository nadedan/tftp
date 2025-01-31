package tftp

import (
	"encoding/binary"
	"fmt"
)

type (
	OpCode   uint16
	ErrCode  uint16
	BlockNum uint16
	Mode     string
	Option   string
)

const (
	OpCodeBytes   = 2
	ErrCodeBytes  = 2
	BlockNumBytes = 2
)

const (
	OpRRQ   OpCode = 1 // Read Request
	OpWRQ   OpCode = 2 // Write Request
	OpDATA  OpCode = 3 // Data Packet
	OpACK   OpCode = 4 // Acknowledgment
	OpERROR OpCode = 5 // Error Packet
	OpOACK  OpCode = 6 // Option Acknowledgment (RFC 2347)
)

const (
	ErrUndefined       ErrCode = 0 // Not defined, see error message
	ErrFileNotFound    ErrCode = 1 // File not found
	ErrAccessViolation ErrCode = 2 // Access violation
	ErrDiskFull        ErrCode = 3 // Disk full or allocation exceeded
	ErrIllegalOp       ErrCode = 4 // Illegal TFTP operation
	ErrUnknownTID      ErrCode = 5 // Unknown transfer ID
	ErrFileExists      ErrCode = 6 // File already exists
	ErrNoSuchUser      ErrCode = 7 // No such user
)

const (
	ModeNetascii Mode = "netascii" // ASCII text mode
	ModeOctet    Mode = "octet"    // Raw binary mode
)

func (m Mode) bytes() []byte {
	return []byte(string(m))
}

const (
	OptionBlockSize Option = "blocksize"
	OptionTimeout   Option = "timeout"
)

func (o Option) bytes() []byte {
	return []byte(string(o))
}

type OptVal interface {
	optValBytes() []byte
}

type (
	OptValBlocksize uint16
	OptValTimeout   uint8
)

func (o OptValBlocksize) optValBytes() []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(o))
	return b
}

func (o OptValTimeout) optValBytes() []byte {
	return []byte{byte(o)}
}

type tftpUint16 interface {
	OpCode | ErrCode | BlockNum | OptValBlocksize
}

func twoBytes[T tftpUint16](v T) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(v))
	return b
}

func putTwoBytes[T tftpUint16](b []byte, v T) {
	if len(b) != 2 {
		panic(fmt.Sprintf("putTwoBytes requires a buffer of 2 bytes, but it got %d", len(b)))
	}
	binary.LittleEndian.PutUint16(b, uint16(v))
}

func fromTwoBytes[T tftpUint16](b []byte) T {
	if len(b) < 2 {
		panic(fmt.Sprintf("fromTwoBytes requires at least 2 bytes, but it got %d", len(b)))
	}
	return T(binary.LittleEndian.Uint16(b))
}

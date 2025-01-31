package tftp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// ReadAck processes to full received buffer with the ACK OpCode and returns the ACKed block num
func ReadAck(b []byte) (BlockNum, error) {
	if len(b) != OpCodeBytes+BlockNumBytes {
		return 0, fmt.Errorf("ReadAck: ACKs must have %d bytes, got %d", OpCodeBytes+BlockNumBytes, len(b))
	}
	opCode := fromTwoBytes[OpCode](b[0:2])
	if opCode != OpACK {
		return 0, fmt.Errorf("ReadAck: invalid OpCode %s", opCode)
	}

	return fromTwoBytes[BlockNum](b[2:4]), nil
}

// ReadOack can read OACK packets and ACK packets. If it gets an OACK packet, it parses and returns the acknowledged options.
// If it gets an ACK packet, it just returns a nil options map.
func ReadOack(b []byte) (map[Option]OptVal, error) {
	if len(b) < OpCodeBytes+BlockNumBytes {
		return nil, fmt.Errorf("ReadOack: must have at least %d bytes, got %d", OpCodeBytes+BlockNumBytes, len(b))
	}
	opCode := fromTwoBytes[OpCode](b[0:2])
	if opCode == OpACK {
		return nil, nil
	}
	if opCode != OpOACK {
		return nil, fmt.Errorf("ReadOack: invalid OpCode %s", opCode)
	}

	opts := make(map[Option]OptVal)
	rdr := bufio.NewReader(bytes.NewReader(b[2:]))
READLOOP:
	for {
		option, err := rdr.ReadString(0x00)
		if err != nil {
			if err == io.EOF {
				break READLOOP
			}
			return nil, fmt.Errorf("ReadOack: could not read option string from buffer: %w", err)
		}
		switch Option(option) {
		case OptionBlockSize:
			val, err := rdr.ReadBytes(0x00)
			if err != nil {
				if err == io.EOF {
					break READLOOP
				}
				return nil, fmt.Errorf("ReadOack: could not read blocksize from buffer: %w", err)
			}
			opts[OptionBlockSize] = fromTwoBytes[OptValBlocksize](val)
		case OptionTimeout:
			val, err := rdr.ReadBytes(0x00)
			if err != nil {
				if err == io.EOF {
					break READLOOP
				}
				return nil, fmt.Errorf("ReadOack: could not read timeout from buffer: %w", err)
			}
			t := val[0]
			opts[OptionTimeout] = OptValTimeout(t)
		default:
			return nil, fmt.Errorf("ReadOack: unsupported option %s", option)
		}
	}

	return opts, nil

}

func ReadError(b []byte) (ErrCode, string, error) {
	if len(b) < OpCodeBytes+ErrCodeBytes {
		return 0, "", fmt.Errorf("ReadError: need at least %d byte, got %d", OpCodeBytes+ErrCodeBytes, len(b))
	}

	opCode := fromTwoBytes[OpCode](b[0:2])
	if opCode != OpERROR {
		return 0, "", fmt.Errorf("ReadError: invalid OpCode %s", opCode)
	}

	errCode := fromTwoBytes[ErrCode](b[2:4])
	errString, err := bufio.NewReader(bytes.NewReader(b[4:])).ReadString(0x00)
	if err != nil {
		if err == io.EOF {
			return errCode, "", nil
		}

		return 0, "", fmt.Errorf("ReadError: could not read error message from buffer: %w", err)
	}

	return errCode, errString, nil
}

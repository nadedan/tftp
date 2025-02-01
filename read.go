package tftp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// readAck processes to full received buffer with the ACK OpCode and returns the ACKed block num
func readAck(b []byte) (blockNum, error) {
	if len(b) != opCodeBytes+blockNumBytes {
		return 0, fmt.Errorf("ReadAck: ACKs must have %d bytes, got %d", opCodeBytes+blockNumBytes, len(b))
	}
	opCode := fromTwoBytes[opCode](b[0:2])
	if opCode != opACK {
		return 0, fmt.Errorf("ReadAck: invalid OpCode %s", opCode)
	}

	return fromTwoBytes[blockNum](b[2:4]), nil
}

// readOack can read OACK packets and ACK packets. If it gets an OACK packet, it parses and returns the acknowledged options.
// If it gets an ACK packet, it just returns a nil options map.
func readOack(b []byte) (map[option]optVal, error) {
	opts := map[option]optVal{
		optionBlockSize: blockSizeDefault,
		optionTimeout:   optValTimeout(5),
	}
	if len(b) < opCodeBytes+blockNumBytes {
		return nil, fmt.Errorf("readOack: must have at least %d bytes, got %d", opCodeBytes+blockNumBytes, len(b))
	}
	opCode := fromTwoBytes[opCode](b[0:2])
	if opCode == opACK {
		return opts, nil
	}
	if opCode != opOACK {
		return nil, fmt.Errorf("readOack: invalid OpCode %s", opCode)
	}

	rdr := bufio.NewReader(bytes.NewReader(b[2:]))
READLOOP:
	for {
		opt, err := rdr.ReadString(0x00)
		if err != nil {
			if err == io.EOF {
				break READLOOP
			}
			return nil, fmt.Errorf("readOack: could not read option string from buffer: %w", err)
		}
		// string the null terminator from option string
		opt = opt[:len(opt)-1]
		switch option(opt) {
		case optionBlockSize:
			val, err := rdr.ReadString(0x00)
			if err != nil {
				if err == io.EOF {
					break READLOOP
				}
				return nil, fmt.Errorf("readOack: could not read blocksize from buffer: %w", err)
			}
			val = val[:len(val)-1]
			blksize, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("readOack: bad blksize: %s", val)
			}
			opts[optionBlockSize] = optValBlocksize(blksize)
		case optionTimeout:
			val, err := rdr.ReadBytes(0x00)
			if err != nil {
				if err == io.EOF {
					break READLOOP
				}
				return nil, fmt.Errorf("readOack: could not read timeout from buffer: %w", err)
			}
			t := val[0]
			opts[optionTimeout] = optValTimeout(t)
		default:
			return nil, fmt.Errorf("readOack: unsupported option %s", opt)
		}
	}

	return opts, nil

}

func readError(b []byte) (errCode, string, error) {
	if len(b) < opCodeBytes+errCodeBytes {
		return 0, "", fmt.Errorf("ReadError: need at least %d byte, got %d", opCodeBytes+errCodeBytes, len(b))
	}

	opCode := fromTwoBytes[opCode](b[0:2])
	if opCode != opERROR {
		return 0, "", fmt.Errorf("ReadError: invalid OpCode %s", opCode)
	}

	errCode := fromTwoBytes[errCode](b[2:4])
	errString, err := bufio.NewReader(bytes.NewReader(b[4:])).ReadString(0x00)
	if err != nil {
		if err == io.EOF {
			return errCode, "", nil
		}

		return 0, "", fmt.Errorf("ReadError: could not read error message from buffer: %w", err)
	}

	return errCode, errString, nil
}

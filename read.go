package tftp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// readAck processes to full received buffer with the ACK OpCode and returns the ACKed block num
func readAck(b []byte) (BlockNum, error) {
	if len(b) != OpCodeBytes+BlockNumBytes {
		return 0, fmt.Errorf("ReadAck: ACKs must have %d bytes, got %d", OpCodeBytes+BlockNumBytes, len(b))
	}
	opCode := fromTwoBytes[OpCode](b[0:2])
	if opCode != OpACK {
		return 0, fmt.Errorf("ReadAck: invalid OpCode %s", opCode)
	}

	return fromTwoBytes[BlockNum](b[2:4]), nil
}

// readOack can read OACK packets and ACK packets. If it gets an OACK packet, it parses and returns the acknowledged options.
// If it gets an ACK packet, it just returns a nil options map.
func readOack(b []byte) (map[Option]optVal, error) {
	opts := map[Option]optVal{
		OptionBlockSize: BlockSizeDefault,
		OptionTimeout:   optValTimeout(5),
	}
	if len(b) < OpCodeBytes+BlockNumBytes {
		return nil, fmt.Errorf("readOack: must have at least %d bytes, got %d", OpCodeBytes+BlockNumBytes, len(b))
	}
	opCode := fromTwoBytes[OpCode](b[0:2])
	if opCode == OpACK {
		return opts, nil
	}
	if opCode != OpOACK {
		return nil, fmt.Errorf("readOack: invalid OpCode %s", opCode)
	}

	rdr := bufio.NewReader(bytes.NewReader(b[2:]))
READLOOP:
	for {
		option, err := rdr.ReadString(0x00)
		if err != nil {
			if err == io.EOF {
				break READLOOP
			}
			return nil, fmt.Errorf("readOack: could not read option string from buffer: %w", err)
		}
		// string the null terminator from option string
		option = option[:len(option)-1]
		switch Option(option) {
		case OptionBlockSize:
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
			opts[OptionBlockSize] = optValBlocksize(blksize)
		case OptionTimeout:
			val, err := rdr.ReadBytes(0x00)
			if err != nil {
				if err == io.EOF {
					break READLOOP
				}
				return nil, fmt.Errorf("readOack: could not read timeout from buffer: %w", err)
			}
			t := val[0]
			opts[OptionTimeout] = optValTimeout(t)
		default:
			return nil, fmt.Errorf("readOack: unsupported option %s", option)
		}
	}

	return opts, nil

}

func readError(b []byte) (ErrCode, string, error) {
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

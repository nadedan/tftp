package tftp

// ackPacket returns an ACK packet as a byte slice
//
// +--------------------------------------+
// | Opcode (2 bytes) | Block # (2 bytes) |
// +--------------------------------------+
func ackPacket(blk blockNum) []byte {
	b := make([]byte, opCodeBytes+blockNumBytes)
	putTwoBytes(b[0:2], opACK)
	putTwoBytes(b[2:4], blk)
	return b
}

// dataPacket returns a DATA packet as a byte slice
//
// +---------------------------------------------------+
// | Opcode (2 bytes) | Block # (2 bytes) | dataPacket Bytes |
// +---------------------------------------------------+
func dataPacket(blkNum blockNum, data []byte) []byte {
	b := make([]byte, opCodeBytes+blockNumBytes+len(data))
	putTwoBytes(b[0:2], opDATA)
	putTwoBytes(b[2:4], blkNum)
	copy(b[4:], data)
	return b
}

// wrqPacket returns a WRQ (Write ReQuest) packet as a byte slice
//
// The mode of the request is required. The opts are optional and a nil map can be given
//
// +------------------------------------------------------------------------------------------------------------+
// | Opcode (2 bytes) | Filename (ascii bytes) | 0x00 | Mode (ascii bytes) | 0x00 | null terminated option list |
// +------------------------------------------------------------------------------------------------------------+
func wrqPacket(filename string, mde mode, opts map[option]optVal) []byte {
	b := make([]byte, 0)
	b = append(b, twoBytes(opWRQ)...)
	b = append(b, []byte(filename)...)
	b = append(b, 0x00)
	b = append(b, mde.bytes()...)
	b = append(b, 0x00)
	for option, value := range opts {
		b = append(b, option.bytes()...)
		b = append(b, 0x00)
		b = append(b, value.optValBytes()...)
		b = append(b, 0x00)
	}

	return b
}

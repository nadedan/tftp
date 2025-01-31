package tftp

// Ack returns an ACK packet as a byte slice
//
// +--------------------------------------+
// | Opcode (2 bytes) | Block # (2 bytes) |
// +--------------------------------------+
func Ack(blockNum BlockNum) []byte {
	b := make([]byte, OpCodeBytes+BlockNumBytes)
	putTwoBytes(b[0:2], OpACK)
	putTwoBytes(b[2:4], blockNum)
	return b
}

// Data returns a DATA packet as a byte slice
//
// +---------------------------------------------------+
// | Opcode (2 bytes) | Block # (2 bytes) | Data Bytes |
// +---------------------------------------------------+
func Data(blockNum BlockNum, data []byte) []byte {
	b := make([]byte, OpCodeBytes+BlockNumBytes+len(data))
	putTwoBytes(b[0:2], OpDATA)
	putTwoBytes(b[2:4], blockNum)
	copy(b[4:], data)
	return b
}

// Wrq returns a WRQ (Write ReQuest) packet as a byte slice
//
// The mode of the request is required. The opts are optional and a nil map can be given
//
// +------------------------------------------------------------------------------------------------------------+
// | Opcode (2 bytes) | Filename (ascii bytes) | 0x00 | Mode (ascii bytes) | 0x00 | null terminated option list |
// +------------------------------------------------------------------------------------------------------------+
func Wrq(filename string, mode Mode, opts map[Option]OptVal) []byte {
	b := make([]byte, 0)
	b = append(b, twoBytes(OpWRQ)...)
	b = append(b, []byte(filename)...)
	b = append(b, 0x00)
	b = append(b, mode.bytes()...)
	b = append(b, 0x00)
	for option, value := range opts {
		b = append(b, option.bytes()...)
		b = append(b, 0x00)
		b = append(b, value.optValBytes()...)
		b = append(b, 0x00)
	}

	return b
}

package lc3

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

/*
Reminder:
	The most significant bit: Left most bit
	The least significant bit: Right most bit

	Left shift (<<) : the most-significant bit is lost, and a 0 bit is inserted on the other end
		Examples:	- 0010 << 1 -> 0100
					- 0010 << 2 -> 1000
	Logical right shift (>>>) : the least-significant bit is lost and a 0 is inserted on the other end
		Examples:	- 1011 >>> 1 -> 0101
					- 1011 >>> 3 -> 0001
	Arithmetic right shift (>>) : the least-significant bit is lost and the most-significant bit is copied
		Examples:	- 1011 >> 1 -> 1101
					- 1011 >> 3 -> 1111
	Bitwise OR : 1 | 1 -> 1, 1 | 0 -> 1, 0 | 1 -> 1, 0 | 0 -> 0
	Bitwise AND : 1 & 1 -> 1, 1 & 0 -> 0, 0 & 1 -> 0, 0 & 0 -> 0
	Bitwise NOT : ~1 -> 0, ~0 -> 1

	Hex:	0x1   -> 1
			0x3   -> 11
			0x7   -> 111
			0xF   -> 1111
			0x1F  -> 11111
			0x3F  -> 111111
			0xFF  -> 11111111
			0x1FF -> 111111111
			0x7FF -> 11111111111
*/

// SignExtend converts a 5 bitCount integer to a 16 bits number (preserving sign)
// As an example we can take the number 13, in binary it is: 01101
// it will perform a bitCount-1 right shift, such as x >> 4 <=> 01101 >> 4
// <=> 00110 >> 3 <=> 00011 >> 2 <=> 00001 >> 1 <=> 00000
// 00000 & 00001 => 0 -> we return 01101
// Example for -13: 10011
// <=> 10011 >> 4 <=> 11001 >> 3 <=> 11100 >> 2 <=> 11110 >> 1 <=> 11111
// 11111 & 1 => 1 -> we apply and bitwise OR 0xFFFF << 5
// <=> 1111111111111111 << 5 <=> 1111111111100000
// 000000000010011 | 1111111111100000 => 1111111111110011 (-13 in 16 bits)
func SignExtend(x uint16, bitCount int) uint16 {
	if ((x >> (bitCount - 1)) & 1) == 1 {
		x |= ^uint16(0) << bitCount
	}
	return x
}

func ReadImageFile() {

}

// ReadImage reads an LC-3 object file (big-endian 16-bit words) into memory.
func ReadImage(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// read origin (big-endian)
	var origin uint16
	if err = binary.Read(f, binary.BigEndian, &origin); err != nil {
		return err
	}
	addr := int(origin)

	// read remaining 16-bit words (big-endian) into memory starting at origin
	buf := make([]byte, 2)
	for {
		_, err = io.ReadFull(f, buf)
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if err != nil {
			return err
		}
		memory[addr] = binary.BigEndian.Uint16(buf)
		addr++
	}
	return nil
}

package main

import (
	"bytes"
	"encoding/binary"
	"os"
)

var HostEndian binary.ByteOrder

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
	if ((x >> (bitCount - 1)) & 1) != 0 {
		x |= 0xFFFF << bitCount
	}
	return x
}

// Swap16 LC-3 programs are big-endian (modern computers are little-endian)
// therefore we should swap each uint16 that is loaded
func Swap16(x uint16) uint16 {
	if HostEndian != binary.BigEndian {
		return x
	}
	return (x << 8) | (x >> 8)
}

func ReadImageFile() {

}

func ReadImage(imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	size := info.Size()
	data := make([]byte, size)
	_, err = file.Read(data)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(data)
	// origin tells us where in memory to place the image
	origin := binary.BigEndian.Uint16(buffer.Next(2))

	for i := 0; i < buffer.Len(); i++ {
		b := buffer.Next(2)
		if len(b) == 0 {
			break
		}
		memory[origin] = binary.BigEndian.Uint16(b)
		origin++
	}

	return nil
}

// HostEndian is the byte order of the host computer
func init() {
	// []byte{0x12, 0x34} : create a two-byte slice representing a 16-bit value.
	// binary.BigEndian.Uint16() : convert the two-byte slice to an uint16 value using big-endian byte order.
	// == 0x1234 : compare the converted value with the expected value 0x1234.
	isBigEndian := binary.BigEndian.Uint16([]byte{0x12, 0x34}) == 0x1234

	if isBigEndian {
		HostEndian = binary.BigEndian
	} else {
		HostEndian = binary.LittleEndian
	}

}

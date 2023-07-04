package main

// MemoryWrite writes val to address
func MemoryWrite(address, val uint16) {
	memory[address] = val
}

// MemoryRead reads uint16 at address
func MemoryRead(address uint16) uint16 {

	/**
	Memory mapped registers make memory access a bit more complicated.
	We can’t read and write to the memory array directly, but must instead call setter and getter functions.
	When memory is read from KBSR, the getter will check the keyboard and update both memory locations.
	*/
	if address == MR_KBSR {

	}

	return memory[address]
}

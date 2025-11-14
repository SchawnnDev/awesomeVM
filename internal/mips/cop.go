package mips

type Coprocessor interface {
	// LoadWord loads a word (4 bytes) from the given address.
	// Instruction: COP_LW (z, rt, memword)
	// z: The coprocessor unit number.
	// rt: Coprocessor general register specifier.
	// memword: A 32-bit word value supplied to the coprocessor.
	LoadWord(addr uint32) uint32

	// StoreWord stores a word (4 bytes) to the given address.
	// Instruction: COP_SW (z, rt, dataword)
	// z: The coprocessor unit number.
	// rt: Coprocessor general register specifier.
	// dataword: 32-bit word value.
	StoreWord(addr uint32, value uint32)

	// LoadDoubleWord loads a double word (8 bytes) from the given address.
	// Instruction: COP_LD (z, rt, memdouble)
	// z: The coprocessor unit number.
	// rt: Coprocessor general register specifier.
	// memdouble: 64-bit doubleword value supplied to the coprocessor.
	LoadDoubleWord(addr uint32) uint64

	// StoreDoubleWord stores a double word (8 bytes) to the given address.
	// Instruction: COP_SD (z, rt, datadouble)
	// z: The coprocessor unit number.
	// rt: Coprocessor general register specifier.
	// datadouble: 64-bit doubleword value.
	StoreDoubleWord(addr uint32, value uint64)
}

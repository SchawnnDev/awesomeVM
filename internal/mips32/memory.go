package mips32

type Memory struct {
	Data []byte
}

func NewMemory(size uint32) *Memory {
	return &Memory{
		Data: make([]byte, size),
	}
}

func (m *Memory) LoadWord(address uint32) (word uint32, ok bool) {
	if !m.isAligned(address) || !m.isAddressInRange(address+3) {
		return 0, false
	}

	return uint32(m.Data[address])<<24 |
		uint32(m.Data[address+1])<<16 |
		uint32(m.Data[address+2])<<8 |
		uint32(m.Data[address+3]), true
}

func (m *Memory) StoreWord(address uint32, value uint32) (ok bool) {
	if !m.isAligned(address) || !m.isAddressInRange(address+3) {
		return false
	}

	m.Data[address] = byte(value >> 24)
	m.Data[address+1] = byte(value >> 16)
	m.Data[address+2] = byte(value >> 8)
	m.Data[address+3] = byte(value)
	return true
}

// isAligned checks if the address is word-aligned (multiple of 4)
func (m *Memory) isAligned(address uint32) bool {
	return address%4 == 0
}

// isAddressInRange checks if the address is within the memory bounds
func (m *Memory) isAddressInRange(address uint32) bool {
	return address+3 < uint32(len(m.Data))
}

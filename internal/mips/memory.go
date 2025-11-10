package mips

type Memory struct {
	Data []byte
}

func NewMemory(size uint32) *Memory {
	return &Memory{
		Data: make([]byte, size),
	}
}

func (m *Memory) LoadWord(address uint32) uint32 {
	return uint32(m.Data[address])<<24 |
		uint32(m.Data[address+1])<<16 |
		uint32(m.Data[address+2])<<8 |
		uint32(m.Data[address+3])
}

func (m *Memory) StoreWord(address uint32, value uint32) {
	m.Data[address] = byte(value >> 24)
	m.Data[address+1] = byte(value >> 16)
	m.Data[address+2] = byte(value >> 8)
	m.Data[address+3] = byte(value)
}

package mips

// instructions here
type Instruction interface {
	Execute(cpu *CPU)
	Decode(instr uint32) Instruction
}

// DecodeInstruction
// we want to check the type of the instruction
// there are 3 types, R, I and J:

// R is opcode = 0
// I is opcode != 0, excluding 2 and 3, since:
// J is opcode = 2 or 3
func DecodeInstruction(instr uint32) Instruction {
	opcode := (instr >> 26) & 0x3F
	var result Instruction

	if opcode == 0x0 {
		result = RTypeInstruction{}
	} else if opcode == 0x2 || opcode == 0x3 {
		result = JTypeInstruction{}
	} else {
		result = ITypeInstruction{}
	}

	return result.Decode(instr)
}

type RTypeInstruction struct {
	Opcode uint8 // 6 bits
	Rs     uint8 // 5 bits
	Rt     uint8 // 5 bits
	Rd     uint8 // 5 bits
	Shamt  uint8 // 5 bits
	Funct  uint8 // 6 bits
}

func (ri RTypeInstruction) Decode(instr uint32) Instruction {
	return &RTypeInstruction{
		Opcode: uint8((instr >> 26) & 0x3F),
		Rs:     uint8((instr >> 21) & 0x1F),
		Rt:     uint8((instr >> 16) & 0x1F),
		Rd:     uint8((instr >> 11) & 0x1F),
		Shamt:  uint8((instr >> 6) & 0x1F),
		Funct:  uint8(instr & 0x3F),
	}
}

func (ri RTypeInstruction) Execute(cpu *CPU) {}

type ITypeInstruction struct {
	Opcode    uint8  // 6 bits
	Rs        uint8  // 5 bits
	Rt        uint8  // 5 bits
	Immediate uint16 // 16 bits
}

func (ii ITypeInstruction) Decode(instr uint32) Instruction {
	return &ITypeInstruction{
		Opcode:    uint8((instr >> 26) & 0x3F),
		Rs:        uint8((instr >> 21) & 0x1F),
		Rt:        uint8((instr >> 16) & 0x1F),
		Immediate: uint16(instr & 0xFFFF),
	}
}

func (ii ITypeInstruction) Execute(cpu *CPU) {}

type JTypeInstruction struct {
	Opcode uint8  // 6 bits
	Addr   uint32 // 26 bits
}

func (ji JTypeInstruction) Decode(instr uint32) Instruction {
	return &JTypeInstruction{
		Opcode: uint8((instr >> 26) & 0x3F),
		Addr:   instr & 0x3FFFFFF,
	}
}

func (ji JTypeInstruction) Execute(cpu *CPU) {}

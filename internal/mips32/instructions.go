package mips32

import "awesomeVM/internal/utils"

type OpCode uint8

// hex <-> binary
// 0 0000
// 1 0001
// 2 0010
// 3 0011
// 4 0100
// 5 0101
// 6 0110
// 7 0111
// 8 1000
// 9 1001
// A 1010
// B 1011
// C 1100
// D 1101
// E 1110
// F 1111

const (
	// R-Type funct codes
	OpCodeADD  OpCode = 0x20
	OpCodeADDU OpCode = 0x28
	OpCodeAND  OpCode = 0x24
	// OpCodeDADD   OpCode = 0x2C MIPS64
	// OpCodeDADDU  OpCode = 0x2D MIPS64
	//OpCodeDDIV   OpCode = 0x1E MIPS64
	// OpCodeDDIVU  OpCode = 0x1F MIPS64
	OpCodeDIV  OpCode = 0x1A
	OpCodeDIVU OpCode = 0x1B
	//OpCodeDMULT  OpCode = 0x1C MIPS64
	//OpCodeDMULTU OpCode = 0x1D MIPS64
	//OpCodeDSLL   OpCode = 0x38 MIPS64
	//OpCodeDSLL32 OpCode = 0x3C MIPS64
	//OpCodeDSLLV  OpCode = 0x14 MIPS64
	//OpCodeDSRA   OpCode = 0x3B MIPS64
	//OpCodeDSRA32 OpCode = 0x3F MIPS64
	//OpCodeDSRAV  OpCode = 0x17 MIPS64
	//OpCodeDSRL   OpCode = 0x3A MIPS64
	//OpCodeDSRL32 OpCode = 0x3E MIPS64
	//OpCodeDSRLV  OpCode = 0x16 MIPS64
	//OpCodeDSUB   OpCode = 0x2E MIPS64
	//OpCodeDSUBU  OpCode = 0x2F MIPS64
	OpCodeJALR  OpCode = 0x09
	OpCodeJR    OpCode = 0x08
	OpCodeMFHI  OpCode = 0x10
	OpCodeMFLO  OpCode = 0x12
	OpCodeMOVN  OpCode = 0x0B
	OpCodeMOVZ  OpCode = 0x0A
	OpCodeMTHI  OpCode = 0x11
	OpCodeMTLO  OpCode = 0x13
	OpCodeMULT  OpCode = 0x18
	OpCodeMULTU OpCode = 0x19
	OpCodeNOR   OpCode = 0x27
	OpCodeOR    OpCode = 0x25
	OpCodeSLL   OpCode = 0x00
	OpCodeSLLV  OpCode = 0x04
	OpCodeSLT   OpCode = 0x2A
	OpCodeSLTU  OpCode = 0x2B
	OpCodeSRA   OpCode = 0x03
	OpCodeSRAV  OpCode = 0x07
	OpCodeSRL   OpCode = 0x02
	OpCodeSRLV  OpCode = 0x06
	OpCodeSUB   OpCode = 0x22
	OpCodeSUBU  OpCode = 0x23
	OpCodeTEQ   OpCode = 0x34
	OpCodeTGE   OpCode = 0x30
	OpCodeTGEU  OpCode = 0x31
	OpCodeTLT   OpCode = 0x32
	OpCodeTLTU  OpCode = 0x33
	OpCodeTNE   OpCode = 0x36
	OpCodeXOR   OpCode = 0x26

	// I-Type opcodes
	OpCodeADDI  OpCode = 0x8
	OpCodeADDIU OpCode = 0x9

	// COP0 related opcodes
	OpCodeCOP0 uint8 = 0x10
	OpCodeERET uint8 = 0x18

	// COP0 functions
	COP0Funct_MFC0  uint8 = 0x00 // Move From CP0
	COP0Funct_MTC0  uint8 = 0x04 // Move To CP0
	COP0Funct_ERET  uint8 = 0x18 // Exception Return
	COP0Funct_TLBP  uint8 = 0x08 // TLB Probe
	COP0Funct_TLBR  uint8 = 0x01 // TLB Read
	COP0Funct_TLBWI uint8 = 0x02 // TLB Write Indexed
	COP0Funct_TLBWR uint8 = 0x06 // TLB Write Random
)

// instructions here
type Instruction interface {
	// Execute executes the instruction on the given CPU.
	// It returns a pointer to the new program counter (PC) if it was changed,
	// or nil if the PC should just advance to the next instruction.
	Execute(cpu *CPU) (nextPC *uint32, delaySlot bool)
	Decode(instr uint32) Instruction
}

// DecodeInstruction
// we want to check the type of the instruction
// there are 3 types, R, I and J:

// DecodeInstruction decodes a 32-bit MIPS instruction into its corresponding Instruction type.
// R is opcode = 0
// I is opcode != 0, excluding 2 and 3, since:
// J is opcode = 2 or 3
// COP0 is opcode = 0x10 (16 in decimal)
func DecodeInstruction(instr uint32) Instruction {
	opcode := (instr >> 26) & 0x3F
	var result Instruction

	if opcode == 0x10 { // OpCodeCOP0
		// COP0 instructions
		result = &COP0Instruction{}
	} else if opcode == 0x0 {
		result = &RTypeInstruction{}
	} else if opcode == 0x2 || opcode == 0x3 {
		result = &JTypeInstruction{}
	} else {
		result = &ITypeInstruction{}
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

func (ri RTypeInstruction) Execute(cpu *CPU) (nextPC *uint32, delaySlot bool) {

	// we want to convert Funct to OpCode type
	funct := OpCode(ri.Funct)

	// Funct is the opCode
	switch funct {
	// ADD rd, rs, rt
	// -> traps on overflow
	// if (NotWordValue(GPR[rs]) or NotWordValue(GPR[rt])) then UndefinedResult() endif
	// temp ←GPR[rs] + GPR[rt]
	//if (32_bit_arithmetic_overflow) then
	//		SignalException(IntegerOverflow)
	//else
	//		GPR[rd] ←sign_extend(temp31..0)
	//endif
	case OpCodeADD:
		rsVal := int32(cpu.GetReg(ri.Rs))
		rtVal := int32(cpu.GetReg(ri.Rt))
		temp := rsVal + rtVal

		// Check for overflow
		if utils.CheckAdditionOverflow(rsVal, rtVal, temp) {
			// Overflow occurred
			vec := cpu.cp0.RaiseException(excOv, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			return nil, false
		}

		cpu.SetReg(ri.Rd, uint32(temp))

		return nil, false

	// ADDU rd, rs, rt
	// if (NotWordValue(GPR[rs]) or NotWordValue(GPR[rt])) then UndefinedResult() endif
	// temp ←GPR[rs] + GPR[rt]
	// GPR[rd]← sign_extend(temp31..0)
	case OpCodeADDU:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		temp := rsVal + rtVal
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// AND rd, rs, rt
	// GPR[rd] ← GPR[rs] and GPR[rt]
	case OpCodeAND:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		temp := rsVal & rtVal
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	//	DIV rs, rt
	//	if (NotWordValue(GPR[rs]) or NotWordValue(GPR[rt])) then UndefinedResult() endif
	//	I- 2 :, I- 1 : LO, HI ← undefined
	//	I: q ← GPR[rs]31..0 div GPR[rt]31..0
	//	LO ← sign_extend(q31..0)
	//	r ← GPR[rs]31..0 mod GPR[rt]31..0
	//	HI ← sign_extend(r31..0)
	case OpCodeDIV:
		rsVal := int32(cpu.GetReg(ri.Rs))
		rtVal := int32(cpu.GetReg(ri.Rt))

		if rtVal == 0 {
			cpu.LO = 0
			cpu.HI = 0
			// no 0 divisor
			return nil, false
		}

		cpu.LO = rsVal / rtVal // Quotient
		cpu.HI = rsVal % rtVal // Remainder
		return nil, false

	// DIVU rs, rt
	//	if (NotWordValue(GPR[rs]) or NotWordValue(GPR[rt])) then UndefinedResult() endif
	//	I- 2 :, I- 1 : LO, HI ← undefined
	//	I: q ← (0 || GPR[rs]31..0) div (0 || GPR[rt]31..0)
	//	LO ← sign_extend(q31..0)
	//	r ← (0 || GPR[rs]31..0) mod (0 || GPR[rt]31..0)
	//	HI ← sign_extend(r31..0)
	case OpCodeDIVU:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)

		if rtVal == 0 {
			cpu.LO = 0
			cpu.HI = 0
			// no 0 divisor
			return nil, false
		}

		cpu.LO = int32(rsVal / rtVal) // Quotient
		cpu.HI = int32(rsVal % rtVal) // Remainder
		return nil, false

	// JALR rs
	// JALR rd, rs
	// Move From HI: rd = HI
	// I: temp ← GPR[rs]
	// GPR[rd] ← PC + 8
	// I+ 1 :PC ← temp
	case OpCodeJALR:
		rsVal := cpu.GetReg(ri.Rs)

		// Compute the return address (PC + 8 because of the delay slot)
		retAddr := cpu.PC + 8

		// Write return address into rd (unless rd = 0, handled by SetReg)
		if ri.Rd != 0 {
			cpu.SetReg(ri.Rd, retAddr)
		}

		// Jump target
		newPC := rsVal

		// true => we enter a delay slot
		return &newPC, true

	// JR rs
	// I: temp ← GPR[rs]
	// I+ 1 :PC ← temp
	case OpCodeJR:
		rsVal := cpu.GetReg(ri.Rs)
		return &rsVal, true

	// MFHI rd
	// GPR[rd] ← HI
	case OpCodeMFHI:
		cpu.SetReg(ri.Rd, uint32(cpu.HI))
		return nil, false

	// MFLO rd
	// GPR[rd] ← LO
	case OpCodeMFLO:
		cpu.SetReg(ri.Rd, uint32(cpu.LO))
		return nil, false

	// MOVN rd, rs, rt
	// if GPR[rt] ≠ 0 then
	// 	GPR[rd] ← GPR[rs]
	// endif
	case OpCodeMOVN:
		rtVal := cpu.GetReg(ri.Rt)
		if rtVal != 0 {
			rsVal := cpu.GetReg(ri.Rs)
			cpu.SetReg(ri.Rd, rsVal)
		}
		return nil, false

	// MOVZ rd, rs, rt
	// if GPR[rt] = 0 then
	// 	GPR[rd] ← GPR[rs]
	// endif
	case OpCodeMOVZ:
		rtVal := cpu.GetReg(ri.Rt)
		if rtVal == 0 {
			rsVal := cpu.GetReg(ri.Rs)
			cpu.SetReg(ri.Rd, rsVal)
		}
		return nil, false

	// MTHI rs
	// I- 2 :, I- 1 :HI ← undefined
	// I: HI ← GPR[rs]
	case OpCodeMTHI:
		rsVal := cpu.GetReg(ri.Rs)
		cpu.HI = int32(rsVal)
		return nil, false

	// MTLO rs
	// I- 2 :, I- 1 :LO ← undefined
	// I: LO ← GPR[rs]
	case OpCodeMTLO:
		rsVal := cpu.GetReg(ri.Rs)
		cpu.LO = int32(rsVal)
		return nil, false

	// MULT rs, rt
	// if (NotWordValue(GPR[rs]) or NotWordValue(GPR[rt])) then UndefinedResult() endif
	// I- 2 :, I- 1 : LO, HI ← undefined
	// I: prod ← GPR[rs]31..0 * GPR[rt]31..0
	// LO ← sign_extend(prod31..0)
	// HI ← sign_extend(prod63..32
	// [ 63 ..................... 32 ][ 31 ...................... 0 ]
	//        HIGH 32 bits                LOW 32 bits
	case OpCodeMULT:
		rsVal := int32(cpu.GetReg(ri.Rs))
		rtVal := int32(cpu.GetReg(ri.Rt))
		prod := int64(rsVal) * int64(rtVal)
		cpu.LO = int32(prod & 0xFFFFFFFF)         // Low 32 bits
		cpu.HI = int32((prod >> 32) & 0xFFFFFFFF) // High 32 bits
		return nil, false

	// MULTU rs, rt
	// if (NotWordValue(GPR[rs]) or NotWordValue(GPR[rt])) then UndefinedResult() endif
	// I- 2 :, I- 1 : LO, HI ← undefined
	// I: prod ← (0 || GPR[rs]31..0) * (0 || GPR[rt]31..0)
	// LO ← sign_extend(prod31..0)
	// HI ← sign_extend(prod63..32
	// [ 63 ..................... 32 ][ 31 ...................... 0 ]
	//        HIGH 32 bits                LOW 32 bits
	case OpCodeMULTU:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		prod := uint64(rsVal) * uint64(rtVal)
		cpu.LO = int32(prod & 0xFFFFFFFF)         // Low 32 bits
		cpu.HI = int32((prod >> 32) & 0xFFFFFFFF) // High 32 bits
		return nil, false

	// NOR rd, rs, rt
	// GPR[rd] ← GPR[rs] nor GPR[rt]
	case OpCodeNOR:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		temp := ^(rsVal | rtVal)
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// OR rd, rs, rt
	// GPR[rd] ← GPR[rs] or GPR[rt]
	case OpCodeOR:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		temp := rsVal | rtVal
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// SLL rd, rt, sa
	// s ← sa
	// temp ← GPR[rt](31-s)..0 || 0s
	// GPR[rd]← sign_extend(temp)
	case OpCodeSLL:
		rtVal := cpu.GetReg(ri.Rt)
		s := ri.Shamt
		temp := rtVal << s
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// SLLV rd, rt, rs
	// s ← GP[rs]4..0
	// temp ← GPR[rt](31-s)..0 || 0s
	// GPR[rd]← sign_extend(temp)
	case OpCodeSLLV:
		rsVal := cpu.GetReg(ri.Rs)
		s := rsVal & 0xF
		temp := cpu.GetReg(ri.Rt) << s
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// SLT rd, rs, rt
	// if GPR[rs] < GPR[rt] then
	//	GPR[rd] ← 0GPRLEN-1 || 1
	// else
	//	GPR[rd] ← 0GPRLEN
	// endif
	case OpCodeSLT:
		rtVal := int32(cpu.GetReg(ri.Rt))
		rsVal := int32(cpu.GetReg(ri.Rs))

		if rsVal < rtVal {
			cpu.SetReg(ri.Rd, 1)
		} else {
			cpu.SetReg(ri.Rd, 0)
		}

		return nil, false

	// SLTU rd, rs, rt
	// if (0 || GPR[rs]) < (0 || GPR[rt]) then
	//	GPR[rd] ← 0GPRLEN-1 || 1
	// else
	//	GPR[rd] ← 0GPRLEN
	// endif
	case OpCodeSLTU:
		rtVal := cpu.GetReg(ri.Rt)
		rsVal := cpu.GetReg(ri.Rs)

		if rsVal < rtVal {
			cpu.SetReg(ri.Rd, 1)
		} else {
			cpu.SetReg(ri.Rd, 0)
		}

		return nil, false

	// SRA rd, rt, sa
	// if (NotWordValue(GPR[rt])) then UndefinedResult() endif
	// s ← sa
	// temp ← (GPR[rt]31)
	// s || GPR[rt]31..s
	// GPR[rd]← sign_extend(temp)
	case OpCodeSRA:
		rtVal := cpu.GetReg(ri.Rt)
		// arithmetic shift: interpret as signed 32-bit then shift
		s := uint(ri.Shamt & 0x1F)
		temp := uint32(int32(rtVal) >> s)
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// SRAV rd, rt, rs
	// if (NotWordValue(GPR[rt])) then UndefinedResult() endif
	// s ← GPR[rs]4..0
	// temp ← (GPR[rt]31)
	// s || GPR[rt]31..s
	// GPR[rd]← sign_extend(temp)
	case OpCodeSRAV:
		rsVal := cpu.GetReg(ri.Rs)
		s := uint(rsVal & 0x1F)
		temp := uint32(int32(cpu.GetReg(ri.Rt)) >> s)
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// SRL rd, rt, sa
	// if (NotWordValue(GPR[rt])) then UndefinedResult() endif
	// s ← sa
	// temp ← 0s || GPR[rt]31..s
	// GPR[rd]← sign_extend(temp)
	case OpCodeSRL:
		rtVal := cpu.GetReg(ri.Rt)
		s := uint(ri.Shamt & 0x1F)
		temp := rtVal >> s
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// SRAV rd, rt, rs
	// if (NotWordValue(GPR[rt])) then UndefinedResult() endif
	// s ← GPR[rs]4..0
	// temp ← 0s || GPR[rt]31..s
	// GPR[rd]← sign_extend(temp)
	case OpCodeSRLV:
		rsVal := cpu.GetReg(ri.Rs)
		s := rsVal & 0x1F
		temp := cpu.GetReg(ri.Rt) >> s
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// SUB rd, rs, rt
	// if (NotWordValue(GPR[rs]) or NotWordValue(GPR[rt])) then UndefinedResult() endif
	// temp ← GPR[rs] - GPR[rt]
	// if (32_bit_arithmetic_overflow) then
	// 	SignalException(IntegerOverflow)
	// else
	// 	GPR[rd] ←temp
	// endif
	case OpCodeSUB:
		rsVal := int32(cpu.GetReg(ri.Rs))
		rtVal := int32(cpu.GetReg(ri.Rt))
		temp := rsVal - rtVal

		// Check for overflow
		if utils.CheckSubtractionOverflow(rsVal, rtVal, temp) {
			// Overflow occurred
			vec := cpu.cp0.RaiseException(excOv, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			return nil, false
		}

		cpu.SetReg(ri.Rd, uint32(temp))
		return nil, false

	// SUBU rd, rs, rt
	// if (NotWordValue(GPR[rs]) or NotWordValue(GPR[rt])) then UndefinedResult() endif
	// temp ←GPR[rs] - GPR[rt]
	// GPR[rd] ←temp
	case OpCodeSUBU:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		temp := rsVal - rtVal
		cpu.SetReg(ri.Rd, temp)
		return nil, false

	// TEQ rs, rt
	// if GPR[rs] = GPR[rt] then
	//	SignalException(Trap)
	// endif
	case OpCodeTEQ:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		if rsVal == rtVal {
			vec := cpu.cp0.RaiseException(excTr, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			return nil, false
		}
		return nil, false

	// TGE rs, rt
	// if GPR[rs] ≥ GPR[rt] then
	//	SignalException(Trap)
	// endif
	case OpCodeTGE:
		rsVal := int32(cpu.GetReg(ri.Rs))
		rtVal := int32(cpu.GetReg(ri.Rt))
		if rsVal >= rtVal {
			vec := cpu.cp0.RaiseException(excTr, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			return nil, false
		}
		return nil, false

	// TGEU rs, rt
	// if (0 || GPR[rs]) ≥ (0 || sign_extend(immediate)) then
	//	SignalException(Trap)
	// endif
	case OpCodeTGEU:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		if rsVal >= rtVal {
			vec := cpu.cp0.RaiseException(excTr, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			return nil, false
		}
		return nil, false

	// TLT rs, rt
	// if GPR[rs] < GPR[rt] then
	//	SignalException(Trap)
	// endif
	case OpCodeTLT:
		rsVal := int32(cpu.GetReg(ri.Rs))
		rtVal := int32(cpu.GetReg(ri.Rt))
		if rsVal < rtVal {
			vec := cpu.cp0.RaiseException(excTr, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			return nil, false
		}
		return nil, false

	// TLTU rs, rt
	// if (0 || GPR[rs]) < (0 || sign_extend(immediate)) then
	//	SignalException(Trap)
	// endif
	case OpCodeTLTU:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		if rsVal < rtVal {
			vec := cpu.cp0.RaiseException(excTr, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			return nil, false
		}
		return nil, false

	// TNE rs, rt
	// if GPR[rs] ≠ GPR[rt] then
	//	SignalException(Trap)
	// endif
	case OpCodeTNE:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		if rsVal != rtVal {
			vec := cpu.cp0.RaiseException(excTr, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			return nil, false
		}
		return nil, false

	// XOR rd, rs, rt
	// GPR[rd] ← GPR[rs] xor GPR[rt]
	case OpCodeXOR:
		rsVal := cpu.GetReg(ri.Rs)
		rtVal := cpu.GetReg(ri.Rt)
		temp := rsVal ^ rtVal
		cpu.SetReg(ri.Rd, temp)
		return nil, false
	}

	return nil, false
}

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

func (ii *ITypeInstruction) Execute(cpu *CPU) (nextPC *uint32, delaySlot bool) {
	return nil, false
}

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

func (ji JTypeInstruction) Execute(cpu *CPU) (nextPC *uint32, delaySlot bool) {
	return nil, false
}

// COP0Instruction handles all CP0 (coprocessor 0) instructions
type COP0Instruction struct {
	Opcode uint8 // Should be 0x10 for COP0
	Rs     uint8 // For MFC0/MTC0, this contains the funct code
	Rt     uint8 // Destination/source register
	Rd     uint8 // CP0 register number
	Sel    uint8 // CP0 register selector
	Funct  uint8 // Function code for ERET, TLBP, etc
}

func (ci COP0Instruction) Decode(instr uint32) Instruction {
	opcode := uint8((instr >> 26) & 0x3F)
	rs := uint8((instr >> 21) & 0x1F)
	rt := uint8((instr >> 16) & 0x1F)
	rd := uint8((instr >> 11) & 0x1F)
	sel := uint8(instr & 0x7)
	funct := uint8(instr & 0x3F)

	return &COP0Instruction{
		Opcode: opcode,
		Rs:     rs,
		Rt:     rt,
		Rd:     rd,
		Sel:    sel,
		Funct:  funct,
	}
}

func (ci COP0Instruction) Execute(cpu *CPU) (nextPC *uint32, delaySlot bool) {
	switch ci.Rs {
	case COP0Funct_MFC0:
		// Move From CP0: rt = CP0[rd,sel]
		val := cpu.GetCP0Reg(int(ci.Rd), int(ci.Sel))
		cpu.SetReg(ci.Rt, val)
		return nil, false

	case COP0Funct_MTC0:
		// Move To CP0: CP0[rd,sel] = rt
		val := cpu.GetReg(ci.Rt)
		cpu.SetCP0Reg(int(ci.Rd), int(ci.Sel), val)
		return nil, false

	case 0x10:
		// TLB and ERET instructions - check Funct field
		switch ci.Funct {
		case COP0Funct_ERET:
			// Exception Return: PC = EPC or ErrorEPC
			nextPCVal := cpu.cp0.ERET()
			return &nextPCVal, false

		case COP0Funct_TLBP:
			// TLB Probe: find entry matching EntryHi
			cpu.cp0.TLBP()
			return nil, false

		case COP0Funct_TLBR:
			// TLB Read: read entry at Index into EntryHi/EntryLo0/EntryLo1/PageMask
			cpu.cp0.TLBR()
			return nil, false

		case COP0Funct_TLBWI:
			// TLB Write Indexed: write entry from EntryHi/EntryLo0/EntryLo1/PageMask into TLB[Index]
			cpu.cp0.TLBWI()
			return nil, false

		case COP0Funct_TLBWR:
			// TLB Write Random: write entry into TLB[Random]
			cpu.cp0.TLBWR()
			return nil, false

		default:
			// Unknown TLB instruction
			return nil, false
		}

	default:
		// Unknown COP0 instruction
		return nil, false
	}
}

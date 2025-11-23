package mips32

import "testing"

func TestDecodeRTypeInstruction(t *testing.T) {
	// add $t0, $t1, $t2
	const instr = uint32(0x012A4020)

	var rInstr = RTypeInstruction{}

	decodedInstr := rInstr.Decode(instr)

	rtype, ok := decodedInstr.(*RTypeInstruction)
	if !ok {
		t.Fatal("Expected RTypeInstruction")
	}

	if rtype.Opcode != 0 {
		t.Errorf("Opcode = %d, want 0", rtype.Opcode)
	}
	if rtype.Rs != 9 {
		t.Errorf("Rs = %d, want 9 ($t1)", rtype.Rs)
	}
	if rtype.Rt != 10 {
		t.Errorf("Rt = %d, want 10 ($t2)", rtype.Rt)
	}
	if rtype.Rd != 8 {
		t.Errorf("Rd = %d, want 8 ($t0)", rtype.Rd)
	}
	if rtype.Shamt != 0 {
		t.Errorf("Shamt = %d, want 0", rtype.Shamt)
	}
	if rtype.Funct != 32 {
		t.Errorf("Funct = %d, want 32 (add)", rtype.Funct)
	}
}

func TestDecodeITypeInstruction(t *testing.T) {
	// addi $t0, $t1, 5
	const instr = uint32(0x21280005)

	var iInstr = ITypeInstruction{}
	decodedInstr := iInstr.Decode(instr)

	itype, ok := decodedInstr.(*ITypeInstruction)
	if !ok {
		t.Fatal("Expected ITypeInstruction")
	}

	if itype.Opcode != 8 {
		t.Errorf("Opcode = %d, want 8 (addi)", itype.Opcode)
	}
	if itype.Rs != 9 {
		t.Errorf("Rs = %d, want 9 ($t1)", itype.Rs)
	}
	if itype.Rt != 8 {
		t.Errorf("Rt = %d, want 8 ($t0)", itype.Rt)
	}
	if itype.Immediate != 5 {
		t.Errorf("Imm = %d, want 5", itype.Immediate)
	}
}

func TestDecodeJTypeInstruction(t *testing.T) {
	// j 0x00000040 (address field = 0x10)
	const instr = uint32(0x08000010)

	var jInstr = JTypeInstruction{}
	decodedInstr := jInstr.Decode(instr)

	jtype, ok := decodedInstr.(*JTypeInstruction)
	if !ok {
		t.Fatal("Expected JTypeInstruction")
	}

	if jtype.Opcode != 2 {
		t.Errorf("Opcode = %d, want 2 (j)", jtype.Opcode)
	}
	if jtype.Addr != 0x00000010 {
		t.Errorf("Address = 0x%X, want 0x10", jtype.Addr)
	}
}

package mips32

import (
	"testing"
)

// TestCOP0Integration tests the integration of CPU and COP0
func TestCOP0Integration(t *testing.T) {
	mem := NewMemory(0x10000) // 64KB memory
	cpu := NewCPU(mem)

	// Test 1: Verify cp0.Step() decrements Random
	initialRandom := cpu.cp0.random
	cpu.cp0.Step()
	if cpu.cp0.random >= initialRandom {
		t.Errorf("Random should decrement, got %d (was %d)", cpu.cp0.random, initialRandom)
	}

	// Test 2: Test SetBadVAddr
	testAddr := uint32(0x12345678)
	cpu.SetBadVAddr(testAddr)
	if cpu.cp0.badVAddr != testAddr {
		t.Errorf("BadVAddr not set correctly: got 0x%x, expected 0x%x", cpu.cp0.badVAddr, testAddr)
	}

	// Test 3: Test GetCP0Reg / SetCP0Reg
	statusVal := uint32(0x10000000)
	cpu.SetCP0Reg(12, 0, statusVal) // Write to Status register
	readVal := cpu.GetCP0Reg(12, 0)
	if readVal != statusVal {
		t.Errorf("CP0 register read/write failed: got 0x%x, expected 0x%x", readVal, statusVal)
	}

	// Test 4: Test MFC0 instruction (Move From CP0)
	// Set a value in the Status register via SetCP0Reg
	statusVal2 := uint32(0x20000000)
	cpu.SetCP0Reg(12, 0, statusVal2)

	// Create an MFC0 instruction: opcode=0x10, rs=0x00 (MFC0), rt=5, rd=12, sel=0
	// Encoding: 010000 00000 00101 01100 00 000
	// Hex: 0x40056000
	instr := uint32(0x40056000)
	decodedInstr := DecodeInstruction(instr)
	cop0Instr, ok := decodedInstr.(*COP0Instruction)
	if !ok {
		t.Errorf("Failed to decode as COP0Instruction")
	}

	// Execute MFC0: should move Status (CP0[12,0]) to register 5
	cop0Instr.Execute(cpu)
	val := cpu.GetReg(5)
	if val != statusVal2 {
		t.Errorf("MFC0 failed: got 0x%x, expected 0x%x", val, statusVal2)
	}

	// Test 5: Test MTC0 instruction (Move To CP0)
	// Set register 7 to a value
	testVal := uint32(0x30000000)
	cpu.SetReg(7, testVal)

	// Create an MTC0 instruction: opcode=0x10, rs=0x04 (MTC0), rt=7, rd=12, sel=0
	// Encoding: 010000 00100 00111 01100 00 000
	// Hex: 0x40876000
	instr2 := uint32(0x40876000)
	decodedInstr2 := DecodeInstruction(instr2)
	cop0Instr2, ok := decodedInstr2.(*COP0Instruction)
	if !ok {
		t.Errorf("Failed to decode as COP0Instruction for MTC0")
	}

	// Execute MTC0: should move register 7 to Status (CP0[12,0])
	cop0Instr2.Execute(cpu)
	readStatus := cpu.GetCP0Reg(12, 0)
	if readStatus != testVal {
		t.Errorf("MTC0 failed: got 0x%x, expected 0x%x", readStatus, testVal)
	}

	// Test 6: Test ERET instruction
	// Set EPC to a test value
	cpu.SetCP0Reg(14, 0, 0x80001000) // EPC
	// Set EXL bit (1 << 1 = 2) to test ERET
	currentStatus := cpu.GetCP0Reg(12, 0)
	cpu.SetCP0Reg(12, 0, currentStatus|(1<<1)) // Set EXL

	// Create an ERET instruction: opcode=0x10, rs=0x10, rt=0, rd=0, sel=0, funct=0x18
	// Encoding: 010000 10000 00000 00000 00 011000
	// Hex: 0x42000018
	instr3 := uint32(0x42000018)
	decodedInstr3 := DecodeInstruction(instr3)
	cop0Instr3, ok := decodedInstr3.(*COP0Instruction)
	if !ok {
		t.Errorf("Failed to decode as COP0Instruction for ERET")
	}

	nextPC, _ := cop0Instr3.Execute(cpu)
	if nextPC == nil || *nextPC != 0x80001000 {
		if nextPC == nil {
			t.Errorf("ERET failed: nextPC is nil")
		} else {
			t.Errorf("ERET failed: got PC 0x%x, expected 0x80001000", *nextPC)
		}
	}

	// Test 7: Verify ERET cleared EXL
	statusAfterEret := cpu.GetCP0Reg(12, 0)
	if (statusAfterEret & (1 << 1)) != 0 {
		t.Errorf("ERET should clear EXL")
	}

	// Test 8: Test TLB operations (TLBWI)
	// Set up EntryHi with VPN2 and ASID
	cpu.SetCP0Reg(10, 0, 0x80000001) // EntryHi: VPN2=0x80000, ASID=1
	cpu.SetCP0Reg(2, 0, 0x00000007)  // EntryLo0: PFN=0, C=7 (uncached), D=0, V=0, G=1
	cpu.SetCP0Reg(3, 0, 0x00000007)  // EntryLo1: PFN=0, C=7, D=0, V=0, G=1
	cpu.SetCP0Reg(5, 0, 0x00000000)  // PageMask=0

	// Create a TLBWI instruction: opcode=0x10, rs=0x10, funct=0x02
	// Encoding: 010000 10000 00000 00000 00 000010
	// Hex: 0x42000002
	instr4 := uint32(0x42000002)
	decodedInstr4 := DecodeInstruction(instr4)
	cop0Instr4, ok := decodedInstr4.(*COP0Instruction)
	if !ok {
		t.Errorf("Failed to decode as COP0Instruction for TLBWI")
	}

	cop0Instr4.Execute(cpu)

	// Verify the TLB entry was written at Index 0
	if cpu.cp0.tlb[0].VPN2 != 0x80000000 {
		t.Errorf("TLBWI failed: VPN2 not written correctly")
	}
}

// TestCOP0ExceptionHandling tests exception handling with COP0
func TestCOP0ExceptionHandling(t *testing.T) {
	mem := NewMemory(0x10000)
	cpu := NewCPU(mem)

	// Test: Verify Address Error sets BadVAddr
	badAddr := uint32(0xDEADBEEF)
	cpu.SetBadVAddr(badAddr)

	vec := cpu.cp0.RaiseException(excAdEL, cpu.PC, false)
	if cpu.cp0.badVAddr != badAddr {
		t.Errorf("BadVAddr not preserved: got 0x%x, expected 0x%x", cpu.cp0.badVAddr, badAddr)
	}

	// Verify exception vector is reasonable
	if vec == 0 {
		t.Errorf("Exception vector should not be 0")
	}

	// Test: Verify EPC is set correctly
	if cpu.cp0.epc != cpu.PC {
		t.Errorf("EPC not set correctly: got 0x%x, expected 0x%x", cpu.cp0.epc, cpu.PC)
	}
}

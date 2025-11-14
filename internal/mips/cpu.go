package mips

import (
	"log"
	"sync/atomic"
)

const cop0TlbSize int = 16

type CPU struct {
	registers [32]uint32
	PC        uint32
	Memory    *Memory
	running   atomic.Bool

	cp0     *COP0
	inDelay bool // Indicates if the CPU is in a delay slot
}

func NewCPU(mem *Memory) *CPU {
	return &CPU{
		registers: [32]uint32{},
		PC:        0,
		Memory:    mem,
		running:   atomic.Bool{},
		cp0:       NewCOP0(cop0TlbSize),
		inDelay:   false,
	}
}

// Run starts the CPU execution loop.
// It fetches, decodes, and executes instructions until stopped.
func (cpu *CPU) Run() {
	if cpu.running.Load() {
		log.Fatal("CPU already running")
		return
	}

	cpu.running.Store(true)

	for cpu.running.Load() {
		instr, ok := cpu.Memory.LoadWord(cpu.PC)
		if !ok {
			// Address error on instruction fetch
			cpu.SetBadVAddr(cpu.PC)
			vec := cpu.cp0.RaiseException(excAdEL, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			log.Printf("CPU exception at PC 0x%x, jumping to vector 0x%x", cpu.PC, vec)
			cpu.handleException(excAdEL)
			continue
		}

		// advance the COP0 per-instruction/cycle
		cpu.cp0.Tick(1)
		cpu.cp0.Step() // Update Random register per instruction

		// check pending interrupts
		if cpu.cp0.PendingInterrupt() {
			// RaiseException should return the exception vector/next PC
			vec := cpu.cp0.RaiseException(excInt, cpu.PC, cpu.inDelay)
			cpu.PC = vec
			cpu.inDelay = false
			continue
		}

		// we're decoding the instruction
		dInstr := DecodeInstruction(instr)

		// and we're executing it.
		newPC, delaySlot := dInstr.Execute(cpu)

		// set the delay slot flag
		cpu.inDelay = delaySlot

		// check if we have a new PC from the instruction
		if newPC != nil {
			// set the PC to the new value
			cpu.PC = *newPC
		} else {
			// otherwise, we just continue to the next instruction sequentially
			cpu.PC += 4
		}

	}

}

// Stop halts the CPU execution loop.
func (cpu *CPU) Stop() {
	cpu.running.Store(false)
}

// GetReg returns the value of register n (0-31)
func (cpu *CPU) GetReg(n int) uint32 {
	if n < 0 || n > 31 {
		return 0
	}
	return cpu.registers[n]
}

// SetReg sets the value of register n (0-31), except $0 which is always zero
func (cpu *CPU) SetReg(n int, val uint32) {
	if n < 0 || n > 31 {
		return
	}
	if n == 0 {
		return // $0 is always zero
	}
	cpu.registers[n] = val
}

// GetCP0Reg reads a CP0 register
func (cpu *CPU) GetCP0Reg(reg, sel int) uint32 {
	return cpu.cp0.Read(reg, sel)
}

// SetCP0Reg writes to a CP0 register
func (cpu *CPU) SetCP0Reg(reg, sel int, val uint32) {
	cpu.cp0.Write(reg, sel, val)
}

// SetBadVAddr sets the BadVAddr register in CP0
func (cpu *CPU) SetBadVAddr(addr uint32) {
	cpu.cp0.badVAddr = addr
}

// handleException processes the exception with the given code.
// The exception has already been raised via cp0.RaiseException() and PC is set to vector.
// This function handles any additional CPU-specific logic or logging.
func (cpu *CPU) handleException(exc int) {
	switch exc {
	case excInt:
		// Interrupt - handled by exception vector, execution continues
		log.Printf("Interrupt at PC 0x%x, vector 0x%x", cpu.cp0.epc, cpu.PC)

	case excMod:
		// TLB modification exception (write to read-only page)
		log.Printf("TLB Modification exception at PC 0x%x, BadVAddr 0x%x",
			cpu.cp0.epc, cpu.cp0.badVAddr)
		cpu.Stop()

	case excTLBL:
		// TLB exception on load/instruction fetch
		log.Printf("TLB Load exception at PC 0x%x, BadVAddr 0x%x",
			cpu.cp0.epc, cpu.cp0.badVAddr)
		cpu.Stop()

	case excTLBS:
		// TLB exception on store
		log.Printf("TLB Store exception at PC 0x%x, BadVAddr 0x%x",
			cpu.cp0.epc, cpu.cp0.badVAddr)
		cpu.Stop()

	case excAdEL:
		// Address error on load/instruction fetch (misaligned or invalid address)
		log.Printf("Address Error Load exception at PC 0x%x, BadVAddr 0x%x",
			cpu.cp0.epc, cpu.cp0.badVAddr)
		cpu.Stop()

	case excAdES:
		// Address error on store (misaligned or invalid address)
		log.Printf("Address Error Store exception at PC 0x%x, BadVAddr 0x%x",
			cpu.cp0.epc, cpu.cp0.badVAddr)
		cpu.Stop()

	case excSys:
		// Syscall exception - could be handled by OS/handler at vector
		log.Printf("Syscall exception at PC 0x%x", cpu.cp0.epc)
		// Don't stop - let exception handler deal with it

	case excBp:
		// Breakpoint exception
		log.Printf("Breakpoint exception at PC 0x%x", cpu.cp0.epc)
		cpu.Stop()

	case excRI:
		// Reserved instruction exception
		log.Printf("Reserved Instruction exception at PC 0x%x", cpu.cp0.epc)
		cpu.Stop()

	case excCpU:
		// Coprocessor unusable exception
		log.Printf("Coprocessor Unusable exception at PC 0x%x", cpu.cp0.epc)
		cpu.Stop()

	case excOv:
		// Arithmetic overflow exception
		log.Printf("Arithmetic Overflow exception at PC 0x%x", cpu.cp0.epc)
		cpu.Stop()

	case excTr:
		// Trap exception
		log.Printf("Trap exception at PC 0x%x", cpu.cp0.epc)
		cpu.Stop()

	default:
		// Unknown exception code
		log.Printf("Unknown exception %d at PC 0x%x", exc, cpu.PC)
		cpu.Stop()
	}
}

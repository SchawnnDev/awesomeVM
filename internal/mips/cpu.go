package mips

import "log"

type CPU struct {
	Registers [32]uint32
	PC        uint32
	Memory    *Memory
	running   bool
}

func NewCPU(mem *Memory) *CPU {
	return &CPU{
		Registers: [32]uint32{},
		PC:        0,
		Memory:    mem,
		running:   false,
	}
}

func (cpu *CPU) Run() {
	if cpu.running {
		log.Fatal("CPU already running")
		return
	}

	cpu.running = true

	for cpu.running {
		instr := uint32(0x1) // todo: change this to the Memory instr

		// we're decoding the instruction
		dInstr := DecodeInstruction(instr)

		// and we're executing it.
		dInstr.Execute(cpu)
	}

}

func (cpu *CPU) Stop() {
	cpu.running = false
}

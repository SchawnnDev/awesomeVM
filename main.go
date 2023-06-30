package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
)

const (
	MEMORY_MAX = 1 << 16
)

var (
	memory = [MEMORY_MAX]uint16{}
	reg    = [R_COUNT]uint16{}
)

func updateFlags(r uint16) {
	if reg[r] == 0 {
		reg[R_COND] = FL_ZRO
	} else if (reg[r] >> 15) == 1 { /* a 1 in the left-most bit indicates negative */
		reg[R_COND] = FL_NEG
	} else {
		reg[R_COND] = FL_POS
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("%s [file]\n", os.Args[0])
		return
	}

	var err error

	for i := 0; i < len(os.Args); i++ {
		err = ReadImage(os.Args[i])
		if err != nil {
			fmt.Printf("Failed to load image: %s\n", os.Args[i])
			return
		}
	}

	// Setup terminal
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// one condition flag should be set at any given time
	reg[R_COND] = FL_ZRO

	// set pc to start position (0x3000 -> default)
	const PC_START = 0x3000
	reg[R_PC] = PC_START

	running := 1

	for running == 1 {
		// fetch
		reg[R_PC]++
		instr := MemoryRead(reg[R_PC])
		// 16 bits numbers: 12 to 15 bits represents the operation type
		op := instr >> 12
		switch op {
		case OP_ADD:
			/*
				ADD operation:
					- 15 -> 12: 0001 (opcode)
				 	- 11 -> 9: DR (destination register)
					- 8 -> 6: SR1 (source register 1)
					- 5: add operation type: 0 for register mode, 1 for immediate mode
				register mode (0):
					- 4 -> 3: unused bits
					- 2 -> 0: SR2 (source register 2)
				immediate mode (1):
					- 4 -> 0: imm5 (2^5 = 32 unsigned max value)
			*/
			/* r0: destination register (DR)
			0001101011000010 >> 9 : 0000000000001101 & 0x7 -> 0000000000000101 (5)
			*/
			var r0 uint16 = (instr >> 9) & 0x7 // 0x7 -> 111 (only extract 3 right-most bits)
			var r1 uint16 = (instr >> 6) & 0x3 // 0x7 -> 111

			// check whether we are in immediate mode
			if (instr >> 5) & 0x1 {
				var imm5 = SignExtend(instr&0x1F, 5)
				reg[r0] = reg[r1] + imm5
			} else {
				var r2 uint16 = instr & 0x7
				reg[r0] = reg[r1] + reg[r2]
			}

			updateFlags(r0)
			break
		case OP_AND:
			break
		case OP_NOT:
			break
		case OP_BR:
			break
		case OP_JMP:
			break
		case OP_JSR:
			break
		case OP_LD:
			break
		case OP_LDI:
			break
		case OP_LDR:
			break
		case OP_LEA:
			break
		case OP_ST:
			break
		case OP_STI:
			break
		case OP_STR:
			break
		case OP_TRAP:
			break
		case OP_RES:
			fallthrough
		case OP_RTI:
			fallthrough
		default:
			// Bad opcode
			break

		}
	}

}

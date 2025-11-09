package main

import (
	"awesomeVM/internal/lc3"
	"fmt"
	"log"
	"os"

	"github.com/eiannone/keyboard"
)

var (
	reg = [lc3.R_COUNT]uint16{}
)

func updateFlags(r uint16) {
	if reg[r] == 0 {
		reg[lc3.R_COND] = lc3.FL_ZRO
	} else if (reg[r] >> 15) == 0x1 { /* a 1 in the left-most bit indicates negative */
		reg[lc3.R_COND] = lc3.FL_NEG
	} else {
		reg[lc3.R_COND] = lc3.FL_POS
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("%s [file]\n", os.Args[0])
		return
	}

	var err error

	for i := 1; i < len(os.Args); i++ {
		err = lc3.ReadImage(os.Args[i])
		if err != nil {
			fmt.Printf("Failed to load image: %s\n", os.Args[i])
			return
		}
	}

	// Setup terminal
	//oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	//if err != nil {
	//	panic(err)
	//}
	//defer term.Restore(int(os.Stdin.Fd()), oldState)

	// one condition flag should be set at any given time
	reg[lc3.R_COND] = lc3.FL_ZRO

	// set pc to start position (0x3000 -> default)
	const PC_START = 0x3000
	reg[lc3.R_PC] = PC_START

	running := 1

	for running == 1 {
		// fetch
		instr := lc3.MemoryRead(reg[lc3.R_PC])
		reg[lc3.R_PC]++
		// 16 bits numbers: 12 to 15 bits represents the operation type
		op := instr >> 12

		switch op {
		case lc3.OP_BR: // conditional branch
			condFlag := (instr >> 9) & 0x7
			if (condFlag & reg[lc3.R_COND]) != 0 {
				reg[lc3.R_PC] += lc3.SignExtend(instr&0x1FF, 9)
			}
			break

		case lc3.OP_ADD:
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
			var r1 uint16 = (instr >> 6) & 0x7 // 0x7 -> 111

			// check whether we are in immediate mode
			if ((instr >> 5) & 0x1) != 0 {
				var imm5 = lc3.SignExtend(instr&0x1F, 5)
				reg[r0] = reg[r1] + imm5
			} else {
				var r2 uint16 = instr & 0x7
				reg[r0] = reg[r1] + reg[r2]
			}

			updateFlags(r0)
			break
		case lc3.OP_AND:
			var r0 uint16 = (instr >> 9) & 0x7 // 0x7 -> 111 (only extract 3 right-most bits)
			var r1 uint16 = (instr >> 6) & 0x7

			// check whether we are in immediate mode
			if ((instr >> 5) & 0x1) != 0 {
				var imm5 = lc3.SignExtend(instr&0x1F, 5)
				reg[r0] = reg[r1] & imm5
			} else {
				var r2 uint16 = instr & 0x7
				reg[r0] = reg[r1] & reg[r2]
			}

			updateFlags(r0)
			break
		case lc3.OP_JMP: // JMP & RET have same OP CODE: RET = JMP R7
			var baseR uint16 = (instr >> 6) & 0x7
			// PC jumps to the location specified in baseR
			// RET jumps automatically to R7
			reg[lc3.R_PC] = reg[baseR]
			break
		case lc3.OP_NOT:
			var r0 uint16 = (instr >> 9) & 0x7 // DR
			var r1 uint16 = (instr >> 6) & 0x7 // SR
			reg[r0] = ^reg[r1]
			updateFlags(r0)
			break
		case lc3.OP_JSR: // jump register
			reg[lc3.R_R7] = reg[lc3.R_PC] // save PC into R7

			if ((instr >> 11) & 0x1) != 0 {
				// JSR
				var PCoffset11 uint16 = instr & 0x7FF
				reg[lc3.R_PC] += lc3.SignExtend(PCoffset11, 11)
			} else {
				// JSRR
				var baseR uint16 = (instr >> 6) & 0x7
				reg[lc3.R_PC] = reg[baseR]
			}

			break
		case lc3.OP_LD: // load
			var r0 uint16 = (instr >> 9) & 0x7 // DR
			reg[r0] = lc3.MemoryRead(reg[lc3.R_PC] + lc3.SignExtend(instr&0x1FF, 9))
			updateFlags(r0)
			break
		case lc3.OP_LDI: // load indirect
			var r0 uint16 = (instr >> 9) & 0x7             // destination register
			var PCoffset9 = lc3.SignExtend(instr&0x1FF, 9) // PCoffset 9
			/*
				add PCoffset to the current PC, look at that memory location to get the final address

				In memory, it may be layed out like this:

				Address Label      Value
				0x123:  far_data = 0x456
				...
				0x456:  string   = 'a'

				if PC was at 0x100
				LDI R0 0x023
				would load 'a' into R0
			*/
			reg[r0] = lc3.MemoryRead(lc3.MemoryRead(reg[lc3.R_PC] + PCoffset9))
			updateFlags(r0)
			break
		case lc3.OP_LDR: // load register
			var r0 uint16 = (instr >> 9) & 0x7 // DR
			var baseR uint16 = (instr >> 6) & 0x7
			reg[r0] = lc3.MemoryRead(reg[baseR] + lc3.SignExtend(instr&0x3F, 6))
			updateFlags(r0)
			break
		case lc3.OP_LEA: // load effective address
			var r0 uint16 = (instr >> 9) & 0x7 // DR
			reg[r0] = reg[lc3.R_PC] + lc3.SignExtend(instr&0x1FF, 9)
			updateFlags(r0)
			break
		case lc3.OP_ST: // store
			var r0 uint16 = (instr >> 9) & 0x7 // SR
			// The contents of the register specified by SR are stored in the memory location
			lc3.MemoryWrite(reg[lc3.R_PC]+lc3.SignExtend(instr&0x1FF, 9), reg[r0])
			break
		case lc3.OP_STI: // store indirect
			var r0 uint16 = (instr >> 9) & 0x7 // SR
			// The contents of the register specified by SR are stored in the memory location
			lc3.MemoryWrite(lc3.MemoryRead(reg[lc3.R_PC]+lc3.SignExtend(instr&0x1FF, 9)), reg[r0])
			break
		case lc3.OP_STR: // store register
			var r0 uint16 = (instr >> 9) & 0x7 // SR
			var baseR uint16 = (instr >> 6) & 0x7
			lc3.MemoryWrite(reg[baseR]+lc3.SignExtend(instr&0x3F, 6), reg[r0])
			break
		case lc3.OP_TRAP:
			var trapVect8 uint16 = instr & 0xFF // trap code
			reg[lc3.R_R7] = reg[lc3.R_PC]

			switch trapVect8 {
			case lc3.TRAP_GETC: // get char without printing and save it to r0
				ch, key, err := keyboard.GetSingleKey()
				if err != nil {
					log.Fatal("[OP_TRAP] TRAP_GETC: Could not read single char from terminal")
				}

				if key == keyboard.KeyCtrlC {
					log.Fatal("interrupt")
				}
				reg[lc3.R_R0] = uint16(ch)
				updateFlags(lc3.R_R0)
				break
			case lc3.TRAP_OUT: // print one char
				fmt.Printf("%c", reg[lc3.R_R0])
				break
			case lc3.TRAP_PUTS:
				i := uint16(0)

				for {
					c := lc3.MemoryRead(reg[lc3.R_R0] + i)
					if c == 0 {
						break
					}
					fmt.Printf("%c", c)
					i++
				}

				break
			case lc3.TRAP_IN:
				var char [1]byte
				ch, key, err := keyboard.GetSingleKey()
				if err != nil {
					log.Fatal("[OP_TRAP] TRAP_IN: Could not read single char from terminal", err)
				}

				if key == keyboard.KeyCtrlC {
					log.Fatal("interrupt")
				}

				reg[lc3.R_R0] = uint16(ch)
				updateFlags(lc3.R_R0)
				fmt.Printf("%c", char[0])
				break
			case lc3.TRAP_PUTSP:
				i := reg[lc3.R_R0]

				for {
					c := lc3.MemoryRead(i)
					if c == 0 {
						break
					}
					fmt.Printf("%c%c", c&0xFF, c>>8)
					i++
				}

				break
			case lc3.TRAP_HALT:
				fmt.Printf("HALT")
				running = 0
				break
			}

			break
		case lc3.OP_RES:
			fallthrough
		case lc3.OP_RTI:
			fallthrough
		default:
			log.Fatal("Bad opcode: ", op) // RES & RTI -> Bad opcode
		}
	}

}

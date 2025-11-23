package main

import (
	"debug/elf"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Usage: go run main.go [-endian=auto|big|little] <mips32_binary_file>")
		return
	}

	fileName := flag.Arg(0)
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close file: %v", err)
		}
	}()

	// Try to parse as ELF file
	elfFile, err := elf.Open(fileName)
	if err == nil {
		defer func() {
			if err := elfFile.Close(); err != nil {
				log.Printf("Failed to close ELF file: %v", err)
			}
		}()
		disassembleELF(elfFile)
		return
	}

	// If not ELF, treat as raw binary
	fmt.Println("Not an ELF file, treating as raw binary")
	disassembleRaw(file)
}

func disassembleELF(elfFile *elf.File) {
	fmt.Printf("ELF File: %s\n", elfFile.Machine)
	fmt.Printf("Entry point: 0x%08X\n", elfFile.Entry)
	fmt.Println()

	// Decide byte order based on ELF endianness
	var order binary.ByteOrder
	if elfFile.ByteOrder == binary.LittleEndian {
		order = binary.LittleEndian
		fmt.Println("Using byte order: little-endian (from ELF header)")
	} else {
		order = binary.BigEndian
		fmt.Println("Using byte order: big-endian (from ELF header)")
	}
	fmt.Println()

	// Print all sections for information
	fmt.Println("ELF Sections:")
	fmt.Println("-------------")
	for _, section := range elfFile.Sections {
		fmt.Printf("  %-20s Type: %-15s Addr: 0x%08X Size: %-8d Flags: %s\n",
			section.Name,
			section.Type.String(),
			section.Addr,
			section.Size,
			sectionFlagsString(section.Flags))
	}
	fmt.Println()

	// Find and disassemble .text section
	textSection := elfFile.Section(".text")
	if textSection == nil {
		fmt.Println("Warning: No .text section found")

		// Try to find any executable section
		for _, section := range elfFile.Sections {
			if section.Flags&elf.SHF_EXECINSTR != 0 {
				fmt.Printf("Found executable section: %s\n", section.Name)
				disassembleSection(section, order)
			}
		}
		return
	}

	fmt.Printf("Disassembling .text section (0x%08X - 0x%08X):\n", textSection.Addr, textSection.Addr+textSection.Size)
	fmt.Println("=======================================================================")
	disassembleSection(textSection, order)
}

func disassembleSection(section *elf.Section, order binary.ByteOrder) {
	data, err := section.Data()
	if err != nil {
		log.Printf("Failed to read section %s: %v", section.Name, err)
		return
	}

	addr := section.Addr
	for i := 0; i < len(data); i += 4 {
		if i+4 > len(data) {
			break
		}

		inst := order.Uint32(data[i : i+4])
		fmt.Printf("0x%08X: 0x%08X\t%s\n", addr+uint64(i), inst, disassemble(inst, uint32(addr+uint64(i))))
	}
}

func sectionFlagsString(flags elf.SectionFlag) string {
	var result string
	if flags&elf.SHF_WRITE != 0 {
		result += "W"
	}
	if flags&elf.SHF_ALLOC != 0 {
		result += "A"
	}
	if flags&elf.SHF_EXECINSTR != 0 {
		result += "X"
	}
	if result == "" {
		result = "-"
	}
	return result
}

func disassembleRaw(file *os.File) {
	// Decide byte order: force big-endian for raw files
	var order binary.ByteOrder = binary.BigEndian
	fmt.Println("Using byte order: big-endian (forced)")

	// Begin reading from start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Fatalf("Failed to seek file: %v", err)
	}

	var offset int64 = 0
	for {
		var inst uint32
		err := binary.Read(file, order, &inst)
		if err != nil {
			break
		}

		fmt.Printf("0x%08X: 0x%08X\t%s\n", offset, inst, disassemble(inst, uint32(offset)))
		offset += 4
	}
}

// Simple MIPS disassembler for a few instructions
func disassemble(inst uint32, pc uint32) string {
	op := inst >> 26

	switch op {
	case 0x0: // R-type
		return disassembleR(inst)
	case 0x1: // REGIMM
		return disassembleRegimm(inst, pc)
	case 0x2: // J
		addr := inst & 0x3FFFFFF
		// target = (pc & 0xF0000000) | (addr << 2)
		target := ((pc + 4) & 0xF0000000) | (addr << 2)
		return fmt.Sprintf("j 0x%08X", target)
	case 0x3: // JAL
		addr := inst & 0x3FFFFFF
		target := ((pc + 4) & 0xF0000000) | (addr << 2)
		return fmt.Sprintf("jal 0x%08X", target)
	default: // I-type
		return disassembleI(op, inst, pc)
	}
}

func disassembleR(inst uint32) string {
	rs := (inst >> 21) & 0x1F
	rt := (inst >> 16) & 0x1F
	rd := (inst >> 11) & 0x1F
	shamt := (inst >> 6) & 0x1F
	funct := inst & 0x3F

	switch funct {
	case 0x20:
		return fmt.Sprintf("add $%d, $%d, $%d", rd, rs, rt)
	case 0x21:
		return fmt.Sprintf("addu $%d, $%d, $%d", rd, rs, rt)
	case 0x22:
		return fmt.Sprintf("sub $%d, $%d, $%d", rd, rs, rt)
	case 0x23:
		return fmt.Sprintf("subu $%d, $%d, $%d", rd, rs, rt)
	case 0x24:
		return fmt.Sprintf("and $%d, $%d, $%d", rd, rs, rt)
	case 0x25:
		return fmt.Sprintf("or $%d, $%d, $%d", rd, rs, rt)
	case 0x2A:
		return fmt.Sprintf("slt $%d, $%d, $%d", rd, rs, rt)
	case 0x00:
		return fmt.Sprintf("sll $%d, $%d, %d", rd, rt, shamt)
	case 0x02:
		return fmt.Sprintf("srl $%d, $%d, %d", rd, rt, shamt)
	case 0x03:
		return fmt.Sprintf("sra $%d, $%d, %d", rd, rt, shamt)
	case 0x04:
		return fmt.Sprintf("sllv $%d, $%d, $%d", rd, rt, rs)
	case 0x06:
		return fmt.Sprintf("srlv $%d, $%d, $%d", rd, rt, rs)
	case 0x07:
		return fmt.Sprintf("srav $%d, $%d, $%d", rd, rt, rs)
	case 0x08:
		return fmt.Sprintf("jr $%d", rs)
	case 0x09:
		return fmt.Sprintf("jalr $%d", rs)
	case 0x10:
		return fmt.Sprintf("mfhi $%d", rd)
	case 0x12:
		return fmt.Sprintf("mflo $%d", rd)
	case 0x18:
		return fmt.Sprintf("mult $%d, $%d", rs, rt)
	case 0x19:
		return fmt.Sprintf("multu $%d, $%d", rs, rt)
	case 0x1A:
		return fmt.Sprintf("div $%d, $%d", rs, rt)
	case 0x1B:
		return fmt.Sprintf("divu $%d, $%d", rs, rt)
	case 0x0C:
		return fmt.Sprintf("syscall")
	case 0x0D:
		return fmt.Sprintf("break")
	case 0x0F:
		return fmt.Sprintf("sync")
	default:
		return fmt.Sprintf("unknown R-funct 0x%02X", funct)
	}
}

func disassembleI(op, inst uint32, pc uint32) string {
	rs := (inst >> 21) & 0x1F
	rt := (inst >> 16) & 0x1F
	imm := inst & 0xFFFF

	signExt := int32(int16(imm))

	switch op {
	case 0x08:
		return fmt.Sprintf("addi $%d, $%d, %d", rt, rs, int16(imm))
	case 0x09:
		return fmt.Sprintf("addiu $%d, $%d, %d", rt, rs, int16(imm))
	case 0x0C:
		return fmt.Sprintf("andi $%d, $%d, %d", rt, rs, imm)
	case 0x0D:
		return fmt.Sprintf("ori $%d, $%d, %d", rt, rs, imm)
	case 0x0E:
		return fmt.Sprintf("xori $%d, $%d, %d", rt, rs, imm)
	case 0x0A:
		return fmt.Sprintf("slti $%d, $%d, %d", rt, rs, int16(imm))
	case 0x0B:
		return fmt.Sprintf("sltiu $%d, $%d, %d", rt, rs, int16(imm))
	case 0x0F:
		return fmt.Sprintf("lui $%d, 0x%04X", rt, imm)
	case 0x23:
		return fmt.Sprintf("lw $%d, %d($%d)", rt, int16(imm), rs)
	case 0x20:
		return fmt.Sprintf("lb $%d, %d($%d)", rt, int16(imm), rs)
	case 0x21:
		return fmt.Sprintf("lh $%d, %d($%d)", rt, int16(imm), rs)
	case 0x24:
		return fmt.Sprintf("lbu $%d, %d($%d)", rt, int16(imm), rs)
	case 0x25:
		return fmt.Sprintf("lhu $%d, %d($%d)", rt, int16(imm), rs)
	case 0x22:
		return fmt.Sprintf("lwl $%d, %d($%d)", rt, int16(imm), rs)
	case 0x26:
		return fmt.Sprintf("lwr $%d, %d($%d)", rt, int16(imm), rs)
	case 0x2B:
		return fmt.Sprintf("sw $%d, %d($%d)", rt, int16(imm), rs)
	case 0x28:
		return fmt.Sprintf("sb $%d, %d($%d)", rt, int16(imm), rs)
	case 0x29:
		return fmt.Sprintf("sh $%d, %d($%d)", rt, int16(imm), rs)
	case 0x2A:
		return fmt.Sprintf("swl $%d, %d($%d)", rt, int16(imm), rs)
	case 0x2E:
		return fmt.Sprintf("swr $%d, %d($%d)", rt, int16(imm), rs)
	case 0x30: // ll
		return fmt.Sprintf("ll $%d, %d($%d)", rt, int16(imm), rs)
	case 0x38: // sc
		return fmt.Sprintf("sc $%d, %d($%d)", rt, int16(imm), rs)
	case 0x31: // lwc1
		return fmt.Sprintf("lwc1 $f%d, %d($%d)", rt, int16(imm), rs)
	case 0x32: // lwc2
		return fmt.Sprintf("lwc2 %d, %d($%d)", rt, int16(imm), rs)
	case 0x35: // ldc1
		return fmt.Sprintf("ldc1 $f%d, %d($%d)", rt, int16(imm), rs)
	case 0x36: // ldc2
		return fmt.Sprintf("ldc2 %d, %d($%d)", rt, int16(imm), rs)
	case 0x39: // swc1
		return fmt.Sprintf("swc1 $f%d, %d($%d)", rt, int16(imm), rs)
	case 0x3A: // swc2
		return fmt.Sprintf("swc2 %d, %d($%d)", rt, int16(imm), rs)
	case 0x3D: // sdc1
		return fmt.Sprintf("sdc1 $f%d, %d($%d)", rt, int16(imm), rs)
	case 0x3E: // sdc2
		return fmt.Sprintf("sdc2 %d, %d($%d)", rt, int16(imm), rs)
	case 0x04: // beq
		offset := signExt << 2
		target := pc + 4 + uint32(offset)
		return fmt.Sprintf("beq $%d, $%d, 0x%08X", rs, rt, target)
	case 0x05: // bne
		offset := signExt << 2
		target := pc + 4 + uint32(offset)
		return fmt.Sprintf("bne $%d, $%d, 0x%08X", rs, rt, target)
	case 0x06: // blez
		offset := signExt << 2
		target := pc + 4 + uint32(offset)
		return fmt.Sprintf("blez $%d, 0x%08X", rs, target)
	case 0x07: // bgtz
		offset := signExt << 2
		target := pc + 4 + uint32(offset)
		return fmt.Sprintf("bgtz $%d, 0x%08X", rs, target)
	case 0x10:
		return disassembleCop0(inst)
	case 0x11:
		return disassembleCop1(inst)
	case 0x12:
		return disassembleCop2(inst)
	default:
		return fmt.Sprintf("unknown I-op 0x%02X", op)
	}
}

func disassembleRegimm(inst uint32, pc uint32) string {
	rs := (inst >> 21) & 0x1F
	rt := (inst >> 16) & 0x1F
	imm := inst & 0xFFFF

	signExt := int32(int16(imm))
	offset := signExt << 2
	target := pc + 4 + uint32(offset)

	switch rt {
	case 0x00: // bltz
		return fmt.Sprintf("bltz $%d, 0x%08X", rs, target)
	case 0x01: // bgez
		return fmt.Sprintf("bgez $%d, 0x%08X", rs, target)
	case 0x10: // bltzal
		return fmt.Sprintf("bltzal $%d, 0x%08X", rs, target)
	case 0x11: // bgezal
		return fmt.Sprintf("bgezal $%d, 0x%08X", rs, target)
	default:
		return fmt.Sprintf("unknown regimm rt=0x%02X", rt)
	}
}

func disassembleCop0(inst uint32) string {
	rs := (inst >> 21) & 0x1F
	rt := (inst >> 16) & 0x1F
	rd := (inst >> 11) & 0x1F

	switch rs {
	case 0x00: // mfc0
		return fmt.Sprintf("mfc0 $%d, $%d", rt, rd)
	case 0x04: // mtc0
		return fmt.Sprintf("mtc0 $%d, $%d", rt, rd)
	case 0x10: // CO
		funct := inst & 0x3F
		switch funct {
		case 0x01:
			return "tlbr"
		case 0x02:
			return "tlbwi"
		case 0x06:
			return "tlbwr"
		case 0x08:
			return "tlbp"
		case 0x18:
			return "eret"
		default:
			return fmt.Sprintf("cop0-co funct=0x%02X", funct)
		}
	default:
		return fmt.Sprintf("unknown cop0 rs=0x%02X", rs)
	}
}

func disassembleCop1(inst uint32) string {
	// FPU instructions
	rs := (inst >> 21) & 0x1F
	switch rs {
	case 0x00: // mfc1
		rt := (inst >> 16) & 0x1F
		fs := (inst >> 11) & 0x1F
		return fmt.Sprintf("mfc1 $%d, $f%d", rt, fs)
	case 0x04: // mtc1
		rt := (inst >> 16) & 0x1F
		fs := (inst >> 11) & 0x1F
		return fmt.Sprintf("mtc1 $%d, $f%d", rt, fs)
	case 0x02: // cfc1
		rt := (inst >> 16) & 0x1F
		fs := (inst >> 11) & 0x1F
		return fmt.Sprintf("cfc1 $%d, $f%d", rt, fs)
	case 0x06: // ctc1
		rt := (inst >> 16) & 0x1F
		fs := (inst >> 11) & 0x1F
		return fmt.Sprintf("ctc1 $%d, $f%d", rt, fs)
	case 0x08: // BC
		// bc1t, bc1f
		return "bc1..."
	case 0x10: // S
		return "cop1-s"
	case 0x11: // D
		return "cop1-d"
	case 0x14: // W
		return "cop1-w"
	default:
		return fmt.Sprintf("unknown cop1 rs=0x%02X", rs)
	}
}

func disassembleCop2(_ uint32) string {
	return "cop2 instruction"
}

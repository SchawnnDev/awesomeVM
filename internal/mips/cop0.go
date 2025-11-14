package mips

// COP0 implements a simplified but comprehensive MIPS32r1/r2 CP0 inspired by QEMU.
// It models key registers, TLB operations, exceptions/interrupts, and a timer.
// The API is self-contained; the CPU can call into it to handle system control ops.
// All registers are 32-bit. Some fields are approximated for practicality in this repo.

type COP0 struct {
	// TLB state
	tlb     []TLBEntry
	tlbSize int

	// CP0 architectural registers (selected common ones)
	index    uint32 // 0, sel0: [5:0]=index, [31]=P (probe fail)
	random   uint32 // 1, sel0: random in [wired..tlbSize-1]
	entryLo0 uint32 // 2, sel0
	entryLo1 uint32 // 3, sel0
	context  uint32 // 4, sel0
	pageMask uint32 // 5, sel0
	wired    uint32 // 6, sel0

	badVAddr uint32 // 8, sel0
	count    uint32 // 9, sel0 (incremented by Tick)
	entryHi  uint32 // 10, sel0
	compare  uint32 // 11, sel0

	status uint32 // 12, sel0
	cause  uint32 // 13, sel0
	epc    uint32 // 14, sel0

	prid  uint32 // 15, sel0 (readonly)
	ebase uint32 // 15, sel1 (vector base)

	config0 uint32 // 16, sel0
	config1 uint32 // 16, sel1 (readonly)

	lladdr  uint32 // 17, sel0
	watchLo uint32 // 18, sel0
	watchHi uint32 // 19, sel0

	xcontext uint32 // 20, sel0 (optional, r2+); kept for completeness

	errorepc uint32 // 30, sel0
}

// TLBEntry models a two-page (even/odd) MIPS TLB entry.
type TLBEntry struct {
	VPN2 uint32 // [31:13] of virtual address (VPN2), aligned to 4KB*2
	ASID uint8  // [7:0] from EntryHi
	G    bool   // Global bit (both EntryLo G must be 1 => effective G)

	PFN0 uint32 // even page PFN (EntryLo0 [25:6])
	C0   uint8  // cache attribute (3 bits)
	D0   bool   // dirty (write permission)
	V0   bool   // valid

	PFN1 uint32 // odd page PFN (EntryLo1 [25:6])
	C1   uint8
	D1   bool
	V1   bool

	Mask uint32 // PageMask register value for this entry
}

// Register numbers and constants
const (
	cp0RegIndex    = 0
	cp0RegRandom   = 1
	cp0RegEntryLo0 = 2
	cp0RegEntryLo1 = 3
	cp0RegContext  = 4
	cp0RegPageMask = 5
	cp0RegWired    = 6
	cp0RegBadVAddr = 8
	cp0RegCount    = 9
	cp0RegEntryHi  = 10
	cp0RegCompare  = 11
	cp0RegStatus   = 12
	cp0RegCause    = 13
	cp0RegEPC      = 14
	cp0RegPRId     = 15
	cp0RegConfig   = 16
	cp0RegLLAddr   = 17
	cp0RegWatchLo  = 18
	cp0RegWatchHi  = 19
	cp0RegXContext = 20
	cp0RegErrorEPC = 30
)

// Cause ExcCode for exceptions (subset)
const (
	excInt  = 0  // Interrupt
	excMod  = 1  // TLB modification
	excTLBL = 2  // TLB load/fetch
	excTLBS = 3  // TLB store
	excAdEL = 4  // Address error load/fetch
	excAdES = 5  // Address error store
	excSys  = 8  // Syscall
	excBp   = 9  // Breakpoint
	excRI   = 10 // Reserved instruction
	excCpU  = 11 // Coprocessor unusable
	excOv   = 12 // Arithmetic overflow
	excTr   = 13 // Trap
)

// Status/Cause bit helpers
const (
	statusIE  uint32 = 1 << 0
	statusEXL uint32 = 1 << 1
	statusERL uint32 = 1 << 2

	// Interrupt mask bits IM[7:0] at [15:8]
	statusIMShift = 8

	causeBD uint32 = 1 << 31
	causeTI uint32 = 1 << 30 // Timer interrupt
	causeIV uint32 = 1 << 23 // Interrupt vector select

	// IP bits at [15:8]
	causeIPShift = 8
)

// NewCOP0 creates a new CP0 with a TLB of the given size.
func NewCOP0(tlbSize int) *COP0 {
	if tlbSize <= 0 {
		tlbSize = 16
	}
	c := &COP0{
		tlb:     make([]TLBEntry, tlbSize),
		tlbSize: tlbSize,
	}

	// Defaults similar to reset state
	c.random = uint32(tlbSize - 1)
	c.wired = 0

	// Status: BEV=1 on many MIPS at reset (Status.BEV bit is in r4000; here we approximate via EBase high bit, we still use vectors BEV=1 path by default)
	// We'll imply BEV via ebase default >= 0x80000000.

	c.prid = 0x00018000 // arbitrary implementer/version
	c.ebase = 0x80000000

	// Config0: set K0=3 uncached, set M=1 (has Config1)
	// Config0 layout: [31:0] implementation specific. We'll set:
	// - K0 at [2:0] = 3 (Cacheable noncoherent)
	// - M at [31] = 1 to indicate Config1 present
	c.config0 = (1 << 31) | 0x3

	// Config1: encode TLB size (MMU size = (N-1)), in fields [25:22] (MMU size) per MIPS32r2
	// We'll set MMs (TLB entries - 1) in MMU size.
	mmuSize := uint32(tlbSize - 1)
	c.config1 = (mmuSize & 0xF) << 25 // Note: actual field is [25:22]; we place starting at 25 and ignore other fields.

	// Count/Compare zeroed.

	return c
}

// Read returns the value of CP0 register (reg,sel).
func (c *COP0) Read(reg, sel int) uint32 {
	switch reg {
	case cp0RegIndex:
		if sel == 0 {
			return c.index
		}
	case cp0RegRandom:
		if sel == 0 {
			return c.random
		}
	case cp0RegEntryLo0:
		if sel == 0 {
			return c.entryLo0
		}
	case cp0RegEntryLo1:
		if sel == 0 {
			return c.entryLo1
		}
	case cp0RegContext:
		if sel == 0 {
			return c.context
		}
	case cp0RegPageMask:
		if sel == 0 {
			return c.pageMask
		}
	case cp0RegWired:
		if sel == 0 {
			return c.wired
		}
	case cp0RegBadVAddr:
		if sel == 0 {
			return c.badVAddr
		}
	case cp0RegCount:
		if sel == 0 {
			return c.count
		}
	case cp0RegEntryHi:
		if sel == 0 {
			return c.entryHi
		}
	case cp0RegCompare:
		if sel == 0 {
			return c.compare
		}
	case cp0RegStatus:
		if sel == 0 {
			return c.status
		}
	case cp0RegCause:
		if sel == 0 {
			return c.cause
		}
	case cp0RegEPC:
		if sel == 0 {
			return c.epc
		}
	case cp0RegPRId:
		switch sel {
		case 0:
			return c.prid
		case 1:
			return c.ebase
		}
	case cp0RegConfig:
		switch sel {
		case 0:
			return c.config0
		case 1:
			return c.config1
		}
	case cp0RegLLAddr:
		if sel == 0 {
			return c.lladdr
		}
	case cp0RegWatchLo:
		if sel == 0 {
			return c.watchLo
		}
	case cp0RegWatchHi:
		if sel == 0 {
			return c.watchHi
		}
	case cp0RegXContext:
		if sel == 0 {
			return c.xcontext
		}
	case cp0RegErrorEPC:
		if sel == 0 {
			return c.errorepc
		}
	}
	return 0
}

// Write sets the value of CP0 register (reg,sel) and applies side effects.
func (c *COP0) Write(reg, sel int, val uint32) {
	switch reg {
	case cp0RegIndex:
		if sel == 0 {
			// [5:0] index; [31] P bit is read-only, preserved
			p := c.index & 0x80000000
			idx := val & 0x3F
			if int(idx) >= c.tlbSize {
				idx = uint32(c.tlbSize - 1)
			}
			c.index = p | idx
		}
	case cp0RegRandom:
		if sel == 0 {
			// Random is read-only in hardware; allow write for flexibility but clamp
			if val < c.wired {
				val = c.wired
			}
			rmax := uint32(c.tlbSize - 1)
			if val > rmax {
				val = rmax
			}
			c.random = val
		}
	case cp0RegEntryLo0:
		if sel == 0 {
			c.entryLo0 = val & 0x3FFFFFFF // keep 30 LSBs (PFN[25:6], C[5:3], D[2], V[1], G[0])
		}
	case cp0RegEntryLo1:
		if sel == 0 {
			c.entryLo1 = val & 0x3FFFFFFF
		}
	case cp0RegContext:
		if sel == 0 {
			c.context = val
		}
	case cp0RegPageMask:
		if sel == 0 {
			// Accept as-is; hardware restricts to valid masks
			c.pageMask = val & 0x01FFE000 // bits [24:13]
		}
	case cp0RegWired:
		if sel == 0 {
			c.wired = val & 0x3F
			if c.wired >= uint32(c.tlbSize) {
				c.wired = uint32(c.tlbSize - 1)
			}
			// Adjust random range
			rmax := uint32(c.tlbSize - 1)
			if c.random < c.wired || c.random > rmax {
				c.random = rmax
			}
		}
	case cp0RegBadVAddr:
		if sel == 0 {
			// RO typically; ignore writes
		}
	case cp0RegCount:
		if sel == 0 {
			c.count = val
		}
	case cp0RegEntryHi:
		if sel == 0 {
			// EntryHi: VPN2 [31:13], ASID [7:0]
			c.entryHi = (val & 0xFFFFE0FF) // keep VPN2 and ASID; ignore R/WI wired bits
		}
	case cp0RegCompare:
		if sel == 0 {
			c.compare = val
			// Writing Compare clears TI and IP7
			c.cause &^= causeTI | (1 << (causeIPShift + 7))
		}
	case cp0RegStatus:
		if sel == 0 {
			// Allow write to IE/EXL/ERL and IM bits; keep others as provided (approx)
			// Preserve reserved bits as-is.
			// Compose new status
			// We just replace fully, as our emulator doesn't rely on other impl-defined bits.
			old := c.status
			_ = old
			c.status = val
		}
	case cp0RegCause:
		if sel == 0 {
			// Writable bits: IV (23), SW pending IP1..IP0 ([9:8])
			// Keep TI and HW IPs and ExcCode/BD
			// Update IV
			if (val & causeIV) != 0 {
				c.cause |= causeIV
			} else {
				c.cause &^= causeIV
			}
			// Update SW IP0/IP1
			sw := (val >> causeIPShift) & 0x3
			c.cause &^= 0x3 << causeIPShift
			c.cause |= sw << causeIPShift
		}
	case cp0RegEPC:
		if sel == 0 {
			c.epc = val
		}
	case cp0RegPRId:
		switch sel {
		case 0:
			// PRId is RO; ignore writes
		case 1:
			// EBase writable partially; we'll accept full for simplicity
			c.ebase = val
		}
	case cp0RegConfig:
		switch sel {
		case 0:
			// Config0 partially writable; accept K0 (2:0); preserve M bit if set
			m := c.config0 & (1 << 31)
			c.config0 = m | (val & 0x7)
		case 1:
			// Config1 RO in our model
		}
	case cp0RegLLAddr:
		if sel == 0 {
			c.lladdr = val
		}
	case cp0RegWatchLo:
		if sel == 0 {
			c.watchLo = val
		}
	case cp0RegWatchHi:
		if sel == 0 {
			c.watchHi = val
		}
	case cp0RegXContext:
		if sel == 0 {
			c.xcontext = val
		}
	case cp0RegErrorEPC:
		if sel == 0 {
			c.errorepc = val
		}
	}
}

// Step should be called periodically (e.g., per instruction) to update Random.
func (c *COP0) Step() {
	if c.tlbSize <= 0 {
		return
	}
	if c.random == 0xFFFFFFFF {
		c.random = uint32(c.tlbSize - 1)
	}
	if c.random <= c.wired {
		c.random = uint32(c.tlbSize - 1)
	} else {
		c.random--
		if c.random < c.wired {
			c.random = uint32(c.tlbSize - 1)
		}
	}
}

// Tick adds cycles to Count and asserts timer interrupt when Count==Compare and Compare!=0.
func (c *COP0) Tick(cycles uint32) {
	prev := c.count
	c.count += cycles
	_ = prev
	if c.compare != 0 && c.count == c.compare {
		c.cause |= causeTI
		// Set IP7
		c.cause |= 1 << (causeIPShift + 7)
	}
}

// SetHWInterrupt sets pending state of a hardware interrupt line [2..6] (IP2..IP6).
func (c *COP0) SetHWInterrupt(line int, pending bool) {
	if line < 2 || line > 6 {
		return
	}
	bit := uint32(1) << (causeIPShift + uint(line))
	if pending {
		c.cause |= bit
	} else {
		c.cause &^= bit
	}
}

// SetSWInterrupt sets pending state of software interrupt 0 or 1 (IP0/IP1).
func (c *COP0) SetSWInterrupt(n int, pending bool) {
	if n < 0 || n > 1 {
		return
	}
	bit := uint32(1) << (causeIPShift + uint(n))
	if pending {
		c.cause |= bit
	} else {
		c.cause &^= bit
	}
}

// PendingInterrupt returns true if an interrupt should be taken.
func (c *COP0) PendingInterrupt() bool {
	// Enabled if IE=1, EXL=0, ERL=0 and (IP & IM) != 0
	if (c.status&statusIE) == 0 || (c.status&(statusEXL|statusERL)) != 0 {
		return false
	}
	ip := (c.cause >> causeIPShift) & 0xFF
	im := (c.status >> statusIMShift) & 0xFF
	return (ip & im) != 0
}

// RaiseException sets Cause.ExcCode, EPC/BD, sets EXL, and returns the exception vector address.
// If inDelaySlot is true, EPC gets pc-4 and BD=1.
func (c *COP0) RaiseException(excCode uint8, pc uint32, inDelaySlot bool) uint32 {
	// Set ExcCode in Cause [6:2]
	c.cause &^= 0x7C
	c.cause |= uint32(excCode&0x1F) << 2

	if inDelaySlot {
		c.cause |= causeBD
		c.epc = pc - 4
	} else {
		c.cause &^= causeBD
		c.epc = pc
	}

	// Set EXL
	c.status |= statusEXL

	// Choose vector: if IV=1 and excCode==Interrupt, use special offset 0x200
	// Use BEV=1 path by default (boot vectors at 0xBFC0_0180 / 0xBFC0_0200)
	baseBEV := uint32(0xBFC00000)
	vec := baseBEV + 0x180
	if excCode == excInt && (c.cause&causeIV) != 0 {
		vec = baseBEV + 0x200
	}

	// If user wants normal vectors (BEV=0), assume 0x8000_0180/0x8000_0200 via EBase
	// We'll infer BEV from whether ebase has high bit set and not in boot ROM area.
	// For simplicity, if ebase >= 0x80000000 and < 0xBFC00000, use it as base.
	if c.ebase >= 0x80000000 && c.ebase < 0xBFC00000 {
		vec = c.ebase + 0x180
		if excCode == excInt && (c.cause&causeIV) != 0 {
			vec = c.ebase + 0x200
		}
	}

	return vec
}

// ERET returns the next PC and clears EXL/ERL accordingly. Also clears BD.
func (c *COP0) ERET() uint32 {
	c.cause &^= causeBD
	if (c.status & statusERL) != 0 {
		c.status &^= statusERL
		return c.errorepc
	}
	c.status &^= statusEXL
	return c.epc
}

// TLBP probes the TLB for a matching entry based on EntryHi (VPN2, ASID).
// On hit, sets Index to entry index (clears P). On miss, sets Index.P=1.
func (c *COP0) TLBP() {
	vpn2 := c.entryHi & 0xFFFFE000
	asid := uint8(c.entryHi & 0xFF)

	for i := 0; i < c.tlbSize; i++ {
		e := &c.tlb[i]
		if e.VPN2 == vpn2 && (e.G || e.ASID == asid) {
			c.index = uint32(i) & 0x3F
			return
		}
	}
	c.index = 0x80000000 // P=1
}

// TLBR reads the TLB entry at Index into EntryHi/EntryLo0/EntryLo1/PageMask.
func (c *COP0) TLBR() {
	if (c.index & 0x80000000) != 0 {
		return // probe fail; do nothing
	}
	idx := int(c.index & 0x3F)
	if idx < 0 || idx >= c.tlbSize {
		return
	}
	e := &c.tlb[idx]

	// EntryHi: VPN2 and ASID
	c.entryHi = (e.VPN2 & 0xFFFFE000) | uint32(e.ASID)

	// PageMask
	c.pageMask = e.Mask & 0x01FFE000

	// EntryLo0
	lo0 := (e.PFN0 & 0xFFFFFC0) // PFN[25:6]
	lo0 |= uint32(e.C0&0x7) << 3
	if e.D0 {
		lo0 |= 1 << 2
	}
	if e.V0 {
		lo0 |= 1 << 1
	}
	if e.G {
		lo0 |= 1 << 0
	}
	c.entryLo0 = lo0

	// EntryLo1
	lo1 := e.PFN1 & 0xFFFFFC0
	lo1 |= uint32(e.C1&0x7) << 3
	if e.D1 {
		lo1 |= 1 << 2
	}
	if e.V1 {
		lo1 |= 1 << 1
	}
	if e.G {
		lo1 |= 1 << 0
	}
	c.entryLo1 = lo1
}

// TLBWI writes the TLB entry from EntryHi/EntryLo0/EntryLo1/PageMask into TLB[Index].
func (c *COP0) TLBWI() {
	if (c.index & 0x80000000) != 0 {
		return // invalid index from probe fail
	}
	idx := int(c.index & 0x3F)
	if idx < 0 || idx >= c.tlbSize {
		return
	}
	c.writeTLBAt(idx)
}

// TLBWR writes the entry into TLB[Random], then updates Random window.
func (c *COP0) TLBWR() {
	if c.tlbSize <= 0 {
		return
	}
	idx := int(c.random & 0x3F)
	if idx < int(c.wired) {
		idx = int(c.wired)
	}
	if idx >= c.tlbSize {
		idx = c.tlbSize - 1
	}
	c.writeTLBAt(idx)
	c.Step()
}

func (c *COP0) writeTLBAt(idx int) {
	e := &c.tlb[idx]

	e.VPN2 = c.entryHi & 0xFFFFE000
	e.ASID = uint8(c.entryHi & 0xFF)

	e.Mask = c.pageMask & 0x01FFE000

	// Decode EntryLo0
	lo0 := c.entryLo0 & 0x3FFFFFFF
	e.PFN0 = (lo0 >> 6) & 0xFFFFFC0
	e.C0 = uint8((lo0 >> 3) & 0x7)
	e.D0 = (lo0 & (1 << 2)) != 0
	e.V0 = (lo0 & (1 << 1)) != 0

	// Decode EntryLo1
	lo1 := c.entryLo1 & 0x3FFFFFFF
	e.PFN1 = (lo1 >> 6) & 0xFFFFFC0
	e.C1 = uint8((lo1 >> 3) & 0x7)
	e.D1 = (lo1 & (1 << 2)) != 0
	e.V1 = (lo1 & (1 << 1)) != 0

	// Global bit: effective G is G0&G1; we store as single boolean final (approx OK)
	e.G = (lo0&1) != 0 && (lo1&1) != 0
}

// Status returns the raw Status register value.
func (c *COP0) Status() uint32 { return c.status }

// Cause returns the raw Cause register value.
func (c *COP0) Cause() uint32 { return c.cause }

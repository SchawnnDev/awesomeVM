SRC = basic.go

BIN = basic_mips32
ASM = $(BIN).s

GOOS = linux
GOARCH = mips      # use 'mipsle' for little-endian

BUILD = GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN) $(SRC)
ASM_BUILD = GOOS=$(GOOS) GOARCH=$(GOARCH) go tool compile -S -N -l -o $(BIN).o $(SRC) > $(ASM)

.PHONY: all clean

all: $(BIN)

$(BIN):
	$(BUILD)

asm: $(ASM)

$(ASM):
	$(ASM_BUILD)

elf: $(BIN)
	readelf -x .rodata basic_mips32
	readelf -S basic_mips32

clean:
	rm -f $(BIN) $(BIN).o $(ASM)
package main

import (
	"awesomeVM/internal/mips"
	"flag"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// parse flags
	verbose := flag.Bool("v", false, "enable verbose logging")
	memoryFlag := flag.Uint64("memory", 1<<20, "memory size in bytes (max 4294967295)")
	flag.Parse()

	printIfVerbose(*verbose, "Starting MIPS VM...")

	// validate memory fits in uint32
	if *memoryFlag > uint64(math.MaxUint32) {
		log.Fatalf("memory size %d exceeds max uint32 %d", *memoryFlag, math.MaxUint32)
	}

	definedMemory := uint32(*memoryFlag)

	printIfVerbose(*verbose, "Allocating %d bytes of memory...", definedMemory)
	memory := mips.NewMemory(definedMemory)

	printIfVerbose(*verbose, "Starting CPU...")
	cpu := mips.NewCPU(memory)

	// create a channel to wait for CPU to stop
	done := make(chan struct{})

	printIfVerbose(*verbose, "Running CPU...")
	start := time.Now()

	// run the CPU in a goroutine so we can handle signals
	go func() {
		cpu.Run()
		close(done)
	}()

	// set up signal handling for Ctrl+C (os.Interrupt) and SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// wait for either the CPU to finish or a signal
	select {
	case <-sigCh:
		printIfVerbose(*verbose, "Signal received, stopping CPU...")
		cpu.Stop()
	case <-done:
		// CPU finished on its own
	}

	elapsed := time.Since(start)

	printIfVerbose(*verbose, "CPU stopped.")

	printIfVerbose(*verbose, "Total execution time: %s", elapsed)
}

// printIfVerbose prints a formatted message if verbose is true.
func printIfVerbose(verbose bool, format string, v ...interface{}) {
	if verbose {
		log.Printf(format, v...)
	}
}

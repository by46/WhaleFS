package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func SetupProfiling(cpuProfile, memProfile string) {
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}

		OnInterrupt(func() {
			pprof.StopCPUProfile()
		})
	}

	if memProfile != "" {
		runtime.MemProfileRate = 1
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal(err)
		}
		OnInterrupt(func() {
			_ = pprof.WriteHeapProfile(f)
		})
	}
}

func Time() {
	fmt.Printf("time: %d\n", time.Now().UnixNano()/1000000)
}

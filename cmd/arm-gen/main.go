package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
)

func decode(insn uint32) int16 {
	for j, fn := range funcs {
		if fn(insn) {
			return int16(j)
		}
	}
	return -1
}

func main() {
	out := flag.String("o", "arm64.dat", "otput file")

	flag.Parse()

	f, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	// max := uint64(65536)
	max := uint64(^uint32(0)) + 1
	vals := make([]int16, max)
	n := runtime.NumCPU()
	chunksz := max / uint64(n)
	wg := sync.WaitGroup{}
	for t := 0; t < n; t++ {
		wg.Add(1)
		go func(tid uint64) {
			start := tid * chunksz
			end := tid*chunksz + chunksz
			for i := start; i < end; i++ {
				insn := uint32(i)
				if tid == 0 && i%200000 == 0 {
					fmt.Printf("%.1f\n", float64(i)/float64(end)*100.0)
				}
				vals[i] = decode(insn)
			}
			wg.Done()
		}(uint64(t))
	}
	wg.Wait()
	err = binary.Write(f, binary.LittleEndian, vals)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
}

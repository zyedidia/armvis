package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/zyedidia/armvis/mra"
)

// map format is
// <32-bit value> (<class>:<subclass>:<count> )*

type line struct {
	class    uint16
	subclass uint16
}

func main() {
	armdat := flag.String("data", "arm64.dat", "full arm64 data file")
	armjson := flag.String("json", "arm64.json", "arm64 JSON encodings file")
	flag.Parse()
	jsondat, err := os.ReadFile(*armjson)
	if err != nil {
		log.Fatal(err)
	}
	var records []mra.Record
	err = json.Unmarshal(jsondat, &records)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(*armdat)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("loading...")
	b := make([]int16, 4*1024*1024*1024)
	err = binary.Read(f, binary.LittleEndian, b)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mapping...")
	classes := make([]line, mra.InstrNumIDs+1)
	n := 256
	for i := 0; i < len(b); i += n {
		if i%10000000 == 0 {
			log.Printf("%.3f\n", float64(i)/float64(len(b))*100.0)
		}
		for i := range classes {
			classes[i] = 0
		}
		one := false
		for j := 0; j < 256; j++ {
			if b[i+j] != -1 {
				r := records[b[i+j]]
				classes[mra.ClassToID(r.InstrClass)]++
				one = true
			}
		}
		if one {
			fmt.Printf("%d ", uint32(i))
			for i, c := range classes {
				if c != 0 {
					fmt.Printf("%d:%d:%d ", i, i, c)
				}
			}
			fmt.Println()
		}
	}
}

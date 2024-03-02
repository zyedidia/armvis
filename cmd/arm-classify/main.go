package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zyedidia/armnalyzer/mra"
)

func main() {
	base := flag.Bool("base", false, "only consider instructions from the ARMv8.0 instruction set")
	classes := flag.String("classes", "all", "comma-separated list of instruction classes")
	variant := flag.String("variant", "", "ISA version")

	flag.Parse()
	args := flag.Args()

	var allrecords []mra.Record

	filepath.WalkDir(args[0], func(path string, insn fs.DirEntry, err error) error {
		if insn != nil && !insn.IsDir() && strings.HasSuffix(path, ".xml") {
			data, err := os.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			var insn mra.InsnSection
			if err := xml.Unmarshal(data, &insn); err != nil {
				return nil
			}

			if insn.Type == "instruction" {
				if *base && !insn.BaseVariant() {
					return nil
				}
				hasclass := false
				for _, class := range strings.Split(*classes, ",") {
					if insn.HasClass(class) {
						hasclass = true
					}
				}
				if !hasclass {
					return nil
				}
				if *variant != "" && !insn.VariantLE(*variant) {
					return nil
				}

				allrecords = append(allrecords, mra.NewRecords(filepath.Base(path), insn)...)
			}
		}
		return nil
	})

	b, err := json.Marshal(allrecords)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

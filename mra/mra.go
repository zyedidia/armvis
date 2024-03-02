package mra

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	InstrGeneral   = "general"
	InstrSystem    = "system"
	InstrFloat     = "float"
	InstrFpSimd    = "fpsimd"
	InstrAdvSimd   = "advsimd"
	InstrSve       = "sve"
	InstrSve2      = "sve2"
	InstrMortlach  = "mortlach"
	InstrMortlach2 = "mortlach2"
)

const (
	InstrGeneralID uint8 = iota
	InstrSystemID
	InstrFloatID
	InstrFpSimdID
	InstrAdvSimdID
	InstrSveID
	InstrSve2ID
	InstrMortlachID
	InstrMortlach2ID
	InstrNumIDs
)

func ClassToID(class string) uint8 {
	switch class {
	case InstrGeneral:
		return InstrGeneralID
	case InstrSystem:
		return InstrSystemID
	case InstrFloat:
		return InstrFloatID
	case InstrFpSimd:
		return InstrFpSimdID
	case InstrAdvSimd:
		return InstrAdvSimdID
	case InstrSve:
		return InstrSveID
	case InstrSve2:
		return InstrSve2ID
	case InstrMortlach:
		return InstrMortlachID
	case InstrMortlach2:
		return InstrMortlach2ID
	}
	return InstrNumIDs
}

var InstrBase = strings.Join([]string{
	InstrGeneral,
	InstrFloat,
	InstrFpSimd,
	InstrAdvSimd,
}, ",")

type DocVar struct {
	XMLName xml.Name `xml:"docvar"`
	Key     string   `xml:"key,attr"`
	Val     string   `xml:"value,attr"`
}

type DocVars struct {
	XMLName xml.Name `xml:"docvars"`
	Vars    []DocVar `xml:"docvar"`
}

func (d DocVars) Mnemonic() string {
	for _, v := range d.Vars {
		if v.Key == "mnemonic" {
			return v.Val
		}
	}
	return ""
}

func (d DocVars) InstrClass() string {
	for _, v := range d.Vars {
		if v.Key == "instr-class" {
			return v.Val
		}
	}
	return ""
}

type PsText struct {
	XMLName xml.Name `xml:"pstext"`
	Content string   `xml:",innerxml"`
}

type Ps struct {
	XMLName xml.Name `xml:"ps"`
	Name    string   `xml:"name,attr"`
	PsText  PsText   `xml:"pstext"`
}

type PsSection struct {
	XMLName xml.Name `xml:"ps_section"`
	Ps      Ps
}

type ArchVariant struct {
	XMLName xml.Name `xml:"arch_variant"`
	Name    string   `xml:"name,attr"`
	Feature string   `xml:"feature,attr"`
}

type ArchVariants struct {
	XMLName  xml.Name      `xml:"arch_variants"`
	Variants []ArchVariant `xml:"arch_variant"`
}

func (a ArchVariants) GetVariants() []string {
	var vars []string
	for _, v := range a.Variants {
		vars = append(vars, v.Name)
	}
	return vars
}

func (a ArchVariants) GetFeatures() []string {
	var feats []string
	for _, v := range a.Variants {
		feats = append(feats, v.Feature)
	}
	return feats
}

type BitC struct {
	XMLName xml.Name `xml:"c"`
	Cols    int      `xml:"colspan,attr"`
	Value   string   `xml:",chardata"`
}

type Box struct {
	XMLName    xml.Name `xml:"box"`
	Bits       []BitC   `xml:"c"`
	Name       string   `xml:"name,attr"`
	Constraint string   `xml:"constraint,attr"`
}

type RegDiagram struct {
	XMLName xml.Name `xml:"regdiagram"`
	Name    string   `xml:"psname,attr"`
	Boxes   []Box    `xml:"box"`
}

func (r RegDiagram) String() string {
	b := &bytes.Buffer{}
	for i, box := range r.Boxes {
		if box.Name != "" {
			fmt.Fprint(b, box.Name)
		}
		if box.Constraint != "" {
			fmt.Fprint(b, strings.ReplaceAll(box.Constraint, " ", ""), "|")
			continue
		} else if box.Name != "" {
			fmt.Fprint(b, "=")
		}
		for _, bit := range box.Bits {
			if bit.Value == "" {
				bit.Value = "x"
			}
			if bit.Cols == 0 {
				bit.Cols = 1
			}
			bit.Value = strings.ReplaceAll(bit.Value, "(1)", "1")
			bit.Value = strings.ReplaceAll(bit.Value, "(0)", "0")
			fmt.Fprint(b, strings.Repeat(bit.Value, bit.Cols))
		}
		if i != len(r.Boxes)-1 {
			fmt.Fprint(b, "|")
		}
	}
	return b.String()
}

type Encoding struct {
	XMLName xml.Name `xml:"encoding"`
	Name    string   `xml:"name,attr"`
	Docs    DocVars  `xml:"docvars"`
}

type IClass struct {
	XMLName      xml.Name     `xml:"iclass"`
	Name         string       `xml:"name,attr"`
	Id           string       `xml:"id,attr"`
	RegDiagram   RegDiagram   `xml:"regdiagram"`
	ArchVariants ArchVariants `xml:"arch_variants"`
	Code         PsSection    `xml:"ps_section"`
	Encodings    []Encoding   `xml:"encoding"`
	Docs         DocVars      `xml:"docvars"`
}

type Classes struct {
	XMLName xml.Name `xml:"classes"`
	IClass  []IClass `xml:"iclass"`
}

type InsnSection struct {
	XMLName xml.Name  `xml:"instructionsection"`
	Docs    DocVars   `xml:"docvars"`
	Type    string    `xml:"type,attr"`
	Id      string    `xml:"id,attr"`
	Classes Classes   `xml:"classes"`
	Code    PsSection `xml:"ps_section"`
}

type Assign struct {
	Dest string
	Src  string
}

func getAssigns(code string) []Assign {
	lines := strings.Split(code, "\n")
	var assigns []Assign
	for _, l := range lines {
		before, after, found := strings.Cut(l, " = ")
		if !found {
			continue
		}
		assigns = append(assigns, Assign{
			Dest: before,
			Src:  after,
		})
	}
	return assigns
}

var linkrx = regexp.MustCompile(`<a.*?>`)

func (is InsnSection) ReadSet() []string {
	lines := strings.Split(is.Code.Ps.PsText.Content, "\n")
	var set []string
	for _, l := range lines {
		if strings.Contains(l, "impl-aarch64.X.read.2") {
			l = strings.ReplaceAll(l, "</a>", "")
			l = linkrx.ReplaceAllLiteralString(l, "")
			_, after, found := strings.Cut(l, "=")
			if !found {
				log.Fatal("no = in ", l)
			}
			set = append(set, strings.TrimSpace(after))
		}
	}
	return set
}

func (is InsnSection) Uses(op string) bool {
	lines := strings.Split(is.Code.Ps.PsText.Content, "\n")
	for _, l := range lines {
		if strings.Contains(l, op) {
			return true
		}
	}
	return false
}

func (is InsnSection) UsesPc() bool {
	return is.Uses("impl-aarch64.PC.read.0")
}

func (is InsnSection) ReadsMem() bool {
	return is.Uses("impl-aarch64.Mem.read.3")
}

func (is InsnSection) WritesMem() bool {
	return is.Uses("impl-aarch64.Mem.write.3")
}

func (is InsnSection) MemAtomic() bool {
	return is.Uses("impl-aarch64.MemAtomic.4")
}

func (is InsnSection) IsBranch() bool {
	return is.Uses("impl-shared.BranchTo.3")
}

func (is InsnSection) BaseVariant() bool {
	for _, c := range is.Classes.IClass {
		if len(c.ArchVariants.Variants) != 0 {
			return false
		}
	}
	return true
}

func (is InsnSection) Variant(version string) bool {
	for _, c := range is.Classes.IClass {
		for _, v := range c.ArchVariants.Variants {
			if v.Name == version {
				return true
			}
		}
	}
	return false
}

func (is InsnSection) VariantLE(version string) bool {
	for _, c := range is.Classes.IClass {
		if len(c.ArchVariants.Variants) == 0 {
			return true
		}
		for _, v := range c.ArchVariants.Variants {
			if strings.Compare(v.Name, version) <= 0 {
				return true
			}
		}
	}
	return false
}

func (is InsnSection) HasClass(class string) bool {
	if class == "all" {
		return true
	}
	for _, c := range is.Classes.IClass {
		if c.Docs.InstrClass() == class {
			return true
		}
	}
	if is.Docs.InstrClass() != "" && is.Docs.InstrClass() == class {
		return true
	}
	return false
}

func (is InsnSection) GetClasses() []string {
	m := make(map[string]bool)
	for _, c := range is.Classes.IClass {
		if c.Docs.InstrClass() != "" {
			m[c.Docs.InstrClass()] = true
		}
	}
	if is.Docs.InstrClass() != "" {
		m[is.Docs.InstrClass()] = true
	}
	var classes []string
	for k := range m {
		classes = append(classes, k)
	}
	return classes
}

func (is InsnSection) BaseArch() bool {
	if !strings.Contains(InstrBase, is.Docs.InstrClass()) {
		return false
	}
	return is.BaseVariant()
}

func (is InsnSection) MatchesArch(name string, feature string) bool {
	for _, c := range is.Classes.IClass {
		for _, v := range c.ArchVariants.Variants {
			if strings.Contains(v.Name, name) && strings.Contains(v.Feature, feature) {
				return true
			}
		}
	}
	return false
}

func (is InsnSection) WriteSet() []string {
	lines := strings.Split(is.Code.Ps.PsText.Content, "\n")
	set := make(map[string]bool)
	for _, l := range lines {
		if strings.Contains(l, "impl-aarch64.X.write.2") {
			l = strings.ReplaceAll(l, "</a>", "")
			l = linkrx.ReplaceAllLiteralString(l, "")
			before, _, found := strings.Cut(l, "=")
			if !found {
				log.Fatal("no = in ", l)
			}
			set[strings.TrimSpace(before)] = true
		}
	}
	var ret []string
	for k := range set {
		ret = append(ret, k)
	}
	return ret
}

func (is InsnSection) Names() []string {
	set := make(map[string]bool)
	for _, c := range is.Classes.IClass {
		for _, e := range c.Encodings {
			set[e.Docs.Mnemonic()] = true
		}
	}
	var names []string
	for k := range set {
		names = append(names, k)
	}
	return names
}

type Record struct {
	File       string
	Name       string
	IClass     string
	Path       string
	Variants   string
	Features   string
	InstrClass string
	RegDiagram string
}

func NewRecords(file string, insn InsnSection) []Record {
	var records []Record
	for _, c := range insn.Classes.IClass {
		set := make(map[string]bool)
		for _, e := range c.Encodings {
			set[e.Docs.Mnemonic()] = true
		}
		var names []string
		for k := range set {
			names = append(names, k)
		}
		records = append(records, Record{
			File:       file,
			Name:       strings.Join(names, ";"),
			IClass:     c.Id,
			Path:       c.RegDiagram.Name,
			Variants:   strings.Join(c.ArchVariants.GetVariants(), ";"),
			Features:   strings.Join(c.ArchVariants.GetFeatures(), ";"),
			InstrClass: c.Docs.InstrClass(),
			RegDiagram: c.RegDiagram.String(),
		})
	}
	return records
}

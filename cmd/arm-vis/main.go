package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"

	tc "github.com/wayneashleyberry/truecolor/pkg/color"
	"github.com/zyedidia/armvis/mra"
)

type Palette struct {
	colors map[uint8]color.RGBA
	bg     color.RGBA
}

func Hex(hex string) color.RGBA {
	values, err := strconv.ParseUint(string(hex[1:]), 16, 32)
	if err != nil {
		panic(err)
	}
	return color.RGBA{R: uint8(values >> 16), G: uint8((values >> 8) & 0xFF), B: uint8(values & 0xFF), A: 255}
}

var solarized = Palette{
	bg: color.RGBA{0, 43, 54, 0xff},
	colors: map[uint8]color.RGBA{
		mra.InstrGeneralID:   {181, 137, 0, 0xff},
		mra.InstrSystemID:    {203, 75, 22, 0xff},
		mra.InstrFloatID:     {220, 50, 47, 0xff},
		mra.InstrFpSimdID:    {211, 54, 130, 0xff},
		mra.InstrAdvSimdID:   {108, 113, 196, 0xff},
		mra.InstrSveID:       {38, 139, 210, 0xff},
		mra.InstrSve2ID:      {42, 161, 152, 0xff},
		mra.InstrMortlachID:  {133, 153, 0, 0xff},
		mra.InstrMortlach2ID: {253, 246, 227, 0xff},
		mra.InstrNumIDs:      {131, 148, 150, 0xff},
	},
}

var monokai = Palette{
	bg: Hex("#272833"),
	colors: map[uint8]color.RGBA{
		mra.InstrGeneralID:   Hex("#66d9ef"),
		mra.InstrSystemID:    Hex("#e6db74"),
		mra.InstrFloatID:     Hex("#fd971f"),
		mra.InstrFpSimdID:    Hex("#f92672"),
		mra.InstrAdvSimdID:   Hex("#fd5ff0"),
		mra.InstrSveID:       Hex("#ae81ff"),
		mra.InstrSve2ID:      Hex("#a1efe4"),
		mra.InstrMortlachID:  Hex("#a6e22e"),
		mra.InstrMortlach2ID: Hex("#f8f8f2"),
		mra.InstrNumIDs:      Hex("#75715e"),
	},
}

var themes = map[string]Palette{
	"monokai":   monokai,
	"solarized": solarized,
}

func main() {
	out := flag.String("o", "arm64.png", "output file")
	themename := flag.String("theme", "solarized", "color theme (solarized,monokai)")

	flag.Parse()
	args := flag.Args()
	if len(args) <= 0 {
		log.Fatal("no map input file")
	}
	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}

	width := 4096
	height := 4096
	hilbertOrder := 12

	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	bg := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})

	if _, ok := themes[*themename]; !ok {
		log.Fatalf("theme %s does not exist", *themename)
	}
	theme := themes[*themename]

	overlay := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})

	// Fill background
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			bg.Set(x, y, theme.bg)
		}
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		s, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		x, y := hilbertXY(uint32(s)/256, hilbertOrder)
		c := theme.bg
		for i, f := range fields[1:] {
			if f != "0" {
				c = theme.colors[uint8(i)]
				a, err := strconv.Atoi(f)
				if err != nil {
					log.Fatal(err)
				}
				c.A = uint8(a - 1)
				if c.A > 0 {
					_, _, _, a := overlay.At(int(x), int(y)).RGBA()
					if c.A > uint8(a) {
						overlay.Set(int(x), int(y), c)
					}
				}
				break
			}
		}
	}

	draw.Draw(img, img.Bounds(), bg, image.Point{0, 0}, draw.Src)
	draw.DrawMask(img, img.Bounds(), overlay, image.Point{0, 0}, overlay, image.Point{0, 0}, draw.Over)

	// Encode as PNG.
	of, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	err = png.Encode(of, img)
	if err != nil {
		log.Fatal(err)
	}

	for i := uint8(0); i < uint8(len(theme.colors)); i++ {
		c := theme.colors[i]
		tc.Black().Background(c.R, c.G, c.B).Print(mra.IDToClass(i))
		fmt.Print(" ")
	}
	fmt.Println()
}

package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
)

func main() {
	getBackground("IDR713")
	images, err := getRadarImages("IDR713", 6)
	if err != nil {
		log.Fatal(err)
	}
	for i, img := range images {
		f, err := os.Create(fmt.Sprintf("radar%d.png", i))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			log.Fatal(err)
		}
	}
}

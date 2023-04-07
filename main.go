package main

import (
	"image/gif"
	"log"
	"os"
)

func main() {
	g, err := getRadarGIF("IDR713")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("IDR713.gif")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := gif.EncodeAll(f, g); err != nil {
		log.Fatal(err)
	}
}

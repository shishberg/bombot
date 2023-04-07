package main

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path"
	"time"

	"github.com/andybons/gogif"
	_ "github.com/andybons/gogif"
	"github.com/jlaffaye/ftp"
)

const (
	ftpAddr      = "ftp.bom.gov.au:21"
	ftpRadarPath = "anon/gen/radar"
	ftpBgPath    = "anon/gen/radar_transparencies"
)

var (
	bgLayers = []string{"background", "topography", "locations", "range"}
)

func getRadarGIF(prefix string) (*gif.GIF, error) {
	bg, err := getBackground(prefix)
	if err != nil {
		return nil, err
	}
	radar, err := getRadarImages(prefix, 6)
	if err != nil {
		return nil, err
	}

	g := &gif.GIF{}
	q := &gogif.MedianCutQuantizer{NumColor: 256}
	last := radar[0]
	radar = append(radar[1:], last)
	for i, rImg := range radar {
		d := image.NewRGBA(bg.Bounds())
		draw.Draw(d, d.Rect, bg, bg.Bounds().Min, draw.Over)
		draw.Draw(d, d.Rect, rImg, rImg.Bounds().Min, draw.Over)
		frame := image.NewPaletted(rImg.Bounds(), palette.Plan9)
		q.Quantize(frame, frame.Rect, d, d.Bounds().Min)
		g.Image = append(g.Image, frame)
		if i == len(radar)-2 {
			g.Delay = append(g.Delay, 100)
		} else {
			g.Delay = append(g.Delay, 30)
		}
	}
	return g, nil
}

func withFTP(f func(c *ftp.ServerConn) error) error {
	c, err := ftp.Dial(ftpAddr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	defer c.Quit()
	if err := c.Login("anonymous", "anonymous"); err != nil {
		return err
	}
	return f(c)
}

func getRadarImages(prefix string, n int) (images []image.Image, err error) {
	glob := fmt.Sprintf("%s*.png", prefix)
	err = withFTP(func(c *ftp.ServerConn) error {
		entries, err := c.List(path.Join(ftpRadarPath, glob))
		if err != nil {
			return err
		}
		for _, e := range entries[len(entries)-n:] {
			img, err := readImage(c, path.Join(ftpRadarPath, e.Name))
			if err != nil {
				return err
			}
			images = append(images, img)
		}
		return nil
	})
	return
}

func getBackground(prefix string) (img image.Image, err error) {
	fn := fmt.Sprintf("%s.bg.png", prefix)
	if f, err := os.Open(fn); err == nil {
		defer f.Close()
		if img, _, err := image.Decode(f); err == nil {
			return img, err
		}
	}

	var drawImg draw.Image
	err = withFTP(func(c *ftp.ServerConn) error {
		for _, layer := range bgLayers {
			layerImg, err := readImage(c, path.Join(ftpBgPath, fmt.Sprintf("%s.%s.png", prefix, layer)))
			if err != nil {
				return err
			}
			if drawImg == nil {
				drawImg = image.NewRGBA(layerImg.Bounds())
			}
			draw.Draw(drawImg, drawImg.Bounds(), layerImg, layerImg.Bounds().Min, draw.Over)
		}
		return nil
	})

	if f, err := os.Create(fn); err == nil {
		defer f.Close()
		png.Encode(f, drawImg)
	}

	return drawImg, nil
}

func readImage(c *ftp.ServerConn, path string) (image.Image, error) {
	resp, err := c.Retr(path)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	img, _, err := image.Decode(resp)
	return img, err
}

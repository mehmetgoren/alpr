package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/golang/freetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"image"
	"image/color"
	"image/draw"
	jpg "image/jpeg"
	"log"
	"os"
)

var clr = color.RGBA{R: 255, B: 255, A: 255}

func addLabel(img *image.RGBA, x, y int, label string) {
	//point := fixed.Point26_6{fixed.I(x), fixed.I(y)}
	//f.Width = 600
	//f.Height = 1300

	//c := freetype.NewContext()
	//size := flag.Float64("size", 24, "font size in points")
	//spacing := flag.Float64("spacing", 2., "line spacing (e.g. 2 means double spaced)")
	//
	//c.SetFontSize(*size)
	//size := 36.0 // font size in pixels

	f := basicfont.Face7x13
	pt := freetype.Pt(x, y)
	//pt.Y += c.PointToFixed(*size * *spacing)

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(clr),
		Face: f,
		Dot:  pt,
	}
	d.DrawString(label)
}

func drawRectangle(img *image.RGBA, rect image.Rectangle) *image.RGBA {
	min := rect.Min
	max := rect.Max
	x0, y0, x1, y1 := min.X, min.Y, max.X, max.Y
	size := 3

	fy := func(i, y int) {
		for j := 0; j < size; j++ {
			img.Set(i, y-j, clr)
		}
		img.Set(i, y, clr)
		for j := 0; j < size; j++ {
			img.Set(i, y+j, clr)
		}
	}
	fx := func(i, x int) {
		for j := 0; j < size; j++ {
			img.Set(x-j, i, clr)
		}
		img.Set(x, i, clr)
		for j := 0; j < size; j++ {
			img.Set(x+j, i, clr)
		}
	}

	for i := x0; i < x1; i++ {
		fy(i, y0)
		fy(i, y1)
	}

	for i := y0; i <= y1; i++ {
		fx(i, x0)
		fx(i, x1)
	}

	return img
}

func getImageFromFilePath(filePath string) (*image.RGBA, error) {

	// read file
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// convert as image.Image
	orig, _, err := image.Decode(f)

	// convert as usable image
	b := orig.Bounds()
	img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(img, img.Bounds(), orig, b.Min, draw.Src)

	return img, err
}

func DrawRect(imgFilename string, x0 int, y0 int, x1 int, y1 int, label string, conf float64) (*image.RGBA, error) {
	// read file and convert it
	src, err := getImageFromFilePath(imgFilename)
	if err != nil {
		log.Println("an error occurred while getting image from disk")
		return nil, err
	}

	myRectangle := image.Rect(x0, y0, x1, y1)
	addLabel(src, x0, y0-5, label+fmt.Sprintf(" conf: %f", conf))

	dst := drawRectangle(src, myRectangle)

	return dst, nil

	//outputFile, err := os.Create("./sil/" + uuid.New().String() + ".jpg")
	//if err != nil {
	//	log.Println("an error occurred while getting image from disk")
	//	return nil, err
	//}

	//jpg.Encode(outputFile, dst, &jpg.Options{Quality: 100})

	//outputFile.Close()
}

func ImageToBase64(img *image.RGBA) (string, error) {
	buf := new(bytes.Buffer)
	err := jpg.Encode(buf, img, &jpg.Options{Quality: 100})
	if err != nil {
		log.Println("an error occurred on ImageToBase64, err: ", err.Error())
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func OverwriteImage(img *image.RGBA, imgFilename string) error {
	outputFile, err := os.Create(imgFilename)
	if err != nil {
		log.Println("an error occurred while overwriting image from disk")
		return err
	}
	defer outputFile.Close()

	jpg.Encode(outputFile, img, &jpg.Options{Quality: 100})
	return nil
}

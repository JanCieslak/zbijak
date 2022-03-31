package main

import (
	"bufio"
	"github.com/JanCieslak/zbijak/common"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"log"
	"os"
)

func loadImage(path string, alpha float64) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(bufio.NewReader(f))
	if err != nil {
		log.Fatal(err)
	}
	origEbitenImage := ebiten.NewImageFromImage(img)

	w, h := origEbitenImage.Size()
	ebitenImage := ebiten.NewImage(w, h)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(1, 1, 1, alpha)
	ebitenImage.DrawImage(origEbitenImage, op)

	return ebitenImage
}

func drawCircle(screen *ebiten.Image, x, y, s float64, c color.Color) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(x, y)
	r, g, b, _ := c.RGBA()
	rf := common.MapValue(float64(r), 0, 0xffff, 0, 1)
	gf := common.MapValue(float64(g), 0, 0xffff, 0, 1)
	bf := common.MapValue(float64(b), 0, 0xffff, 0, 1)
	op.ColorM.Scale(rf, gf, bf, 1)
	screen.DrawImage(circleImage, &op)
}

func drawCircleOutline(screen *ebiten.Image, x, y, s float64) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(x, y)
	screen.DrawImage(circleOutlineImage, &op)
}

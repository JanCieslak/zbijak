package utils

import (
	"bufio"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	_ "image/png"
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

package main

import (
	"bytes"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	_ "image/png"
	"io/ioutil"
	"log"
)

type Player struct {
	x     float64
	y     float64
	image *ebiten.Image
}

type Game struct {
	player Player
}

func (g *Game) Update() error {
	speed := 10.0

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.x -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.x += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.y -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.y += speed
	}

	fmt.Println("Fps: ", ebiten.CurrentFPS())
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.player.x, g.player.y)
	op.GeoM.Scale(0.1, 0.1)
	screen.DrawImage(g.player.image, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1366, 768
}

func main() {
	game := &Game{
		player: Player{
			x:     250,
			y:     250,
			image: imageFromFilename("resources/gopher.png"),
		},
	}

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(1240, 768)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(144)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalln(err)
	}
}

// TODO Check utils from ebiten
func imageFromFilename(filename string) *ebiten.Image {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
	}

	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {
		log.Fatalln(err)
	}

	return ebiten.NewImageFromImage(img)
}

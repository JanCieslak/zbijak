package utils

import (
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/inconsolata"
	"image/color"
)

var (
	circleOutlineImage = loadImage("resources/circle.png", 0.2)
	circleImage        = loadImage("resources/filled_circle.png", 1.0)

	OrangeTeamColor = color.RGBA{R: 235, G: 131, B: 52, A: 255}
	BlueTeamColor   = color.RGBA{R: 52, G: 158, B: 235, A: 255}

	face = inconsolata.Bold8x16
)

func GetTeamColor(team constants.Team) color.Color {
	if team == constants.TeamOrange {
		return OrangeTeamColor
	} else if team == constants.TeamBlue {
		return BlueTeamColor
	}
	return color.White
}

func DrawCircle(screen *ebiten.Image, cx, cy, radius, scale float64, color color.Color) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx-radius, cy-radius)

	r, g, b, _ := color.RGBA()
	rf := MapValue(float64(r), 0, 0xffff, 0, 1)
	gf := MapValue(float64(g), 0, 0xffff, 0, 1)
	bf := MapValue(float64(b), 0, 0xffff, 0, 1)
	op.ColorM.Scale(rf, gf, bf, 1)

	screen.DrawImage(circleImage, &op)
}

func DrawCircleOutline(screen *ebiten.Image, cx, cy, radius, scale float64) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx-radius, cy-radius)

	screen.DrawImage(circleOutlineImage, &op)
}

func DrawText(screen *ebiten.Image, message string, cx, cy float64) {
	textBounds := text.BoundString(face, message)

	text.Draw(screen, message, face, int(cx)-textBounds.Dx()/2, int(cy)-textBounds.Dy()/2, color.White)
}

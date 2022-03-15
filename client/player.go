package main

import (
	"github.com/JanCieslak/zbijak/common/vector"
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	pos   vector.Vec2
	state State
}

func (p *Player) Update() {
	moveVector := calcMoveVector()
	p.state.Update(p, moveVector)
}

func calcMoveVector() vector.Vec2 {
	moveVector := vector.Vec2{}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		moveVector.Add(-1, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		moveVector.Add(1, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		moveVector.Add(0, -1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		moveVector.Add(0, 1)
	}

	return moveVector
}

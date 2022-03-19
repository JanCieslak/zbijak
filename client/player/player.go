package player

import (
	"github.com/JanCieslak/zbijak/common/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"time"
)

const (
	NormalSpeed = 2.5

	FullChargeSpeed    = 0.2 * NormalSpeed
	FullChargeDuration = time.Second

	DashSpeed    = 2 * NormalSpeed
	DashDuration = 250 * time.Millisecond
	DashCooldown = time.Second
)

type Player struct {
	Pos           vector.Vec2
	Velocity      vector.Vec2
	MovementState State
	PlayerState   State
}

func NewPlayer() *Player {
	return &Player{
		Pos:           vector.Vec2{X: 250, Y: 250},
		Velocity:      vector.Vec2{},
		MovementState: NormalMovementState{},
		PlayerState:   NormalPlayerState{},
	}
}

func (p *Player) Update() {
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

	moveVector.Normalize()
	p.Velocity = moveVector

	p.MovementState.Update(p)
	p.PlayerState.Update(p)

	p.Pos.AddVec(p.Velocity)
}

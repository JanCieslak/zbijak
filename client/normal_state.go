package main

import (
	"github.com/JanCieslak/zbijak/common/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"time"
)

type NormalState struct {
	lastDashTime time.Time
}

func (s NormalState) Update(player *Player, moveVector vector.Vec2) {
	if time.Since(s.lastDashTime) > dashCooldown && ebiten.IsKeyPressed(ebiten.KeySpace) {
		moveVector.Normalize()
		moveVector.Mul(dashSpeed)

		player.state = DashState{
			startTime:  time.Now(),
			dashVector: moveVector,
		}
	}

	moveVector.Normalize()
	moveVector.Mul(speed)
	player.pos.AddVec(moveVector)
}

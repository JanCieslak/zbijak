package player

import (
	"github.com/hajimehoshi/ebiten/v2"
	"time"
)

type NormalMovementState struct {
	lastDashTime time.Time
}

func (s NormalMovementState) Update(p *Player) {
	if time.Since(s.lastDashTime) > DashCooldown && ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.Velocity.Mul(DashSpeed)

		p.MovementState = DashMovementState{
			startTime:  time.Now(),
			dashVector: p.Velocity,
		}
	}

	p.Velocity.Mul(NormalSpeed)
}

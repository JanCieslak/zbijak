package player

import (
	"github.com/JanCieslak/zbijak/common/vector"
	"time"
)

type DashMovementState struct {
	startTime  time.Time
	dashVector vector.Vec2
}

func (s DashMovementState) Update(p *Player) {
	if time.Since(s.startTime) > DashDuration {
		p.MovementState = NormalMovementState{
			lastDashTime: time.Now(),
		}
	}

	p.Velocity = s.dashVector
}

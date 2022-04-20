package player

import (
	"github.com/JanCieslak/zbijak/common/vec"
	"time"
)

type StopMovementState struct {
	startTime  time.Time
	dashVector vec.Vec2
}

func (s StopMovementState) Update(p *Player) {
	p.Velocity.Mul(0)
}

package player

import (
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"reflect"
	"time"
)

type ChargePlayerState struct {
	startTime time.Time
}

func (s ChargePlayerState) Update(p *Player) {
	if !ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		p.PlayerState = NormalPlayerState{}
	}

	if reflect.TypeOf(p.MovementState) == reflect.TypeOf(NormalMovementState{}) {
		interpolationFactor := float64(time.Since(s.startTime)) / float64(FullChargeDuration)

		p.Velocity.Normalize()
		if interpolationFactor < 1 {
			speed := packets.Lerp(NormalSpeed, FullChargeSpeed, interpolationFactor)
			log.Println("Interpolation factor", interpolationFactor, "Speed", speed)
			p.Velocity.Mul(speed)
		} else {
			p.Velocity.Mul(FullChargeSpeed)
		}
	}
}

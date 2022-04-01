package player

import (
	"github.com/JanCieslak/Zbijak/client/utils"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/hajimehoshi/ebiten/v2"
	"reflect"
	"time"
)

type ChargePlayerState struct {
	startTime time.Time
}

func (s ChargePlayerState) Update(p *Player) {
	if !ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		p.PlayerState = NormalPlayerState{}

		netman.SendUnreliable(netman.Fire, netman.FirePacketData{
			ClientId: p.Id,
		})
	}

	if reflect.TypeOf(p.MovementState) == reflect.TypeOf(NormalMovementState{}) {
		interpolationFactor := float64(time.Since(s.startTime)) / float64(FullChargeDuration)

		p.Velocity.Normalize()
		if interpolationFactor < 1 {
			speed := utils.Lerp(NormalSpeed, FullChargeSpeed, interpolationFactor)
			//log.Println("Interpolation factor", interpolationFactor, "Speed", speed)
			p.Velocity.Mul(speed)
		} else {
			p.Velocity.Mul(FullChargeSpeed)
		}
	}
}

package game

import (
	"fmt"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/hajimehoshi/ebiten/v2"
	"reflect"
	"time"
)

type ChargePlayerState struct {
	startTime time.Time
}

func (s ChargePlayerState) Update(g *Game, p *Player) {
	if !ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		p.PlayerState = NormalPlayerState{}

		packets.Send(g.Conn, packets.Fire, packets.FirePacketData{
			ClientId: g.Id,
		})
		fmt.Println("FIRE SENT")
	}

	if reflect.TypeOf(p.MovementState) == reflect.TypeOf(NormalMovementState{}) {
		interpolationFactor := float64(time.Since(s.startTime)) / float64(FullChargeDuration)

		p.Velocity.Normalize()
		if interpolationFactor < 1 {
			speed := packets.Lerp(NormalSpeed, FullChargeSpeed, interpolationFactor)
			//log.Println("Interpolation factor", interpolationFactor, "Speed", speed)
			p.Velocity.Mul(speed)
		} else {
			p.Velocity.Mul(FullChargeSpeed)
		}
	}
}

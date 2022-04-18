package player

import (
	"github.com/JanCieslak/Zbijak/client/utils"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"reflect"
	"time"
)

type ChargePlayerState struct {
	startTime time.Time
}

func (s ChargePlayerState) Update(p *Player) {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		elapsedTime := utils.Min(time.Since(s.startTime), FullChargeDuration)
		percentage := float64(elapsedTime) / float64(FullChargeDuration)
		multiplier := utils.Lerp(MinChargeMultiplier, FullChargeMultiplier, percentage)

		log.Println("Threw with multiplier:", multiplier)

		p.PlayerState = NormalPlayerState{}

		netman.SendUnreliable(netman.Fire, netman.FirePacketData{
			ClientId:   p.Id,
			Multiplier: multiplier,
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

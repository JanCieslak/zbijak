package player

import (
	"github.com/hajimehoshi/ebiten/v2"
	"time"
)

type NormalPlayerState struct{}

func (s NormalPlayerState) Update(p *Player) {
	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		p.PlayerState = ChargePlayerState{
			startTime: time.Now(),
		}
	}
}

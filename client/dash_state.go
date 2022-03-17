package main

import (
	"github.com/JanCieslak/zbijak/common/vector"
	"time"
)

const (
	dashSpeed    = 2 * Speed
	dashDuration = 250 * time.Millisecond
	dashCooldown = time.Second
)

type DashState struct {
	startTime  time.Time
	dashVector vector.Vec2
}

func (s DashState) Update(player *Player, moveVector vector.Vec2) {
	if time.Since(s.startTime) > dashDuration {
		player.State = NormalState{
			lastDashTime: time.Now(),
		}
	}

	player.Pos.AddVec(s.dashVector)
}

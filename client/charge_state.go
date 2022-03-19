package main

import (
	"github.com/JanCieslak/zbijak/common/vector"
)

type ChargeState struct {
	PrevState State
}

func (s ChargeState) Update(player *Player, moveVector vector.Vec2) {
	// TODO Use PrevState to update normal/dash state and apply slowly decreasing speed in normal state (but how to make it so while charging state one can switch between normal and dash state)
}

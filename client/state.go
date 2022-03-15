package main

import "github.com/JanCieslak/zbijak/common/vector"

type State interface {
	Update(player *Player, moveVector vector.Vec2)
}

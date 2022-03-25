package game

import "github.com/JanCieslak/zbijak/common/vec"

type Ball struct {
	Id      uint8
	OwnerId uint8
	Pos     vec.Vec2
}

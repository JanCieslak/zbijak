package constants

import "time"

const (
	ScreenWidth         = 640
	ScreenHeight        = 480
	TickRate            = 144
	InterpolationOffset = 50 * time.Millisecond
	PlayerRadius        = 16
	BallRadius          = 8
	BallOrbitRadius     = 40
	TickTime            = time.Second / TickRate
	NoGoZonePadding     = BallOrbitRadius + BallRadius
)

type Team uint8

const (
	TeamA Team = iota
	TeamB
	NoTeam = 255
)

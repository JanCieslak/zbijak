package constants

import "time"

const (
	ScreenWidth         = 640
	ScreenHeight        = 480
	TickRate            = 144
	InterpolationOffset = 100 * time.Millisecond
	PlayerRadius        = 16
	BallRadius          = 8
	BallOrbitRadius     = 40
	TickTime            = time.Second / TickRate
)

type Team uint8

const (
	TeamBlue Team = iota
	TeamOrange
	NoTeam = 255
)

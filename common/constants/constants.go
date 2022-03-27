package constants

const (
	ScreenWidth         = 640
	ScreenHeight        = 480
	TickRate            = 144
	InterpolationOffset = 100
	PlayerRadius        = 16
	BallRadius          = 8
)

type Team uint8

const (
	TeamBlue Team = iota
	TeamOrange
)

package constants

import "time"

const (
	ServerTickRate = 60
	ServerTickTime = time.Second / ServerTickRate
)

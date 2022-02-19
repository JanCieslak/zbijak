package common

type PacketType uint8

const (
	Hello PacketType = iota
	Welcome
)

package netman

import (
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/vec"
	"time"
)

type PacketKind uint8

const (
	Hello PacketKind = iota
	Welcome
	PlayerUpdate
	ServerUpdate
	Fire
	Bye
	ByeAck
)

type HelloPacketData struct{}

type WelcomePacketData struct {
	ClientId uint8
	Team     constants.Team
}

type PlayerUpdatePacketData struct {
	ClientId uint8
	Team     constants.Team
	Name     string
	Pos      vec.Vec2
	Rotation float64
	InDash   bool
}

type PlayerData struct {
	ClientId uint8
	Team     constants.Team
	Name     string
	Pos      vec.Vec2
	Rotation float64
	InDash   bool
}

type BallData struct {
	Id    uint8
	Owner uint8
	Pos   vec.Vec2
}

type ServerUpdatePacketData struct {
	PlayersData map[uint8]PlayerData
	Balls       []BallData
	Timestamp   time.Time
}

type FirePacketData struct {
	ClientId uint8
}

type ByePacketData struct {
	ClientId uint8
}

type ByeAckPacketData struct {
	ClientId uint8
}

type PacketData interface {
	HelloPacketData |
		WelcomePacketData |
		PlayerUpdatePacketData |
		ServerUpdatePacketData |
		FirePacketData |
		ByePacketData |
		ByeAckPacketData
}

type Packet[T PacketData] struct {
	Kind PacketKind
	Data T
}

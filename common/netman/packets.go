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

// TODO This should be only user input (user shouldn't be able to manipulate data that is sent to sever)
type PlayerUpdatePacketData struct {
	ClientId uint8
	Team     constants.Team
	Name     string
	Pos      vec.Vec2
	Rotation float64
	InDash   bool
}

type PlayerData struct {
	ClientId uint8 // TODO Is it needed ?
	Team     constants.Team
	Name     string
	Pos      vec.Vec2
	Rotation float64
	InDash   bool // TODO Is it needed ?
}

type BallData struct {
	Id    uint8 // TODO Is it needed ?
	Owner uint8 // TODO Is it needed ?
	Pos   vec.Vec2
}

type ServerUpdatePacketData struct {
	PlayersData map[uint8]PlayerData
	Balls       []BallData
	Timestamp   time.Time
}

type FirePacketData struct { // TODO should be deleted when using only player inputs
	ClientId   uint8
	Multiplier float64 // TODO This should be calculated from input updates
}

type ByePacketData struct {
	ClientId uint8 // TODO Should be needed
}

type ByeAckPacketData struct {
	ClientId uint8 // TODO this is fine but server inner identification should be other than client's (server prob should be using ip addresses and clients might user uint8s)
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

package packets

import (
	"github.com/JanCieslak/zbijak/common/vector"
	"time"
)

type PacketKind uint8

const (
	Hello PacketKind = iota
	Welcome
	PlayerUpdate
	ServerUpdate
	Bye
	ByeAck
)

type HelloPacketData struct{}

type WelcomePacketData struct {
	ClientId uint8
}

type PlayerUpdatePacketData struct {
	ClientId uint8
	Pos      vector.Vec2
	InDash   bool
}

type PlayerData struct {
	ClientId uint8
	Pos      vector.Vec2
	InDash   bool
}

type BallData struct {
	Owner uint8
	Pos   vector.Vec2
}

type ServerUpdatePacketData struct {
	PlayersData map[uint8]PlayerData
	Balls       []BallData
	Timestamp   time.Time
}

type ByePacketData struct {
	ClientId uint8
}

type ByeAckPacketData struct {
	ClientId uint8
}

type PacketData interface {
	HelloPacketData | WelcomePacketData | PlayerUpdatePacketData | ServerUpdatePacketData | ByePacketData | ByeAckPacketData
}

type Packet[T PacketData] struct {
	Kind PacketKind
	Data T
}

package main

import (
	"github.com/JanCieslak/zbijak/common/packets"
	"log"
	"net"
	"sync"
)

type Player struct {
	x, y float64
}

type Server struct {
	players         map[uint8]*Player
	playersMutex    sync.Mutex
	currentPlayerId uint8
}

func main() {
	log.SetPrefix("Server - ")

	s := Server{
		players:         make(map[uint8]*Player),
		playersMutex:    sync.Mutex{},
		currentPlayerId: 0,
	}

	addr, err := net.ResolveUDPAddr("udp", ":8083")
	if err != nil {
		log.Fatalln("Resolve Addr error:", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listening on:", addr)

	welcomeBuffer := make([]byte, 1)

	for {
		_, remoteAddr, err := conn.ReadFromUDP(welcomeBuffer)
		if err != nil {
			log.Fatalln("Listen error:", remoteAddr, err)
		}

		switch packets.PacketKind(welcomeBuffer[0]) {
		case packets.Hello:
			s.playersMutex.Lock()
			packets.SendWelcomePacket(conn, remoteAddr, conn.LocalAddr().String(), s.currentPlayerId)
			clientId := s.currentPlayerId
			s.players[clientId] = &Player{}
			s.currentPlayerId++
			s.playersMutex.Unlock()
			break
		case packets.PlayerUpdate:
			playerUpdatePacket := packets.ReceivePlayerUpdatePacket(false, conn)
			s.playersMutex.Lock()
			player := s.players[0]
			player.x = playerUpdatePacket.X
			player.y = playerUpdatePacket.Y
			s.playersMutex.Unlock()
			break
		default:
			log.Fatalln("Something went wrong")
		}
	}
}

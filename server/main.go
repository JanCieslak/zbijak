package main

import (
	"github.com/JanCieslak/zbijak/common"
	"github.com/JanCieslak/zbijak/common/packets"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	players         map[uint8]*Player
	playersMutex    sync.Mutex
	currentPlayerId uint8
}

type Player struct {
	ClientId uint8
	X, Y     float64
	addr     *net.UDPAddr
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

	go func() {
		for {
			time.Sleep(time.Second)
			s.playersMutex.Lock()
			players := make([]common.RemotePlayer, len(s.players))
			for clientId, player := range s.players {
				players[clientId] = common.RemotePlayer{
					ClientId: clientId,
					X:        player.X,
					Y:        player.Y,
				}
			}

			log.Printf("players payload %v players map %v", players, s.players)

			for _, player := range s.players {
				packets.SendServerUpdatePacket(conn, player.addr, players)
			}
			s.playersMutex.Unlock()
		}
	}()

	buffer := make([]byte, 32)

	for {
		_, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatalln("Listen error:", remoteAddr, err)
		}

		switch packets.PacketKind(buffer[0]) {
		case packets.Hello:
			s.playersMutex.Lock()
			packets.SendWelcomePacket(conn, remoteAddr, conn.LocalAddr().String(), s.currentPlayerId)
			clientId := s.currentPlayerId
			s.players[clientId] = &Player{
				ClientId: clientId,
				addr:     remoteAddr,
			}
			s.currentPlayerId++
			s.playersMutex.Unlock()
			break
		case packets.PlayerUpdate:
			remotePlayer := packets.ParsePlayerUpdatePacket(buffer)
			s.playersMutex.Lock()
			player := s.players[remotePlayer.ClientId]
			player.X = remotePlayer.X
			player.Y = remotePlayer.Y
			player.ClientId = remotePlayer.ClientId
			s.playersMutex.Unlock()
			break
		default:
			log.Fatalln("Something went wrong")
		}
	}
}

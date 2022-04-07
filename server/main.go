package main

import (
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/JanCieslak/zbijak/common/vec"
	"log"
	"sync"
)

func main() {
	log.SetPrefix("Server - ")
	//log.SetOutput(ioutil.Discard)

	log.Println("Listening on: 8083")

	balls := sync.Map{}

	// TODO Testing balls
	balls.Store(0, &RemoteBall{
		id:      0,
		pos:     vec.NewVec2(300, 300),
		vel:     vec.NewVec2(0, 0),
		ownerId: 0,
	})
	balls.Store(1, &RemoteBall{
		id:      1,
		pos:     vec.NewVec2(600, 300),
		vel:     vec.NewVec2(0, 0),
		ownerId: constants.NoTeam,
	})

	server := &Server{
		players:      sync.Map{},
		nextClientId: 0,
		nextTeam:     constants.TeamOrange,
		balls:        balls,
	}

	netman.InitializeServerSockets(":8083", ":8084", server)
	netman.RegisterTCP(netman.Hello, handleHelloPacket)
	netman.RegisterTCP(netman.Bye, handleByePacket)
	netman.RegisterUDP(netman.PlayerUpdate, handlePlayerUpdatePacket)
	netman.RegisterUDP(netman.Fire, handleFirePacket)
	go netman.ListenUDP()
	go netman.AcceptNewTCPConnections()

	server.Update()

	// TODO will never happen (for now at least - server.Update is infinite for loop)
	netman.ShutDown()
}

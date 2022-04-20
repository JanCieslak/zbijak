package main

import (
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/JanCieslak/zbijak/common/vec"
	"log"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

func main() {
	log.SetPrefix("Server - ")
	//log.SetOutput(ioutil.Discard)

	log.Println("Listening on: 8083")

	balls := sync.Map{}

	// TODO Testing balls
	balls.Store(0, &RemoteBall{
		id:      0,
		team:    constants.NoTeam,
		pos:     vec.NewVec2(300, 300),
		vel:     vec.NewVec2(0, 0),
		ownerId: constants.NoTeam,
	})
	balls.Store(1, &RemoteBall{
		id:      1,
		team:    constants.NoTeam,
		pos:     vec.NewVec2(800, 300),
		vel:     vec.NewVec2(0, 0),
		ownerId: constants.NoTeam,
	})

	server := &Server{
		players:      sync.Map{},
		nextClientId: 0,
		nextTeam:     constants.TeamB,
		balls:        balls,
		shouldRun:    netman.NewAtomicBool(true),
	}

	//go profile(server)

	netman.InitializeServerSockets(":8083", ":8084", server)
	netman.RegisterTCP(netman.Hello, handleHelloPacket)
	netman.RegisterTCP(netman.Bye, handleByePacket)
	netman.RegisterUDP(netman.PlayerUpdate, handlePlayerUpdatePacket)
	netman.RegisterUDP(netman.Fire, handleFirePacket)
	go netman.ListenUDP()
	go netman.AcceptNewTCPConnections()

	//go cmd(server)

	server.Update()

	// TODO will never happen (for now at least - server.Update is infinite for loop)
	netman.ShutDown()
}

func profile(s *Server) {
	f, err := os.Create("profile.pb.gz")
	if err != nil {
		log.Fatal(err)
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(15 * time.Second)
	s.shouldRun.Set(false)
	pprof.StopCPUProfile()
}

//func cmd(s *Server) {
//	reader := bufio.NewReader(os.Stdin)
//
//	for {
//		fmt.Print("-> ")
//		text, _ := reader.ReadString('\n')
//		// convert CRLF to LF
//		text = strings.Replace(text, "\n", "", -1)
//
//		if strings.Compare("exit", text) == 0 {
//			s.shouldRun.Set(false)
//		}
//	}
//}

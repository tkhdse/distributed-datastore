package main

import (
	RPC "MP3/RPC"
	"fmt"
	"os"
	"sync"
)

/*
The main logic of server can be found in /RPC/serverUtils.go
It is placed there so the RPCs can have knowledge of the Server struct
This code serves to just start the server and call the
*/

func main() {
	arguments := os.Args

	if len(arguments) < 3 {
		fmt.Printf("Invalid Usage, expected: ./server [NAME] [CONFIG_FILE]\n")
		return
	}
	server := &RPC.Server{}
	server.Cond = sync.NewCond(&server.Mu)

	server.Name = arguments[1]
	server.Peers = make([](*RPC.Peer), 0)
	server.Accounts = make(map[string](*RPC.Account))
	server.CoordinatorLog = make(map[int64](map[string]int64))

	server.Interacted = make(map[int64](map[RPC.InteractEntry]bool))

	server.CommitCh = make(chan bool, 1)
	server.Waiting = false
	// server.Mu = sync.Mutex{}

	configFile := arguments[2]

	// Read the config file to find out what addresses the peers are at to connect
	server.ParseConfigFile(configFile)

	// Waitgroup to wait for server boot up and run indefinitely
	var wg sync.WaitGroup

	//  defer so we never end the server
	defer wg.Wait()

	// Start the server up to handle RPC requests
	go RPC.StartServer(&wg, server)

	// Connect to all other servers
	RPC.ConnectToAllRPCServers(server.Peers)

}

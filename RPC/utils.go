package RPC

/*
	General Utilities that are required for all or
	multiple RPCs. Specific RPC info in its own files

*/

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

// The definition of ServerService here causes lint errors
// in all RPC files, but the program still compiles completely
// fine

// Server Service defines the service on which the server will receive RPCs
// type ServerService struct {
// 	SERVER *Server
// }

// Retry time for retrying a connection to another server again in milleseconds
var RETRYTIME int64 = 250

var DEBUGMODE bool = false

// Enum for Operation types
type OperationID int

const (
	BEGIN OperationID = iota
	DEPOSIT
	BALANCE
	WITHDRAW
	COMMIT
	ABORT
)

// Enum for StatusCode types
type StatusCode int

const (
	OK        StatusCode = iota
	INVALID              // INVALID (fail)
	READ_OK              // BALANCE OK
	WRITE_OK             // DEPOSIT/WITHDRAW OK
	NA_ABORT             // NOT FOUND -> ABORT
	COMMIT_OK            // COMMIT OK
	ABORTED              // ABORT
)

// Peer struct, can be typedef'd to Coordinator in client for clarity
type Peer struct {
	Name    string
	Address string
	Port    string
	Conn    *rpc.Client
}

// Starts Server at a given port
func StartServer(wg *sync.WaitGroup, server *Server) {
	wg.Add(1)

	// Will never get executed but defer it anyways
	defer wg.Done()

	// Register RPC server
	rpc.Register(server)
	rpc.HandleHTTP()

	// Listen for requests on Server Port
	l, e := net.Listen("tcp", ":"+server.Port)
	if e != nil {
		fmt.Print("listen error:", e)
	}
	http.Serve(l, nil)
}

// Attempts a single connection to an RPC server
func ConnectToRPCServer(address string, port string) (*rpc.Client, error) {
	dial := address + ":" + port
	c, err := rpc.DialHTTP("tcp", dial)
	return c, err
}

// Connects to all other RPC servers. Used in server to establish connections to peers.
// Used in client to establish connections to all servers as potential coordinators.
// Assumes that no server will fail (during startup phase [or ever?]). If a server is
// faulty, this function will not terminate
func ConnectToAllRPCServers(peers []*Peer) {

	connections := 0
	for connections < len(peers) {
		for i, peer := range peers {
			// Peer is already connected, skip
			if (*peer).Conn != nil {
				continue
			}

			// Attempt the connection
			c, err := ConnectToRPCServer(peer.Address, peer.Port)
			if err != nil {
				// Skip and try other connections
				if DEBUGMODE {
					fmt.Printf("Failed to connect to Server %s\n", peer.Name)
				}
				continue
			}

			// Connection was successful
			if !DEBUGMODE {
				fmt.Printf("Connected to Server %s\n", peer.Name)
			}
			peer.Conn = c
			connections++

			// Reassign peer so struct is actually updated
			peers[i] = peer
		}
		// Got all connections needed, exit
		if connections == len(peers) {
			return
		}
		time.Sleep(time.Duration(RETRYTIME) * time.Millisecond)
	}
}

func ArrayMax(list []int64) int64 {
	if len(list) == 0 {
		return 0
	}

	var mx int64 = 0
	for i := 0; i < len(list); i++ {
		mx = max(list[i], mx)
	}

	return mx
}

func max(x int64, y int64) int64 {
	if x > y {
		return x
	}

	return y
}

func Pop(list []TW_Entry, idx int) []TW_Entry {
	return append(list[:idx], list[idx+1:]...)
}

func Contains(list []int64, value int64) bool {
	for _, item := range(list){
		if(item == value){
			return true
		}
	}
	return false
}
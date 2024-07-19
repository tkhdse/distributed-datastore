package RPC

import (
	// "fmt"
	"sync"
	// "strings"
)

type SendOperationArgs struct {
	Operation    OperationID // Type of Operation
	TargetServer string      // Name of the server that the
	// account is on

	AccountName string // Name of Account
	Value       int64  // Amount to change by
	Timestamp   int64  // Timestamp of the transaction
	// that the operation is in
}

type SendOperationReply struct {
	Status StatusCode
	// Only used for Balance
	Value int64
}

/*
Sends the operation from the client to the Coordinator
e.g. DEPOSIT A.foo 100 with Coordinator B would look like

[Client --->--- B] --->--- A

Where SendOperation handles the Server-to-Server
communication [in brackets]. For the Server to Server communication
RPC, see respective Write, Read, and Commit RPCs
*/
// TODO: IMPLEMENT THIS CORRECTLY, CURRENTLY JUST A
// CONTAINS CODE TO TEST CLIENT FUNCTIONALITY WITH
// STATUS CODES
func (server *Server) SendOperation(args *SendOperationArgs, reply *SendOperationReply) error {
	if args.Operation == BEGIN {
		reply.Status = OK
		server.Interacted[args.Timestamp] = make(map[InteractEntry]bool)

		return nil
	}
	
	if args.Operation == DEPOSIT || args.Operation == WITHDRAW {

		write_args := &WriteArgs{
			Timestamp:   args.Timestamp,
			Value:       args.Value,
			Operation:   args.Operation,
			AccountName: args.AccountName,
		}

		write_reply := &WriteReply{}
		err := server.GetPeerByName(args.TargetServer).Conn.Call("Server.WriteOperation", write_args, &write_reply)
		if err != nil {
			reply.Status = INVALID
			reply.Value = -1
			return nil
		}

		reply.Status = write_reply.Status

		if reply.Status == WRITE_OK {
			entry := InteractEntry{
				ServerName: args.TargetServer,
				AccountName: args.AccountName,
			}
			// Add the interaction to the set
			server.Interacted[args.Timestamp][entry] = true
		}

		return nil
	}

	if args.Operation == BALANCE {

		// Construct args and reply
		read_args := &ReadArgs{
			Timestamp:   args.Timestamp,
			Value:       args.Value,
			Operation:   args.Operation,
			AccountName: args.AccountName,
		}

		read_reply := &ReadReply{}

		// Call the RPC
		err := server.GetPeerByName(args.TargetServer).Conn.Call("Server.ReadOperation", read_args, &read_reply)
		if err != nil {
			reply.Status = INVALID
			reply.Value = -1
			return nil
		}

		reply.Status = read_reply.Status
		if reply.Status == READ_OK {
			entry := InteractEntry{
				ServerName: args.TargetServer,
				AccountName: args.AccountName,
			}
			// Add the interaction to the set
			server.Interacted[args.Timestamp][entry] = true
			reply.Value = read_reply.ReadValue
		}

		return nil
	}

	if args.Operation == COMMIT {
		var wg sync.WaitGroup 
		var bool_mu sync.Mutex
		can_commit := true

		for k := range(server.Interacted[args.Timestamp]){
			wg.Add(1)
			go func(entry InteractEntry){
				// fmt.Printf("Sent prepares to %s.%s", entry.ServerName, entry.AccountName)
				prep_args := PrepareArgs{
					AccountName: entry.AccountName,
					Timestamp: args.Timestamp,
				}
				
				prep_reply := &PrepareReply{}

				server.GetPeerByName(entry.ServerName).Conn.Call("Server.PrepareCommit", prep_args, &prep_reply)

				bool_mu.Lock()
				can_commit = can_commit && prep_reply.Response
				bool_mu.Unlock()

				wg.Done()
			}(k)
		}

		wg.Wait()
		if(can_commit){
			// fmt.Println("GOT HERE")
			server.InitiateCommit(args.Timestamp)
			reply.Status = COMMIT_OK
			return nil
		} else{
			server.InitiateAbort(args.Timestamp)
			reply.Status = ABORTED
			return nil
		}
	}

	if args.Operation == ABORT{
		server.InitiateAbort(args.Timestamp)
		reply.Status = ABORTED
		return nil
	}
	reply.Status = NA_ABORT
	return nil
}

package RPC

/*
	RPC to attempt a read on an account
	Used for balance
*/

// import "fmt"

// "fmt"

type AbortArgs struct {
	Timestamp   int64
	AccountName string
}

type AbortReply struct {
}

func (server *Server) AbortTransaction(args *AbortArgs, reply *AbortReply) error {
	server.Abort(args.AccountName, args.Timestamp)
	// fmt.Printf("Tentative Writes: %v", server.Accounts[args.AccountName].TW)
	// fmt.Printf("RTS: %v", server.Accounts[args.AccountName].RTS)
	return nil
}

package RPC

/*
	RPC to commit a transaction
	Used for balance
*/


type CommitArgs struct {
	Timestamp   int64
	AccountName string
}

type CommitReply struct {
	Success 	bool
}

func (server *Server) CommitTransaction(args *CommitArgs, reply *CommitReply) error {
	server.Commit(args.AccountName, args.Timestamp)
	return nil
}

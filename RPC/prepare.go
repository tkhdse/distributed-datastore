package RPC

/*
	RPC to attempt a commit or abort
*/

type PrepareArgs struct {
	Timestamp 	int64
	AccountName string
}

type PrepareReply struct {
	Response bool //true --> we can commit, false --> do not commit
}

func (server *Server) PrepareCommit(args *PrepareArgs, reply *PrepareReply) error {
	// COMMIT/ABORT
	// server.mu.Lock()
	account, found := server.Accounts[args.AccountName]
	// server.mu.Unlock()
	
	if !found {
		reply.Response = true
		return nil
	}

	for _, entry := range(account.TW){
		// Timestamp in TW list
		if(entry.TS == args.Timestamp){
			if entry.Value < 0 {
				reply.Response = false
				return nil
			} else {
				reply.Response = true
				return nil
			}
		}
	}

	// Don't reject a commit if timestamp not in TW list
	reply.Response = true
	return nil
}

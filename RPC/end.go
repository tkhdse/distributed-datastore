package RPC

/*
	RPC to attempt a commit or abort
*/

type EndTransactionArgs struct {
	CommitTimestamp int64
	CommitAccounts  []string
	CommitBalances  []int64
	CommitStatus    bool
}

type EndTransactionReply struct {
	Finished bool
}

func (server *Server) EndTransaction(args *EndTransactionArgs, reply *EndTransactionReply) error {
	// Perform COMMIT/ABORT
	// think about what has to be changed; RTS? TW?
	// update Balance

	if args.CommitStatus {
		for i, acc := range args.CommitAccounts {
			server.Accounts[acc].Balance = args.CommitBalances[i]
			server.Accounts[acc].CTimestamp = args.CommitTimestamp
		}

	} else {
		// DELETE TENTATIVE WRITES (pop entries with TW_entry.TS == CommitTimestsamp) for each account

		for _, acc := range args.CommitAccounts {
			v, err := server.Accounts[acc]
			if !err {
				continue
			}

			for i, tw := range v.TW {
				if tw.TS == args.CommitTimestamp {
					server.Accounts[acc].TW = Pop(server.Accounts[acc].TW, i)
				}
			}

			// if there are no more tentative writes remaining, delete account
			if len(server.Accounts[acc].TW) == 0 {
				delete(server.Accounts, acc)
			}
		}

	}

	// if server.Waiting {
	// 	server.CommitCh <- true
	// }

	server.Mu.Lock()
	server.Cond.Broadcast()
	server.Mu.Unlock()

	// fmt.Println("Broadcast CV")

	// fmt.Println("Signaled and Transaction ended!")

	reply.Finished = true
	return nil
}

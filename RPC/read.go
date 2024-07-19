package RPC

/*
	RPC to attempt a read on an account
	Used for balance
*/

type ReadArgs struct {
	Value       int64
	Timestamp   int64
	Operation   OperationID
	AccountName string
}

type ReadReply struct {
	Status    StatusCode
	ReadValue int64
}

func (server *Server) ReadOperation(args *ReadArgs, reply *ReadReply) error {

	value, status := server.ReadAccountValue(args.AccountName, args.Timestamp)
	reply.ReadValue = value
	reply.Status = status

	return nil
}

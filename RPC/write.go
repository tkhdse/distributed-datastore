package RPC

/*
	RPC to attempt a write on an account
	Used for withdraw and deposit
*/

type WriteArgs struct {
	Value       int64
	Timestamp   int64
	Operation   OperationID
	AccountName string
}

type WriteReply struct {
	Status       StatusCode
	UpdatedValue int64
}

// READ then WRITE
func (server *Server) WriteOperation(args *WriteArgs, reply *WriteReply) error {

	var status StatusCode

	if(args.Operation == DEPOSIT){
		status = server.WriteAccountValue(args.AccountName, args.Value, args.Timestamp)
	} else{
		status = server.WriteAccountValue(args.AccountName, -1 * args.Value, args.Timestamp)
	}
	
	reply.Status = status

	return nil
}

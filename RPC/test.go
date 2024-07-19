package RPC

import "fmt"

type Test struct {
	AccountName string
	Amount      int
}

type TestArgs struct {
	Value string
}

type TestReply struct {
	Response string
}

func (t *Server) GetText(args *TestArgs, reply *TestReply) error {
	fmt.Printf("GOT: %s\n", args.Value)
	reply.Response = "GOOD!"
	return nil
}

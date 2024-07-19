package main

import (
	RPC "MP3/RPC"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type OperationID int

type Client struct {
	clientID string

	// These are the servers
	coordinators []*RPC.Peer

	inTransaction bool
	coordinator   RPC.Peer
	timestamp     int64

	// TODO: Store local transaction changes
}

type Transaction struct {
	Coordinator RPC.Peer
	Operations  []Operation
	Ok          bool
}

type Operation struct {
	ID OperationID
}

func main() {
	arguments := os.Args

	if len(arguments) < 3 {
		fmt.Printf("Invalid Usage, expected: ./client [CLIENTID] [CONFIG_FILE]")
	}
	client := &Client{}
	client.clientID = arguments[1]
	configFile := arguments[2]

	// Parse the configuration file
	client.ParseConfigFile(configFile)

	// Connect to all RPC Servers before proceeding
	// Will stall until all servers are connected
	RPC.ConnectToAllRPCServers(client.coordinators)

	// Start reading stdin for transactions
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)

		if text == "exit" {
			break
		}

		status := client.ProcessOperation(text)

		switch status {
		case RPC.OK:
			fmt.Println("OK")
		case RPC.NA_ABORT:
			fmt.Println("NOT FOUND, ABORTED")
			client.inTransaction = false
			os.Exit(0)
		case RPC.COMMIT_OK:
			fmt.Println("COMMIT OK")
			os.Exit(0)
		case RPC.ABORTED:
			fmt.Println("ABORTED")
			os.Exit(0)

		// Don't print anything if invalid, just ignore
		case RPC.INVALID:
		//  Balance should be printed within Process Operation
		case RPC.READ_OK:
		//
		case RPC.WRITE_OK:
			fmt.Println("OK")

		}

		continue
	}
}

// Processes operation in standard in and returns
// status code indicating success of operation
func (this *Client) ProcessOperation(command string) RPC.StatusCode {
	command_args := strings.Split(command, " ")
	// fmt.Printf("command_args: %v\n", command_args)
	args := &RPC.SendOperationArgs{}
	reply := &RPC.SendOperationReply{}

	// Ignore any input that happends outside of a transaction
	if command_args[0] != "BEGIN" && !this.inTransaction {
		return RPC.INVALID
	}

	// Handle Begin Case
	if len(command_args) == 1 && command_args[0] == "BEGIN" && !this.inTransaction{
		this.inTransaction = true
		this.coordinator = *this.GetRandomCoordinator()
		this.timestamp = this.GetTimestamp()

		args.Operation = RPC.BEGIN
		args.Timestamp = this.timestamp

		// Handle Deposit Case (WRITE)
	} else if len(command_args) == 3 &&
		command_args[0] == "DEPOSIT" {

		args.Operation = RPC.DEPOSIT

		args.TargetServer, args.AccountName = GetTargetAndAccount(command_args[1])

		val, err := strconv.Atoi(command_args[2])
		if err != nil {
			return RPC.INVALID
		}
		args.Value = int64(val)
		args.Timestamp = this.timestamp

		// Handle Withdraw Case (WRITE)
	} else if len(command_args) == 3 &&
		command_args[0] == "WITHDRAW" {

		args.Operation = RPC.WITHDRAW

		args.TargetServer, args.AccountName = GetTargetAndAccount(command_args[1])

		val, err := strconv.Atoi(command_args[2])
		if err != nil {
			return RPC.INVALID
		}
		args.Value = int64(val)

		args.Timestamp = this.timestamp

		// Handle the Balance Case (READ)
	} else if len(command_args) == 2 &&
		command_args[0] == "BALANCE" {

		args.Operation = RPC.BALANCE

		args.TargetServer, args.AccountName = GetTargetAndAccount(command_args[1])

		args.Timestamp = this.timestamp

		// Handle the Abort Case
	} else if len(command_args) == 1 &&
		command_args[0] == "ABORT" {

		// No longer in transaction
		this.inTransaction = false

		args.Operation = RPC.ABORT
		args.Timestamp = this.timestamp

		// Handle the Commit Case
	} else if len(command_args) == 1 &&
		command_args[0] == "COMMIT" {
		// args.
		// No longer in transaction
		this.inTransaction = false

		args.Operation = RPC.COMMIT
		args.Timestamp = this.timestamp

	} else {
		return RPC.INVALID
	}

	// Send the RPC
	err := this.coordinator.Conn.Call("Server.SendOperation", args, &reply)
	if err != nil {
		fmt.Println("CLIENT CALL DID NOT GO THROUGH")
		return RPC.INVALID
	}

	// Print Balance here, everything else will be printed
	// in main
	if args.Operation == RPC.BALANCE && reply.Status == RPC.READ_OK {
		fmt.Printf("%s.%s = %d\n", args.TargetServer, args.AccountName, reply.Value)
	}
	return reply.Status
}

// Assumes well-formed input
func GetTargetAndAccount(input string) (string, string) {
	data := strings.Split(input, ".")
	target := data[0]
	account := data[1]
	return target, account
}

// Gets random coordinator for each transaction
func (this *Client) GetRandomCoordinator() *RPC.Peer {
	return this.coordinators[rand.Intn(len(this.coordinators))]
}

func (this *Client) GetTimestamp() int64 {
	return time.Now().UnixNano()
}

// Parses the config file to populate client
func (this *Client) ParseConfigFile(filename string) {
	file, ferr := os.Open(filename)

	// Attempt to open file
	if ferr != nil {
		fmt.Println("Cannot find file: " + filename)
		panic(ferr)
	}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")

		// Assemble Peer
		coordinator := RPC.Peer{
			Name:    fields[0],
			Address: fields[1],
			Port:    fields[2],
			Conn:    nil,
		}

		// Add peer to the server's list
		this.coordinators = append(this.coordinators, &coordinator)
	}
}

package RPC

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

/*
Servers act as servers to the clients and clients/servers to fellow servers (Peers).
As such we set up an open connection that clients and peers will connect to
and make sure that we initiate connections to all other servers.
*/

var NULL int64 = -1

type InteractEntry struct {
	ServerName 	string
	AccountName string
}

type Server struct {
	// Server intrinsics
	Mu       sync.Mutex
	Cond     *sync.Cond
	Waiting  bool // has transaction with Server ended?
	CommitCh chan bool

	Name    string
	Address string
	Port    string

	// Peers for the server to connect to
	Peers          [](*Peer)
	Accounts       map[string](*Account)
	CoordinatorLog map[int64](map[string]int64) // timestamp : (account: tent_amt)
	Interacted 	map[int64](map[InteractEntry]bool) // Maps timestamp -> (map of ServerName, AccountName -> Presence)
}

type TW_Entry struct {
	TS    int64
	Value int64
}

type Account struct {
	// mu      sync.Mutex
	Balance    int64 // Committed value
	CTimestamp int64 // Timestamp that wrote the committed value
	Name       string
	RTS        []int64
	TW         []TW_Entry

	aborted  bool
	tent_bal int64

	Mu       	*sync.Mutex
	Cond     *sync.Cond
	// CommitCond	 *sync.Cond
}

/*
Function that attempts to read the account value in conjunction with timestamped

param : accountName -> name of the account
param : timestamp -> timestamp of the  
*/
func (server *Server) ReadAccountValue(accountName string, timestamp int64) (int64, StatusCode) {
	// TODO: Fix Locking	
	// server.Mu.Lock()
	account, found := server.Accounts[accountName]
	// server.Mu.Unlock()

	// The account doesn't exist on the server yet, abort
	// Note: We check the deposit case in Write
	if(!found){
		return -1, NA_ABORT
	}
	
	// account.Mu.Lock()
	if(account.CTimestamp == 0 && account.Balance == 0 && len(account.TW) == 0) {
		// account.Mu.Unlock()
		return -1, NA_ABORT
	}

	if(timestamp > account.CTimestamp){
		// Ds is the largest timestamp across TW and Committed Timestamp
		Ds_value, Ds_timestamp  := account.GetLargestWriteTimestamp(timestamp)

		if(Ds_timestamp == account.CTimestamp){
			// Read Ds and add Tc to RTS list (if not already added)
			if(!Contains(account.RTS, timestamp)){
				account.RTS = append(account.RTS, timestamp)
			}
			// account.Mu.Unlock()
			return Ds_value, READ_OK
		}
		// Else
		if(Ds_timestamp == timestamp){
			// account.Mu.Unlock()
			return Ds_value, READ_OK
		}
		
		// TODO: Implement blocking for the read
		// Block until a commit occurs
		account.Mu.Lock()
		account.Cond.Wait()
		account.Mu.Unlock()
		return server.ReadAccountValue(accountName, timestamp)
	}

	// Else timestamp < account.CTimestamp
	// Too late, we must abort
	// account.Mu.Unlock()
	return -1, ABORTED
}

// Attempt a write on a
func (server *Server) WriteAccountValue(accountName string, delta int64, timestamp int64) StatusCode {
	// non-negative delta denotes a deposit
	// negative delta denotes a withdraw
	value, status := server.ReadAccountValue(accountName, timestamp)

	// Read failed, we must abort the write
	if(status == ABORTED){
		return ABORTED
	}

	// Account was not found and we are doing a withdraw transaction
	if(status == NA_ABORT && delta < 0){
		return NA_ABORT
	}

	// Deposit transaction, create account
	if(status == NA_ABORT){
		value += 1
		// server.Mu.Lock()
		server.InitAccount(accountName);
		// server.Mu.Unlock()
	}

	// server.Mu.Lock()
	account, found := server.Accounts[accountName]

	if(!found){
		return ABORTED
	}
	// account.Mu.Lock()
	// server.Mu.Unlock()
	write_value := value + delta


	max_read_TS := ArrayMax(account.RTS)
	if(timestamp >= max_read_TS && timestamp > account.CTimestamp){
		for i, entry := range(account.TW){
			if(entry.TS == timestamp){
				entry.Value = write_value
				account.TW[i] = entry

				// account.Mu.Unlock()
				return WRITE_OK				
			}
		}
		// No existing tentative write for the timestamp, create
		// a new TW entry
		tentative_write := TW_Entry{TS: timestamp, Value: write_value}
		account.TW = append(account.TW, tentative_write)
		// account.Mu.Unlock()
		return WRITE_OK
	}


	// account.Mu.Unlock()
	return ABORTED
}

func (server *Server) Commit(accountName string, timestamp int64) StatusCode {
	// server.Mu.Lock()
	account, found := server.Accounts[accountName]
	// account.Mu.Lock()
	// server.Mu.Unlock()
	// If account is not on this server, there is no issue committing
	// fmt.Println("GOT TO LINE 190")
	if !found {
		// account.Mu.Unlock()
		return OK
	}

	// Max value of int64
	var min_timestamp int64 = (1 << 63) - 1
	var write_value int64 = -1
	index := 0
	for i, entry := range(account.TW){
		if(entry.TS < min_timestamp){
			min_timestamp = entry.TS
			write_value = entry.Value
			index = i
		}
	}
	// If it is the minimum, attempt to commit
	if(timestamp == min_timestamp){
		account.CTimestamp = timestamp
		account.Balance = write_value

		account.TW = Pop(account.TW, index)
		// fmt.Printf("Balance: %d, Committed by: %d", account.Balance, account.CTimestamp)
		account.Cond.Broadcast()
		// account.Mu.Unlock()
		return OK
	} else{
		// Otherwise we block until a commit goes through and then we check again
		account.Mu.Lock()
		account.Cond.Wait()
		account.Mu.Unlock()
		return server.Commit(accountName, timestamp)
	}
}

func (server *Server) Abort(accountName string, timestamp int64) StatusCode {
	// server.Mu.Lock()
	account, found := server.Accounts[accountName]
	// account.Mu.Lock()
	// server.Mu.Unlock()

	if(found){
		// Remove read timestamp from the list
		read_index := -1 
		for i, entry := range(account.RTS){
			if(entry == timestamp){
				read_index = i
			}
		}
		if read_index != -1 {
			account.RTS = append(account.RTS[:read_index], account.RTS[read_index+1:]...)
		}

		// Remove tentative write timestamp from the list
		write_index := -1
		for i, entry := range(account.TW){
			if(entry.TS == timestamp){
				write_index = i
			}
		}
		if write_index != -1 {
			account.TW = Pop(account.TW, write_index)
		}

		account.Cond.Broadcast()
	}
	// account.Mu.Unlock()
	return ABORTED
}

func (server *Server) InitiateAbort(timestamp int64) {
	server.Mu.Lock()
	// Get the set of accounts
	interactedAccounts := server.Interacted[timestamp]
	
	// Iterate through set and send an Abort RPC
	for k := range(interactedAccounts){
		targetServer := k.ServerName
		accountName := k.AccountName
		
		args := &AbortArgs{
			AccountName: accountName,
			Timestamp: timestamp,
		}
		reply := &AbortReply{}
	
		err := server.GetPeerByName(targetServer).Conn.Call("Server.AbortTransaction", args, &reply)
		if err != nil{
			fmt.Print("FAILED TO ABORT")
		}
	}

	server.Mu.Unlock()
}

func (server *Server) InitiateCommit(timestamp int64) {
	server.Mu.Lock()
	defer server.Mu.Unlock()
	
	interactedAccounts := server.Interacted[timestamp]
	
	// Iterate through set and send an Abort RPC
	for k := range(interactedAccounts){
		targetServer := k.ServerName
		accountName := k.AccountName
		
		args := &CommitArgs{
			AccountName: accountName,
			Timestamp: timestamp,
		}
		reply := &CommitReply{}
	
		err := server.GetPeerByName(targetServer).Conn.Call("Server.CommitTransaction", args, &reply)
		if err != nil{
			fmt.Print("FAILED TO COMMIT")
		}
	}
}


// Returns The largest TW or committed value timestamp and 
// associated value
func (a *Account) GetLargestWriteTimestamp(transactionTimestamp int64) (int64, int64){

	timestamp := a.CTimestamp
	value := a.Balance

	for _, entry := range(a.TW){
		if(entry.TS > timestamp && entry.TS <= transactionTimestamp){
			timestamp = entry.TS
			value = entry.Value
		}
	}

	return value, timestamp
}


// Parses the configuration file [WORKS]
func (s *Server) ParseConfigFile(filename string) {
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

		if fields[0] == s.Name {
			s.Address = fields[1]
			s.Port = fields[2]
			// continue
		}

		// Assemble Peer
		peer := Peer{
			Name:    fields[0],
			Address: fields[1],
			Port:    fields[2],
			// Initially nil, initialized by
			// ConnectToAllRPCServers
			Conn: nil,
		}

		// Add peer to the server's list
		s.Peers = append(s.Peers, &peer)
	}
}

// Nice Print for Server [WORKS]
func (s *Server) String() string {
	peersStr := ""
	for _, peer := range s.Peers {
		peersStr += fmt.Sprintf("\t{Name: %s, Address: %s, Port: %s},\n ", peer.Name, peer.Address, peer.Port)
	}
	return fmt.Sprintf("Server{\n  Name: %s,\n  Address: %s,\n  Port: %s,\n  Peers: [\n%s]}", s.Name, s.Address, s.Port, peersStr)
}

// Utility to get peer by its name [WORKS]
func (s *Server) GetPeerByName(name string) Peer {

	for _, peer := range s.Peers {
		if peer.Name == name {
			return *peer
		}
	}
	fmt.Printf("ERROR: Couldn't find server with name %v\n", name)
	os.Exit(1)
	// Return empty Peer to stop compiler error
	return Peer{
		Name:    "",
		Address: "",
		Port:    "",
		Conn:    nil,
	}
}

func (s *Server) InitAccount(name string) {
	// Check to see if account already;
	_, exists := s.Accounts[name]
	if exists {
		return
	}

	mu := &sync.Mutex{}
	(*s).Accounts[name] = &Account{
		Name:       name,
		Balance:    0,
		CTimestamp: 0,
		RTS:        make([]int64, 0),
		TW:         make([]TW_Entry, 0),
		Mu: 		mu,
		Cond: 		sync.NewCond(mu),
	}
}

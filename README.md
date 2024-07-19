To start server:
go run server.go [BRANCH NAME] [PORT]


To start client:
go run client.go [CLIENT ID] [config.txt]

Tasks:
* Create user on new_user found
* Coordinator keeps track?



DEPOSIT:
* given target and account; from Coordinator, route to appropriate server (via RPC) and set up account if needed; or else start writing while following timestamped order rules.

WITHDRAW:
* same as DEPOSIT


DEPOSIT / WITHDRAW:
* Treat as READ then WRITE


BALANCE:
* Implement wait until COMMIT/ABORT Ti (max WTS <= Tc)
* Handle this when we multicast Commit (EndTransaction)
* COMMIT -> read new value
* ABORT -> read old value


If ABORTED:
* Remove newly added RTS?
* Quit Client that coordinator send ABORT to?


Think about what do we Commit?
* All of the updated values from the transaction
* So, as we send the updatee RPC's, we need to receive the values back too (update Write structs)
* Coordinator server must store these values
* On commit time, check if any are negative --> deny commit and ABORT
* on Commit --> value and timestamp are updated, TW is removed

* Create accounts that were created via DEPOSIT in the given transaction


Concurrent Transactions during 2PC?
* READ's must wait until commit anyway so should be fine??

EndTransactionArgs might have to change dynamically since only a few accounts/balances exist on a given server


What if T1 calls DEPOSIT A.xyz 100 and creates account xyz on server A,
and then T2 wants to process off of this value, but T1 ABORTs?
* Does T1's ABORT cascade to T2?
* idt it's possible for T1 to ABORT
* According to Read rule, T2 has to wait until T1 commits/aborts anyway


Optimizations:
* use a Set instead of an array for RTS and TW


Current Bugs/Fixes to Address:
* -/ Implement 2PC
* -/ COMMIT needs to actually update values
* -/ Return value from Read (BALANCE) --> reads are dynamic depending on committed transactions
* Check logic of tent_bal
* Implement Read then Write for DEPOSIT/WITHDRAW
* Aborted account logic

How do we deal with ABORTS that happen within a WriteOperation/ReadOperation e.g?



Test ABORTS
TODO: 
1. Explain how to sync up the smart contract with the DKG
2. Explain how to sync up the smart contract with the Shuffle
3. Explain how security impacts the smart contract (at the decrypt stage)

## Create an election

This transaction requires the following 3 parameters:
1. `title` of the election
2. `admin` ID of the creator of the election
3. `format` of the election

Key / Value pairs sent in the transaction in order to create an election:
| | |
|-|-|
|"go.dedis.ch/dela.ContractArg"|[]byte(evoting.ContractName)|
|evoting.CreateElectionArg|createElectionBuf|
|evoting.CmdArg|[]byte(evoting.CmdCreateElection|

where:
``` go
evoting.ContractName = "go.dedis.ch/dela.Evoting"
evoting.CreateElectionArg = "evoting:create_election"
createElectionBuf = marshalled version of types.CreateElectionTransaction{
        Title: title,
        AdminID: admin,
        Format: format
    }
evoting.CmdArg = "evoting:command"
evoting.CmdCreateElection = "CREATE_ELECTION"
```

On success, the result of this transaction returns a `transactionID`. This `transactionID` is then 
processed as follow to computed the unique `electionID`:

``` go
hash := sha256.New()
hash.Write(transactionID)
electionID := hash.Sum(nil)
```

## Open an election

This transaction requires an `electionID`

Key / Value pairs sent in the transaction in order to create an election:
| | |
|-|-|
|"go.dedis.ch/dela.ContractArg"|[]byte(evoting.ContractName)|
|evoting.OpenElectionArg|openElectionBuf|
|evoting.CmdArg|[]byte(evoting.CmdOpenElection|

where:
``` go
    evoting.ContractName = "go.dedis.ch/dela.Evoting"
    evoting.OpenElectionArg = "evoting:open_election"
    openElectionBuf = marshalled version of types.OpenElectionTransaction{
            ElectionID: hex.EncodeToString(electionID),
        }
    evoting.CmdArg = "evoting:command"
    evoting.CmdOpenElection = "OPEN_ELECTION"
```

## Cast a vote

This transaction requires the following 3 parameters:
1. `actor` of type dkg.Actor
2. `electionID` (see Create Election above)
3. `userID` of the voter
4. `vote` to be casted

Key / Value pairs sent in the transaction in order to create an election:
| | |
|-|-|
|"go.dedis.ch/dela.ContractArg"|[]byte(evoting.ContractName)|
|evoting.CastVoteArg|castVoteBuf|
|evoting.CmdArg|[]byte(evoting.CmdCastVote|

where:
``` go   
    evoting.ContractName = "go.dedis.ch/dela.Evoting"
    evoting.CastVoteArg = "evoting:cast_vote"
    castVoteBuf = a marshalled version of types.CastVoteTransaction{
			ElectionID: hex.EncodeToString(electionID),
			UserID:     userID,
			Ballot:     ballot,   // a vote encrypted by the actor
		}
    evoting.CmdArg = "evoting:command"
    evoting.CmdCastVote = "CAST_VOTE"
```

## Close an election

This transaction requires an `electionID` and an `adminID`.

Key / Value pairs sent in the transaction in order to create an election:
| | |
|-|-|
|"go.dedis.ch/dela.ContractArg"|[]byte(evoting.ContractName)|
|evoting.CloseElectionArg|closeElectionBuf|
|evoting.CmdArg|[]byte(evoting.CmdOpenElection|

where:
``` go
    evoting.ContractName = "go.dedis.ch/dela.Evoting"
    evoting.CloseElectionArg = "evoting:close_election"
    closeElectionBuf = marshalled version of types.CloseElectionTransaction{
		ElectionID: hex.EncodeToString(electionID),
		UserID:     adminID,
	}
    evoting.CmdArg = "evoting:command"
    evoting.CmdCloseElection = "CLOSE_ELECTION"
```
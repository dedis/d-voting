TODO:

1. Explain how to sync up the smart contract with the DKG
2. Explain how to sync up the smart contract with the Shuffle
3. Explain how security impacts the smart contract (at the decrypt stage)

## Create a form

This transaction requires the following 3 parameters:

1. `title` of the form
2. `admin` ID of the creator of the form
3. `format` of the form

Key / Value pairs sent in the transaction in order to create a form:
| | |
|-|-|
|"go.dedis.ch/dela.ContractArg"|[]byte(evoting.ContractName)|
|evoting.CreateFormArg|createFormBuf|
|evoting.CmdArg|[]byte(evoting.CmdCreateForm|

where:

```go
evoting.ContractName = "go.dedis.ch/dela.Evoting"
evoting.CreateFormArg = "evoting:create_form"
createFormBuf = marshalled version of types.CreateFormTransaction{
        Title: title,
        AdminID: admin,
        Format: format
    }
evoting.CmdArg = "evoting:command"
evoting.CmdCreateForm = "CREATE_FORM"
```

On success, the result of this transaction returns a `transactionID`. This `transactionID` is then
processed as follow to computed the unique `formID`:

```go
hash := sha256.New()
hash.Write(transactionID)
formID := hash.Sum(nil)
```

## Open a form

This transaction requires an `formID`

Key / Value pairs sent in the transaction in order to create a form:
| | |
|-|-|
|"go.dedis.ch/dela.ContractArg"|[]byte(evoting.ContractName)|
|evoting.OpenFormArg|openFormBuf|
|evoting.CmdArg|[]byte(evoting.CmdOpenForm|

where:

```go
    evoting.ContractName = "go.dedis.ch/dela.Evoting"
    evoting.OpenFormArg = "evoting:open_form"
    openFormBuf = marshalled version of types.OpenFormTransaction{
            FormID: hex.EncodeToString(formID),
        }
    evoting.CmdArg = "evoting:command"
    evoting.CmdOpenForm = "OPEN_FORM"
```

## Cast a vote

This transaction requires the following 3 parameters:

1. `actor` of type dkg.Actor
2. `formID` (see Create Form above)
3. `VoterID` of the voter
4. `vote` to be casted

Key / Value pairs sent in the transaction in order to create a form:
| | |
|-|-|
|"go.dedis.ch/dela.ContractArg"|[]byte(evoting.ContractName)|
|evoting.CastVoteArg|castVoteBuf|
|evoting.CmdArg|[]byte(evoting.CmdCastVote|

where:

```go
    evoting.ContractName = "go.dedis.ch/dela.Evoting"
    evoting.CastVoteArg = "evoting:cast_vote"
    castVoteBuf = a marshalled version of types.CastVoteTransaction{
			FormID: hex.EncodeToString(formID),
			VoterID:     VoterID,
			Ballot:     ballot,   // a vote encrypted by the actor
		}
    evoting.CmdArg = "evoting:command"
    evoting.CmdCastVote = "CAST_VOTE"
```

## Close a form

This transaction requires an `formID` and an `adminID`.

Key / Value pairs sent in the transaction in order to create a form:
| | |
|-|-|
|"go.dedis.ch/dela.ContractArg"|[]byte(evoting.ContractName)|
|evoting.CloseFormArg|closeFormBuf|
|evoting.CmdArg|[]byte(evoting.CmdOpenForm|

where:

```go
    evoting.ContractName = "go.dedis.ch/dela.Evoting"
    evoting.CloseFormArg = "evoting:close_form"
    closeFormBuf = marshalled version of types.CloseFormTransaction{
		FormID: hex.EncodeToString(formID),
		VoterID:     adminID,
	}
    evoting.CmdArg = "evoting:command"
    evoting.CmdCloseForm = "CLOSE_FORM"
```

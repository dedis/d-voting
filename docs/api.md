# API documentation

_Documentation Last Review: 27.06.2024_

## Regular workflow:

The form workflow involves 3 actors:

- Smart contract
- DKG service
- Neff shuffle service
- transaction manager

Services are side components that augment the smart contract functionalities.
Services are accessed via the `evoting/services/<dkg>|<neff>/*` endpoint, and
the smart contract via `/evoting/forms/*`.

## Signed requests

Requests marked with 🔐 are encapsulated into a signed request as described in
[msg_sig.md](msg_sig.md).

```
Smart contract   DKG       Neff shuffle             Transaction manager
--------------   ---       ------------              ------------------
    │             │        NS1:Init (on startup)            ▲
    ▼             │              │                          │
SC1:Create        │              │                          │
    │             │              │                          │
    │             ▼              │                          │
    │          DK1:Init          │                          │
    │             │              │                          │
    │             ▼              │                          │
    │          DK2:Setup         │                          │
    │             │              │                          │
    │             ▼              │                          │
    │          DK3: DKG get info │                          │
    │             │              │                          │
    ▼             │              │                          │
SC3:Open          │              │                          │
    │             │              │                          │
    ▼             │              │          T1:When the election checks if a transaction 
SC4:Cast          │              │                included in the blockchain
    │             │              │                          │
    ▼             │              │                          │
SC5:Close         │              │                          │
    │             │              │                          │
    │             │              ▼                          │
    │             │          NS2:Shuffle                    │
    │             │                                         │
    │             ▼                                         │
    │         DK4:ComputePubshares                          │
    │                                                       │
    ▼                                                       │
SC6:CombineShares                                           │
    │                                                       │
    ▼                                                       │
SC2:FormGetInfo                                             ▼



```

In case of error:

`500 ERROR` `application/json`

```json
{
  "Title": "",
  "Code": "<uint>",
  "Message": "",
  "Args": {}
}
```

For the election related responses, the `Status` field is indicating whether the transaction for the request was included in the blockchain or not. If the transaction was not included, the `Status` field is set to `0`. Otherwise, it is set to `1`.
The `Token` field is a URL encoded string that allows the proxy of the blockchain node to identify the transaction. It represents the URL encoding of the following structure:

```json
{
  "Status" : "<uint>",   
	"TransactionID" : "<hex encoded>",
	"LastBlockIdx"  : "<uint>",
	"Time"          : "<uint>",
	"Hash"          : "<hex encoded>",
	"Signature"     : "<hex encoded>"
}
```
Where `LastBlockIdx` is the index of the last block of the blockchain before the transaction was submitted, `Hash` is the hash of all the above fields and `Signature` is the signature of the hash by the blockchain node's proxy.

# SC1: Form create 🔐

|        |                    |
| ------ | ------------------ |
| URL    | `/evoting/forms`   |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
  "Configuration": {<Configuration>}
}
```

Return:

`200 OK` 

```json
{
  "FormID": "<hex encoded>",
  "Token" : "<URL encoded>"
}
```

# SC2: Form get info

|        |                           |
| ------ | ------------------------- |
| URL    | `/evoting/forms/{FormID}` |
| Method | `GET`                     |

Return:

`200 OK` 

```json
{
  "FormID": "<hex encoded>",
  "Status": "",
  "Pubkey": "<hex encoded>",
  "Result": [
    {
      "SelectResultIDs": ["<string>"],
      "SelectResult": [["<bool>"]],
      "RankResultIDs": ["<string>"],
      "RankResult": [["<int8>"]],
      "TextResultIDs": ["<string>"],
      "TextResult": [["<string>"]]
    }
  ],
  "Roster": ["<string>"],
  "ChunksPerBallot": "<int>",
  "BallotSize": "<int>",
  "Configuration": {<Configuration>},
  "Voters": ["<string>"]
}
```

# SC3: Form open 🔐

|        |                           |
| ------ | ------------------------- |
| URL    | `/evoting/forms/{FormID}` |
| Method | `PUT`                     |
| Input  | `application/json`        |

```json
{
  "Action": "open"
}
```

Return:

`200 OK` 

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# SC4: Form cast vote 🔐

|        |                                |
| ------ | ------------------------------ |
| URL    | `/evoting/forms/{FormID}/vote` |
| Method | `POST`                         |
| Input  | `application/json`             |

```json
{
  "VoterID": "",
  "Ballot": [
    {
      "K": "<bin>",
      "C": "<bin>"
    }
  ]
}
```

Return:

`200 OK` 

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# SC5: Form close 🔐

|        |                           |
| ------ | ------------------------- |
| URL    | `/evoting/forms/{FormID}` |
| Method | `PUT`                     |
| Input  | `application/json`        |

```json
{
  "Action": "close"
}
```

Return:

`200 OK` 
```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# NS2: Form shuffle 🔐

|        |                                      |
| ------ | ------------------------------------ |
| URL    | `/evoting/services/shuffle/{FormID}` |
| Method | `PUT`                                |
| Input  | `application/json`                   |

```json
{
  "Action": "shuffle"
}
```

Return:

`200 OK` 
```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# SC6: Form combine shares 🔐

|        |                           |
| ------ | ------------------------- |
| URL    | `/evoting/forms/{FormID}` |
| Method | `PUT`                     |
| Input  | `application/json`        |

```json
{
  "Action": "combineShares"
}
```

Return:

`200 OK` `{Status, Token}`

```

```

# SC7: Form cancel 🔐

|        |                           |
| ------ | ------------------------- |
| URL    | `/evoting/forms/{FormID}` |
| Method | `PUT`                     |
| Input  | `application/json`        |

```json
{
  "Action": "cancel"
}
```

Return:

`200 OK` 

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# SC8: Form delete

|         |                            |
| ------- | -------------------------- |
| URL     | `/evoting/forms/{FormID}`  |
| Method  | `DELETE`                   |
| Input   |                            |
| Headers | {Authorization: `<token>`} |

The `<token>` value must be the hex-encoded signature of the hex-encoded
formID:

```
<token> = hex( sig( hex( formID ) ) )
```

Return:

`200 OK` 

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# SC9: Form infos from all forms

|        |                  |
| ------ | ---------------- |
| URL    | `/evoting/forms` |
| Method | `GET`            |
| Input  |                  |

Return:

`200 OK` `application/json`

```json
{
  "Forms": [
    {
      "FormID": "<hex encoded>",
      "Title": "",
      "Status": "",
      "Pubkey": "<hex encoded>"
    }
  ]
}
```

# SC10: Add an owner to a form 🔐

|        |                                   |
| ------ |-----------------------------------|
| URL    | `/evoting/form/{formID}/addowner` |
| Method | `POST`                            |
| Input  | `application/json`                |
```json
{
  "TargetUserID": "<SCIPER>",
  "PerformingUserID": "<SCIPER>"
}
```

Return:

`200 OK`

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# SC11: Remove an owner from the form 🔐

|        |                                      |
|--------|--------------------------------------|
| URL    | `/evoting/form/{formID}/removeowner` |
| Method | `POST`                               |
| Input  | `application/json`                   |
```json
{
  "TargetUserID": "<SCIPER>",
  "PerformingUserID": "<SCIPER>"
}
```

Return:

`200 OK`

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# SC12: Add a voter to the Form 🔐

|        |                                   |
| ------ |-----------------------------------|
| URL    | `/evoting/form/{formID}/addvoter` |
| Method | `POST`                            |
| Input  | `application/json`                |
```json
{
  "TargetUserID": "<SCIPER>",
  "PerformingUserID": "<SCIPER>"
}
```

Return:

`200 OK`

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# SC13: Remove a voter from the Form 🔐

|        |                                      |
|--------|--------------------------------------|
| URL    | `/evoting/form/{formID}/removevoter` |
| Method | `POST`                               |
| Input  | `application/json`                   |
```json
{
  "TargetUserID": "<SCIPER>",
  "PerformingUserID": "<SCIPER>"
}
```

Return:

`200 OK`

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# DK1: DKG init 🔐

|        |                                |
| ------ | ------------------------------ |
| URL    | `/evoting/services/dkg/actors` |
| Method | `POST`                         |
| Input  | `application/json`             |

```json
{
  "FormID": "<hex encoded>",
  "Proxy:": ""
}
```

Return:

`200 OK` `{Status, Token}`

```

```

# DK2: DKG setup 🔐

|        |                                         |
| ------ | --------------------------------------- |
| URL    | `/evoting/services/dkg/actors/{FormID}` |
| Method | `PUT`                                   |
| Input  | `application/json`                      |

```json
{
  "Action": "setup",
  "Proxy:": ""
}
```

Return:

`200 OK` `text/plain`

```

```

# DK3: DKG get info

|        |                                         |
| ------ | --------------------------------------- |
| URL    | `/evoting/services/dkg/actors/{FormID}` |
| Method | `GET`                                   |
| Input  |                                         |

Return:

`200 OK` `application/json`

```json
{
  "Status": "<int>",
  "Error": {
    "Title": "",
    "Code": "<uint>",
    "Message": "",
    "Args": {}
  }
}
```

# DK4: DKG begin decryption 🔐

|        |                                         |
| ------ | --------------------------------------- |
| URL    | `/evoting/services/dkg/actors/{FormID}` |
| Method | `PUT`                                   |
| Input  | `application/json`                      |

```json
{
  "Action": "computePubshares"
}
```

Return:

`200 OK` `text/plain`

```

```

# T1: Check election transaction included



|        |                                |
| ------ | -------------------------------|
| URL    | `/evoting/transactions/{Token}` |
| Method | `GET`                          |
| Input  | `application/json`             |

Return:

`200 OK` 

```json
{
  "Status": "<int>",
  "Token": "<URL encoded>"
}
```
Status can be:
- 0: transaction not yet included
- 1: transaction included
- 2: transaction not included

The token is an updated version of the token in the URL that can be used to check again the status of the transaction if it is not yet included.

# A1: Add an admin to the AdminList 🔐

|        |                     |
| ------ |---------------------|
| URL    | `/evoting/addadmin` |
| Method | `POST`              |
| Input  | `application/json`  |
```json
{
  "TargetUserID": "<SCIPER>",
  "PerformingUserID": "<SCIPER>"
}
```

Return:

`200 OK`

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# A2: Remove an admin from the AdminList 🔐

|        |                        |
| ------ |------------------------|
| URL    | `/evoting/removeadmin` |
| Method | `POST`                 |
| Input  | `application/json`     |

```json
{
  "TargetUserID": "<SCIPER>",
  "PerformingUserID": "<SCIPER>"
}
```

Return:

`200 OK`

```json
{
  "Status": 0,
  "Token": "<URL encoded>"
}
```

# A3: Get the AdminList 



|        |                      |
| ------ |----------------------|
| URL    | `/evoting/adminlist` |
| Method | `GET`                |
| Input  |                      |

Return:

`200 OK`

```json
{
   "<SCIPER>", "<SCIPER>", "..."
}
```



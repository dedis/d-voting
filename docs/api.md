# API documentation

_Documentation Last Review: 11.04.2022_

## Regular workflow:

The election workflow involves 3 actors:

- Smart contract
- DKG service
- Neff shuffle service

Services are side components that augment the smart contract functionalities.
Services are accessed via the `evoting/services/<dkg>|<neff>/*` endpoint, and
the smart contract via `/evoting/elections/*`.

## Signed requests

Requests marked with 🔐 are encapsulated into a signed request as described in
[msg_sig.md](msg_sig.md).

```
Smart contract   DKG       Neff shuffle
--------------   ---       ------------
    │             │        NS1:Init (on startup)
    ▼             │              │
SC1:Create        │              │
    │             │              │
    │             ▼              │
    │          DK1:Init          │
    │             │              │
    │             ▼              │
    │          DK2:Setup         │
    │             │              │
    │             ▼              │
    │          DK3: DKG get info │
    │             │              │
    ▼             │              │
SC3:Open          │              │
    │             │              │
    ▼             │              │
SC4:Cast          │              │
    │             │              │
    ▼             │              │
SC5:Close         │              │
    │             │              │
    │             │              ▼
    │             │          NS2:Shuffle
    │             │
    │             ▼
    │         DK4:ComputePubshares
    │
    ▼
SC6:CombineShares
    │
    ▼
SC2:ElectionGetInfo



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

# SC1: Election create 🔐

|        |                      |
| ------ | -------------------- |
| URL    | `/evoting/elections` |
| Method | `POST`               |
| Input  | `application/json`   |

```json
{
  "Configuration": {<Configuration>}
}
```

Return:

`200 OK` `application/json`

```json
{
  "ElectionID": "<hex encoded>"
}
```

# SC2: Election get info

|        |                                   |
| ------ | --------------------------------- |
| URL    | `/evoting/elections/{ElectionID}` |
| Method | `GET`                             |
| Input  |                                   |

Return:

`200 OK` `application/json`

```json
{
  "ElectionID": "<hex encoded>",
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

# SC3: Election open 🔐

|        |                                   |
| ------ | --------------------------------- |
| URL    | `/evoting/elections/{ElectionID}` |
| Method | `PUT`                             |
| Input  | `application/json`                |

```json
{
  "Action": "open"
}
```

Return:

`200 OK` `text/plain`

```

```

# SC4: Election cast vote 🔐

|        |                                        |
| ------ | -------------------------------------- |
| URL    | `/evoting/elections/{ElectionID}/vote` |
| Method | `POST`                                 |
| Input  | `application/json`                     |

```json
{
  "UserID": "",
  "Ballot": [
    {
      "K": "<bin>",
      "C": "<bin>"
    }
  ]
}
```

Return:

`200 OK` `text/plain`

```

```

# SC5: Election close 🔐

|        |                                   |
| ------ | --------------------------------- |
| URL    | `/evoting/elections/{ElectionID}` |
| Method | `PUT`                             |
| Input  | `application/json`                |

```json
{
  "Action": "close"
}
```

Return:

`200 OK` `text/plain`

```

```

# NS2: Election shuffle 🔐

|        |                                          |
| ------ | ---------------------------------------- |
| URL    | `/evoting/services/shuffle/{ElectionID}` |
| Method | `PUT`                                    |
| Input  | `application/json`                       |

```json
{
  "Action": "shuffle"
}
```

Return:

`200 OK` `text/plain`

```

```

# SC6: Election combine shares 🔐

|        |                                   |
| ------ | --------------------------------- |
| URL    | `/evoting/elections/{ElectionID}` |
| Method | `PUT`                             |
| Input  | `application/json`                |

```json
{
  "Action": "combineShares"
}
```

Return:

`200 OK` `text/plain`

```

```

# SC?: Election cancel 🔐

|        |                                   |
| ------ | --------------------------------- |
| URL    | `/evoting/elections/{ElectionID}` |
| Method | `PUT`                             |
| Input  | `application/json`                |

```json
{
  "Action": "cancel"
}
```

Return:

`200 OK` `text/plain`

```

```

# SC?: Election delete

|         |                                   |
| ------- | --------------------------------- |
| URL     | `/evoting/elections/{ElectionID}` |
| Method  | `DELETE`                          |
| Input   |                                   |
| Headers | {Authorization: `<token>`}        |

The `<token>` value must be the hex-encoded signature of the hex-encoded
electionID:

```
<token> = hex( sig( hex( electionID ) ) )
```

Return:

`200 OK` `text/plain`

```

```

# SC?: Election get all infos

|        |                      |
| ------ | -------------------- |
| URL    | `/evoting/elections` |
| Method | `GET`                |
| Input  |                      |

Return:

`200 OK` `application/json`

```json
{
  "Elections": [
    {
      "ElectionID": "<hex encoded>",
      "Title": "",
      "Status": "",
      "Pubkey": "<hex encoded>"
    }
  ]
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
  "ElectionID": "<hex encoded>",
  "Proxy:": ""
}
```

Return:

`200 OK` `text/plain`

```

```

# DK2: DKG setup 🔐

|        |                                             |
| ------ | ------------------------------------------- |
| URL    | `/evoting/services/dkg/actors/{ElectionID}` |
| Method | `PUT`                                       |
| Input  | `application/json`                          |

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

|        |                                             |
| ------ | ------------------------------------------- |
| URL    | `/evoting/services/dkg/actors/{ElectionID}` |
| Method | `GET`                                       |
| Input  |                                             |

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

|        |                                             |
| ------ | ------------------------------------------- |
| URL    | `/evoting/services/dkg/actors/{ElectionID}` |
| Method | `PUT`                                       |
| Input  | `application/json`                          |

```json
{
  "Action": "computePubshares"
}
```

Return:

`200 OK` `text/plain`

```

```

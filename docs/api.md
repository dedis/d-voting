# API documentation

_Documentation Last Review: 11.04.2022_

## Regular workflow:

The form workflow involves 3 actors:

- Smart contract
- DKG service
- Neff shuffle service

Services are side components that augment the smart contract functionalities.
Services are accessed via the `evoting/services/<dkg>|<neff>/*` endpoint, and
the smart contract via `/evoting/forms/*`.

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
SC2:FormGetInfo



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

`200 OK` `{Status, Token}`

```json
{
  "FormID": "<hex encoded>"
}
```

# SC2: Form get info

|        |                           |
| ------ | ------------------------- |
| URL    | `/evoting/forms/{FormID}` |
| Method | `GET`                     |
| Input  |                           |

Return:

`200 OK` `{Status, Token}`

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

`200 OK` `{Status, Token}`

```

```

# SC4: Form cast vote 🔐

|        |                                |
| ------ | ------------------------------ |
| URL    | `/evoting/forms/{FormID}/vote` |
| Method | `POST`                         |
| Input  | `application/json`             |

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

`200 OK` `{Status, Token}`

```

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

`200 OK` `{Status, Token}` 

```

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

`200 OK` `{Status, Token}`

```

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

# SC?: Form cancel 🔐

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

`200 OK` `{Status, Token}`

```

```

# SC?: Form delete

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

`200 OK` `{Status, Token}`

```

```

# SC?: Form get all infos

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

# API documentation

Regular workflow:

```
Smart contract   DKG       Neff shuffle
--------------   ---       ------------
    ▼             │          │
SC1:Create        │          │
    │             │          │
    │             ▼          ▼
    │          DK1:Init   NS1:Init
    │             │
    │             ▼
    │          DK2:Setup
    │
    ▼
SC3:Open
    │
    ▼
SC4:Cast
    │
    ▼
SC5:Close
    │
    ▼
SC6:Shuffle
    │
    ▼
SC7:BeginDecryption
    │
    ▼
SC8:CombineShares
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

# SC1: Election create

|        |                      |
| ------ | -------------------- |
| URL    | `/evoting/elections` |
| Method | `POST`               |
| Input  | `application/json`   |

```json
{
  "Configuration": ""
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
  "ChunksPerBallot": "",
  "BallotSize": "",
  "Configuration": {<Configuration>}
}
```

# SC3: Election open

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

# SC4: Election cast vote

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

# SC5: Election close

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

# SC6: Election shuffle

|        |                                   |
| ------ | --------------------------------- |
| URL    | `/evoting/elections/{ElectionID}` |
| Method | `PUT`                             |
| Input  | `application/json`                |

```json
{
  "Action": "shuffle"
}
```

Return:

`200 OK` `text/plain`

```

```

# SC7: Election begin decryption

|        |                                   |
| ------ | --------------------------------- |
| URL    | `/evoting/elections/{ElectionID}` |
| Method | `PUT`                             |
| Input  | `application/json`                |

```json
{
  "Action": "beginDecryption"
}
```

Return:

`200 OK` `text/plain`

```

```

# SC8: Election combine shares

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

# SC?: Election cancel

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

# DK1: DKG init

|        |                     |
| ------ | ------------------- |
| URL    | `/evoting/dkg/init` |
| Method | `POST`              |
| Input  | `application/json`  |

```json
"<hex encoded electionID>"
```

Return:

`200 OK` `application/json`

```json
<empty>
```

# DK2: DKG setup

|        |                      |
| ------ | -------------------- |
| URL    | `/evoting/dkg/setup` |
| Method | `POST`               |
| Input  | `application/json`   |

```json
"<hex encoded electionID>"
```

Return:

`200 OK` `application/json`

```json
"<hex encoded dkg pub key>"
```

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
SC9:GetResult

```

# SC1: Election create

|        |                    |
| -      | -                  |
| URL    | `/evoting/create`  |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
    "Title": "",
    "AdminID": "",
    "Token": "",
    "Format": ""
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

|        |                    |
| -      | -                  |
| URL    | `/evoting/info`    |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
    "ElectionID": "",
    "Token": ""
}
```

# SC3: Election open

|        |                    |
| -      | -                  |
| URL    | `/evoting/open`    |
| Method | `POST`             |
| Input  | `application/json` |

```json
 "<hex encoded electionID>"
```

Return:

`200 OK` `application/json`

```json
<empty>
```

# SC4: Election cast vote

|        |                    |
| -      | -                  |
| URL    | `/evoting/cast`    |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
    "ElectionID": "",
    "UserID": "",
    "Ballot": {
        "K": "",
        "C": ""
    },
    "Token": ""
}
```

Return:

`200 OK` `application/json`

```json
{}
```

# SC5: Election close

|        |                    |
| -      | -                  |
| URL    | `/evoting/close`   |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
    "ElectionID": "",
    "UserID": ""
}
```

Return:

`200 OK` `application/json`

```json
{}
```

# SC6: Election shuffle

|        |                    |
| -      | -                  |
| URL    | `/evoting/shuffle` |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
    "ElectionID": "",
    "UserID": "",
    "Token": ""
}
```

Return:

`200 OK` `application/json`

```json
{
    "Message": ""
}
```

# SC7: Election begin decryption

|        |                            |
| -      | -                          |
| URL    | `/evoting/beginDecryption` |
| Method | `POST`                     |
| Input  | `application/json`         |

```json
{
    "ElectionID": "",
    "UserID": "",
    "Token": ""
}
```

Return:

`200 OK` `application/json`

```json
{
    "Message": ""
}
```

# SC8: Election combine shares

|        |                          |
| -      | -                        |
| URL    | `/evoting/combineShares` |
| Method | `POST`                   |
| Input  | `application/json`       |

```json
{
    "ElectionID": "",
    "UserID": "",
    "Token": ""
}
```

Return:

`200 OK` `application/json`

```json
{}
```

# SC9: Election get result

|        |                    |
| -      | -                  |
| URL    | `/evoting/result`  |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
    "ElectionID": "",
    "Token": ""
}
```

Return:

`200 OK` `application/json`

```json
{
    "Result": [
        {
            "Vote": ""
        }
    ]
}
```

# SC?: Election cancel

|        |                    |
| -      | -                  |
| URL    | `/evoting/cancel`  |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
    "ElectionID": "",
    "UserID": "",
    "Token": ""
}
```

Return:

`200 OK` `application/json`

```json
{
    "Message": ""
}
```

# SC?: Election get all infos

|        |                    |
| -      | -                  |
| URL    | `/evoting/all`     |
| Method | `POST`             |
| Input  | `application/json` |

```json
{
    "Token": ""
}
```

Return:

`200 OK` `application/json`

```json
{
    "AllElectionsInfos": [
        {
            "ElectionID": "",
            "Title": "",
            "Status": "",
            "Pubkey": "",
            "Result": [
                "Vote": ""
            ],
            "Format": ""
        }
    ]
}
```

# DK1: DKG init

|        |                     |
| -      | -                   |
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
| -      | -                    |
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

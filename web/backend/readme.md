# Installation

Once the project cloned type `npm install` to install the packages.

# Config

Copy `config.env.template` to `config.env` and replace the variables as needed.

Copy the database credentials from `config.env.template` to `/d-voting/bachend/src/.env` .

## Generate a keypair

Here is a small piece of code to help generating the keys:

```go
func GenerateKey() {
	privK := suite.Scalar().Pick(suite.RandomStream())
	pubK := suite.Point().Mul(privK, nil)

  fmt.Println("PUBLIC_KEY:", pubK)
	fmt.Println("PRIVATE_KEY:", privK)
}
```

# Run the program

Once all the previous steps done, the project can be run using `npm start`

# CLI

To control the administration right of different users, the CLI tool can be used.
For example,
```sh
$ npx cli addAdmin --sciper MY_SCIPER_NUMBER
```

You can also consult the following command for more information
```sh
$ npx cli help
Usage: cli [options] [command]

Options:
  -h, --help                     display help for command

Commands:
  addAdmin [options]             Given a SCIPER number, the owner would gain full admin permissions
  listUserPermissions [options]  Lists the permissions -if any- of the owner of a given SCIPER
  removeAdmin [options]          Given a SCIPER number, the owner would lose all admin privileges -if any-
  help [command]                 display help for command
```

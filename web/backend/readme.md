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

# Installation

Once the project cloned type `npm install` to install the packages.

# Config

Copy `config.env.template` to `config.en` and replace the variables as needed.

## Generate a keypair

Here is  a small piece of code to help generating the keys:

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

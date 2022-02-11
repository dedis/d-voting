# Installation

Once the project cloned type `npm install` to install the packages.

# Config

The project contains one file that is not in git (because it is in the .gitignore).
This file is called `config.json` and is located at the root of the express project.
Please use the `config.json.template` to start with.

This files contains all the secrets and also the running information. It should be formatted this way :

```json
{
  "FRONT_END_URL" : "<url of the current site, this is used for the tequila callback>",
  "DELA_NODE_URL" : "<url of the dela node>",
  "SESSION_SECRET" : "<session secret>",
  "PUBLIC_KEY" : "<public key>",
  "PRIVATE_KEY" : "<private key>"
}
```

Here is  a small piece of code to help generating the keys:

```go
func GenerateKey() {
	privK := suite.Scalar().Pick(suite.RandomStream())
	pubK := suite.Point().Mul(privK, nil)

  fmt.Println("PUBLIC_KEY:", pubK)
	fmt.Println("PRIVATE_KEY:", privK)
}
```

Tip: you might need to add the following line to your `/etc/hosts` file:

```
127.0.0.1       dvoting-dev.dedis.ch
```

# Run the program

Once all the previous steps done, the project can be run using `npm start`

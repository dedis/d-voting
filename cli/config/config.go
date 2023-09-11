package main

import (
	"fmt"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/util/key"
)

func main() {
	pair := key.NewKeyPair(&edwards25519.SuiteEd25519{})

	fmt.Printf("PUBLIC_KEY=%s\nPRIVATE_KEY=%s\n", pair.Public, pair.Private)
}

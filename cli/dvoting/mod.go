package main

import (
	"fmt"
	"io"
	"os"

	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/ucli"
	"go.dedis.ch/kyber/v3/suites"
)

var suite = suites.MustFind("Ed25519")

var builder cli.Builder = ucli.NewBuilder("dvoting", nil)
var printer io.Writer = os.Stderr

func main() {
	err := run(os.Args, initializer{})
	if err != nil {
		fmt.Fprintf(printer, "%+v\n", err)
	}
}

func run(args []string, inits ...cli.Initializer) error {
	for _, init := range inits {
		init.SetCommands(builder)
	}

	app := builder.Build()
	err := app.Run(args)
	if err != nil {
		return err
	}

	return nil
}

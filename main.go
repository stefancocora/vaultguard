package main

import (
	"log"

	"github.com/pkg/errors"
	"github.com/stefancocora/vaultguard/cmd"
	"github.com/stefancocora/vaultguard/pkg/version"
)

var binversion string

func main() {
	vers, err := version.Printvers()
	if err != nil {
		log.Fatalf("[FATAL] main - %v", errors.Cause(err))
	}

	binversion = vers
	log.Printf("bin: %v binver: %v pkg: %v", version.BinaryName, vers, "main")

	log.Print("starting engines")
	buildctx, err := version.BuildContext()
	if err != nil {
		log.Fatalf("[FATAL] main wrong BuildContext - %v", errors.Cause(err))
	}

	log.Printf("build context ( %v )", buildctx)

	cmd.Execute()

	log.Print("stopping engines, we're done")
}

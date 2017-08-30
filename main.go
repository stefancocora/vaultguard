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
	log.Printf("bin: %v binver: %v pkg: %v message: %v", version.BinaryName, vers, "main", "starting engines")

	buildctx, err := version.BuildContext()
	if err != nil {
		log.Fatalf("[FATAL] main wrong BuildContext - %v", errors.Cause(err))
	}

	log.Printf("bin: %v binver: %v pkg: %v message: %v ( %v )", version.BinaryName, vers, "main", "build context", buildctx)

	cmd.Execute()

	log.Printf("bin: %v binver: %v pkg: %v message: %v", version.BinaryName, vers, "main", "stopping engines, we're done")
}

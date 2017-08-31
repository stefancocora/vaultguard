/*
Copyright 2015 All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

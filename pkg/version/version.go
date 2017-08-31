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

package version

import (
	"fmt"

	"github.com/pkg/errors"
)

// Version the main version number that is being run at the moment.
var Version string

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.

// VersionPrerelease is the hardcoded version metadata "dev"/"rc" useful pre-release, and stripped pre-release
// VersionPrerelease a marker for the version number useful pre-release
var VersionPrerelease string

// GitCommit will be the git commit filled in by the go link tool via -ldflags
// package scoped global variables - versioning
var GitCommit string

// Gitbranch captures the git branch name used to build this binary
var Gitbranch string

// Gitbuilduser captures the user that has built the binaryV
var Gitbuilduser string

// Gitbuilddate captures the time when the elf was created
var Gitbuilddate string

// Buildruntime captures the version of the golang runtime used to build this binary
var Buildruntime string

// BinaryName contains the name of the binary to be used as a
// project wide variable when importing this pkg
const BinaryName = "vaultguard"

// AppEnvironment denotes the application environment: local, testing, staging, production, ...
// used to set logging context and other useful plumbing
var AppEnvironment string

// Printvers prints the version of the built elf
//
// IN:
//
// OUT
//   v0.0.1-dev-08c1e21+UNCOMMITEDCHANGES
func Printvers() (string, error) {
	if AppEnvironment == "production" {
		return fmt.Sprintf("v%v", Version), nil
	} else if AppEnvironment == "dev" {
		return fmt.Sprintf("v%v-%v-%v", Version, VersionPrerelease, GitCommit), nil
	} else {
		// log.Fatalf("[FATAL] unable to work with these application environment %v", AppEnvironment)
		return "", errors.New(fmt.Sprintf("[FATAL] unable to work with this application environment %v", AppEnvironment))
	}
}

// BuildContext prints the DVCS and go build context that have created this package
//
// IN:
//
// OUT:
//  build context ( version=v0.0.1-dev-ccaf57c+UNCOMMITEDCHANGES, appenvironment=local, runtime=go1.7.5, branch=version_controlled_from_build_script, user=stefan@workstation, date=20170211-21:12:43 )
func BuildContext() (string, error) {
	version, err := Printvers()
	if err != nil {
		return "", errors.Wrapf(err, "unable to print BuildContext, Printvers() is: \"%v\"", version)
	}
	return fmt.Sprintf("version=%v, appenvironment=%v, runtime=%v, branch=%v, user=%v, date=%v", version, AppEnvironment, Buildruntime, Gitbranch, Gitbuilduser, Gitbuilddate), nil
}

// BuildContextCli prints the DVCS and go build context that have created this package in a cli friendly way
//
// IN:
//
// OUT:
//  version:v0.0.21-dev-cd0e1d6+UNCOMMITEDCHANGES
//  appenvironment:dev
//  runtime:go1.8.3
//  branch:master
//  user:stefan@workstation
//  date=20170211-21:12:43
func BuildContextCli() (string, error) {
	version, err := Printvers()
	if err != nil {
		return "", errors.Wrapf(err, "unable to print BuildContextCli, Printvers() is: \"%v\"", version)
	}
	return fmt.Sprintf("\nversion:%v\nappenvironment:%v\nruntime:%v\nbranch:%v\nuser:%v\ndate:%v\n", version, AppEnvironment, Buildruntime, Gitbranch, Gitbuilduser, Gitbuilddate), nil
}

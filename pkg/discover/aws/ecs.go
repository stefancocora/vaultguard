package ecs

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var dbgEcsPkg bool
var dbgEcsConf bool

// docs.aws.amazon.com/sdk-for-go/api/aws/endpoints/index.html#pkg-constants
var supportedRegions = []string{
	"eu-west-1",
	"eu-west-2",
}

// AwsEcsConf contains the config needed to setup the AWs client and use it to discover vault servers running in ECS
type AwsEcsConf struct {
	Region  string
	Cluster string
}

// AwsEcsErr returns errors to upstream callers with additional information so that the callers can distinguish between permanent and temporary errors. Callers can distinguish if the error is a standard AWS error or a general error
type AwsEcsErr interface {
	error
	EcsErr() bool
	Temporary() bool
}

// AwsEcsDsc captures the format of the discovered vault endpoints
type AwsEcsDsc struct {
	Cluster      string
	VaultServers []string
}

// func (ec AwsEcsConf) listTaskDef(ctx context.Context, res chan string, err chan err) error {
func (ec AwsEcsConf) listTaskDef() error {

	// step: create a session
	sess := session.Must(session.NewSession())

	found := false
	for i := range supportedRegions {
		if supportedRegions[i] == ec.Region {
			found = true
		}
	}
	if !found {
		errm := fmt.Sprintf("ecs: unsupported region %v", ec.Region)
		return errors.New(errm)
	}

	// step: create a svc session
	ecsSvc := ecs.New(sess, aws.NewConfig().WithRegion(ec.Region))

	// go run the listTaskDef
	input := &ecs.ListTasksInput{
		Cluster: aws.String(ec.Cluster),
	}

	results, err := ecsSvc.ListTasks(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecs.ErrCodeServerException:
				// fmt.Println(ecs.ErrCodeServerException, aerr.Error())
				return aerr
			case ecs.ErrCodeClientException:
				// fmt.Println(ecs.ErrCodeClientException, aerr.Error())
				return aerr
			case ecs.ErrCodeInvalidParameterException:
				// fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
				return aerr
			case ecs.ErrCodeClusterNotFoundException:
				// fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
				return aerr
			case ecs.ErrCodeServiceNotFoundException:
				// fmt.Println(ecs.ErrCodeServiceNotFoundException, aerr.Error())
				return aerr
			default:
				// fmt.Println(aerr.Error())
				return aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			// fmt.Println(err.Error())
			return err
		}
	}
	for res := range results.TaskArns {
		log.Printf("ecs: found task arn: %#v", *results.TaskArns[res])
	}

	// return task arns
	return nil
}

// EntryPoint is the ecs pkg entrypoint
func EntryPoint(ec []AwsEcsConf) (AwsEcsDsc, error) {

	// step: list task definitions
	for i := range ec {
		if dbgEcsPkg {
			log.Printf("ecs: listing task definitions for cluster: %v", ec[i].Cluster)
		}
		if err := ec[i].listTaskDef(); err != nil {
			errm := fmt.Sprintf("ecs: error received when listing task definitions: %v", err)
			return AwsEcsDsc{}, errors.New(errm)
		}
	}

	// step: describe tasks ?

	// step: describe container instances and get priv ip ?

	return AwsEcsDsc{}, nil
}

// PropagateDebug propagates the debug flag from main into this pkg, when explicitly called
func PropagateDebug(dbg bool, confDbg bool) {
	dbgEcsPkg = dbg
	dbgEcsConf = confDbg
}

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
func (ec AwsEcsConf) listTaskDef() ([]string, error) {

	var td []string

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
		return []string{}, errors.New(errm)
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
				return []string{}, aerr
			case ecs.ErrCodeClientException:
				// fmt.Println(ecs.ErrCodeClientException, aerr.Error())
				return []string{}, aerr
			case ecs.ErrCodeInvalidParameterException:
				// fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
				return []string{}, aerr
			case ecs.ErrCodeClusterNotFoundException:
				// fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
				return []string{}, aerr
			case ecs.ErrCodeServiceNotFoundException:
				// fmt.Println(ecs.ErrCodeServiceNotFoundException, aerr.Error())
				return []string{}, aerr
			default:
				// fmt.Println(aerr.Error())
				return []string{}, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			// fmt.Println(err.Error())
			return []string{}, err
		}
	}
	if dbgEcsPkg {
		for res := range results.TaskArns {
			log.Printf("ecs: discovered task ARN: %#v", *results.TaskArns[res])
		}
	}

	// return task arns
	for res := range results.TaskArns {
		td = append(td, *results.TaskArns[res])
	}
	return td, nil
}

// describeTasks runs a DescribeTasks AWS ECS call to get the list of ECS instance ARNs
//
// IN
//
// OUT
//  []string of instance ARNs
//  []string of DescribeTasksOutput []*Failures
//  error
func (ec AwsEcsConf) describeTasks(td []string) ([]string, []ecs.Failure, error) {

	var ia []string
	var iaFailures []ecs.Failure

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
		return []string{}, []ecs.Failure{}, errors.New(errm)
	}

	// step: create a svc session
	svc := ecs.New(sess, aws.NewConfig().WithRegion(ec.Region))

	// step: prepare inputs
	var tsk []*string

	for i := range td {
		tsk = append(tsk, &td[i])
	}
	input := &ecs.DescribeTasksInput{
		Cluster: &ec.Cluster,
		Tasks:   tsk,
	}

	result, err := svc.DescribeTasks(input)
	// complete failure cases
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecs.ErrCodeServerException:
				// fmt.Println(ecs.ErrCodeServerException, aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeClientException:
				// fmt.Println(ecs.ErrCodeClientException, aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeInvalidParameterException:
				// fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeClusterNotFoundException:
				// fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			default:
				// fmt.Println(aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			// fmt.Println(err.Error())
			return []string{}, []ecs.Failure{}, aerr
		}
	}

	if dbgEcsPkg {
		// extra debug
		// log.Printf("results: %v", result.Tasks)
		for res := range result.Tasks {
			log.Printf("ecs: discovered container instance ARN: %#v", *result.Tasks[res].ContainerInstanceArn)
		}
	}

	// step: return instance ARNs
	for res := range result.Tasks {
		ia = append(ia, *result.Tasks[res].ContainerInstanceArn)
	}
	for f := range result.Failures {
		iaFailures = append(iaFailures, *result.Failures[f])
	}
	// partial failures test
	if len(iaFailures) == 0 {
		return ia, []ecs.Failure{}, nil
	}
	return ia, iaFailures, nil
}

// describeContInst interogates the DescribeContainerInstances AWS ECS API endpoint to retrieve the container instance ids
//
// IN
//
// OUT
func (ec AwsEcsConf) describeContInst(ia []string) ([]string, []ecs.Failure, error) {

	var iid []string
	var iidFailures []ecs.Failure

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
		return []string{}, []ecs.Failure{}, errors.New(errm)
	}

	// step: create a svc session
	svc := ecs.New(sess, aws.NewConfig().WithRegion(ec.Region))

	// step: prepare inputs
	var arns []*string

	for i := range ia {
		arns = append(arns, &ia[i])
	}

	input := &ecs.DescribeContainerInstancesInput{
		Cluster:            &ec.Cluster,
		ContainerInstances: arns,
	}

	result, err := svc.DescribeContainerInstances(input)
	// complete failure cases
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecs.ErrCodeServerException:
				// fmt.Println(ecs.ErrCodeServerException, aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeClientException:
				// fmt.Println(ecs.ErrCodeClientException, aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeInvalidParameterException:
				// fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeClusterNotFoundException:
				// fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			default:
				// fmt.Println(aerr.Error())
				return []string{}, []ecs.Failure{}, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			// fmt.Println(err.Error())
			return []string{}, []ecs.Failure{}, aerr
		}
	}

	if dbgEcsPkg {
		// extra debug
		// log.Printf("results: %v", result.ContainerInstances)
		for res := range result.ContainerInstances {
			log.Printf("ecs: discovered container id: %#v", *result.ContainerInstances[res].Ec2InstanceId)
		}
	}

	// step: return instance ids
	for res := range result.ContainerInstances {
		iid = append(iid, *result.ContainerInstances[res].Ec2InstanceId)
	}
	// partial failures test
	if len(iidFailures) == 0 {
		return iid, []ecs.Failure{}, nil
	}

	// log.Printf("all instance ids: %v", result)
	// log.Println(iid)
	// log.Println(iidFailures)
	return iid, iidFailures, nil
}

// Discover is the ecs pkg entrypoint
func Discover(ec []AwsEcsConf) (AwsEcsDsc, error) {

	for i := range ec {
		if dbgEcsPkg {
			log.Printf("ecs: listing task definitions for cluster: %v", ec[i].Cluster)
		}

		// step: get task arns
		td, err := ec[i].listTaskDef()
		if err != nil {
			errm := fmt.Sprintf("ecs: error received when listing task definitions: %v", err)
			return AwsEcsDsc{}, errors.New(errm)
		}

		if dbgEcsPkg {
			log.Printf("ecs: listing ECS instance ARNs for cluster: %v", ec[i].Cluster)
		}
		// step: get ECS instance arns
		ia, iaf, err := ec[i].describeTasks(td)
		if err != nil {
			errm := fmt.Sprintf("ecs: error received when describing instance ARNs: %v", err)
			return AwsEcsDsc{}, errors.New(errm)
		}
		if len(iaf) != 0 {
			log.Printf("ecs: partial failures when running DescribeTasks(): %v", iaf)
		}

		// step: describe container instances and get instance id
		iid, iipsf, err := ec[i].describeContInst(ia)
		if err != nil {
			errm := fmt.Sprintf("ecs: error received when describing container instances: %v", err)
			return AwsEcsDsc{}, errors.New(errm)
		}
		if len(iipsf) != 0 {
			log.Printf("ecs: partial failures when running DescribeContainerInstances(): %v", iipsf)
		}

		log.Printf("all instance ids: %v", iid)

		// this should return the struct of discovered vault servers
		return AwsEcsDsc{}, nil
	}

	return AwsEcsDsc{}, nil
}

// PropagateDebug propagates the debug flag from main into this pkg, when explicitly called
func PropagateDebug(dbg bool, confDbg bool) {
	dbgEcsPkg = dbg
	dbgEcsConf = confDbg
}

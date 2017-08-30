package ecs

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var dbgEcsPkg bool
var dbgEcsConf bool
var dbgAwsResp bool

// docs.aws.amazon.com/sdk-for-go/api/aws/endpoints/index.html#pkg-constants
var supportedRegions = []string{
	"eu-west-1",
	"eu-west-2",
}

// AwsEcsInput contains the config needed to setup the AWs client and use it to discover vault servers running in ECS
type AwsEcsInput struct {
	Region  string
	Cluster string
}

// AwsEcsErr returns errors to upstream callers with additional information so that the callers can distinguish between permanent and temporary errors. Callers can distinguish if the error is a standard AWS error or a general error
type AwsEcsErr interface {
	error
	EcsErr() bool
	Temporary() bool
}

// AwsEcsOutput captures the format of the discovered vault endpoints
type AwsEcsOutput struct {
	Cluster      string
	VaultServers []VaultSrvOutput
	Fault        []error
}

// VaultSrvOutput holds the definition of a single vault endpoint
type VaultSrvOutput struct {
	IP   string
	Port string
}

// dscInstIP holds the definition of a single discovered ECS instance with its instance id and private IP
type dscInstIP struct {
	ID     string
	PrivIP string
}

// dscInstPort holds the definition of a single discovered ECS instance with its instance id and port
type dscInstPort struct {
	ID   string
	Port string
}

// descTaskOutput holds the definition of a discovered task
type descTaskOutput struct {
	iarn string
	port int64
}

// descECSInstOutput holds the definition of a discovered ECS instance
type descECSInstOutput struct {
	iarn string
	iid  string
}

// descEC2InstOutput holds the definition of a discovered EC2 instance
type descEC2InstOutput struct {
	iid     string
	iprivip string
}

// listTaskDef returns the ECS task ARN
//
// IN
//
//  AwsEcsInput with a cluster and region
//
// OUT
//
//  []string == task ARNs
func (ec AwsEcsInput) listTaskDef() ([]string, error) {

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

	result, err := ecsSvc.ListTasks(input)
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
		if dbgAwsResp {
			var temp []string
			for i := range result.TaskArns {
				ts := *result.TaskArns[i]
				temp = append(temp, ts)
			}
			log.Printf("ecs ListTasks: %v", temp)
		}
		for res := range result.TaskArns {
			log.Printf("ecs: discovered task ARN: %#v", *result.TaskArns[res])
			td = append(td, *result.TaskArns[res])
		}
	} else {
		for res := range result.TaskArns {
			td = append(td, *result.TaskArns[res])
		}
	}

	// return task arns
	return td, nil
}

// describeTasks runs a DescribeTasks AWS ECS call to get the list of ECS instance ARNs
//
// IN
//
//  []string of ECS task definition ARNs
//
// OUT
//  []descTaskOutput of task ARNs
//  []string of DescribeTasksOutput []*Failures
//  error
func (ec AwsEcsInput) describeTasks(td []string) ([]descTaskOutput, []ecs.Failure, error) {

	var ia []descTaskOutput
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
		return []descTaskOutput{}, []ecs.Failure{}, errors.New(errm)
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
				return []descTaskOutput{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeClientException:
				// fmt.Println(ecs.ErrCodeClientException, aerr.Error())
				return []descTaskOutput{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeInvalidParameterException:
				// fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
				return []descTaskOutput{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeClusterNotFoundException:
				// fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
				return []descTaskOutput{}, []ecs.Failure{}, aerr
			default:
				// fmt.Println(aerr.Error())
				return []descTaskOutput{}, []ecs.Failure{}, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			// fmt.Println(err.Error())
			return []descTaskOutput{}, []ecs.Failure{}, aerr
		}
	}

	if dbgEcsPkg {
		if dbgAwsResp {
			// extra debug
			log.Printf("ecs DescribeTasks: %v", result.Tasks)
		}
		for res := range result.Tasks {
			log.Printf("ecs: discovered container instance ARN: %#v", *result.Tasks[res].ContainerInstanceArn)
			var io descTaskOutput
			io.iarn = *result.Tasks[res].ContainerInstanceArn
			for i := range result.Tasks[res].Containers {
				for j := range result.Tasks[res].Containers[i].NetworkBindings {
					io.port = *result.Tasks[res].Containers[i].NetworkBindings[j].HostPort
				}
			}
			ia = append(ia, io)
		}
	} else {
		for res := range result.Tasks {
			var io descTaskOutput
			io.iarn = *result.Tasks[res].ContainerInstanceArn
			for i := range result.Tasks[res].Containers {
				for j := range result.Tasks[res].Containers[i].NetworkBindings {
					io.port = *result.Tasks[res].Containers[i].NetworkBindings[j].HostPort
				}
			}
			ia = append(ia, io)
		}
	}

	// step: return instance ARNs
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
//  []string of ECS instance ARNs
//
// OUT
//  []descECSInstanceOutput of instance ARNs
//  []string of DescribeContainerInstances []*Failures
//  error
func (ec AwsEcsInput) describeContInst(ia []string) ([]descECSInstOutput, []ecs.Failure, error) {

	var iid []descECSInstOutput
	var iidFailures []ecs.Failure

	if dbgEcsConf {
		log.Printf("received input: %v", ia)
	}

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
		return []descECSInstOutput{}, []ecs.Failure{}, errors.New(errm)
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
				return []descECSInstOutput{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeClientException:
				// fmt.Println(ecs.ErrCodeClientException, aerr.Error())
				return []descECSInstOutput{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeInvalidParameterException:
				// fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
				return []descECSInstOutput{}, []ecs.Failure{}, aerr
			case ecs.ErrCodeClusterNotFoundException:
				// fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
				return []descECSInstOutput{}, []ecs.Failure{}, aerr
			default:
				// fmt.Println(aerr.Error())
				return []descECSInstOutput{}, []ecs.Failure{}, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			// fmt.Println(err.Error())
			return []descECSInstOutput{}, []ecs.Failure{}, aerr
		}
	}

	if dbgEcsPkg {
		if dbgAwsResp {
			// extra debug
			log.Printf("ecs DescribeContainerInstances: %v", result.ContainerInstances)
		}
		for res := range result.ContainerInstances {
			log.Printf("ecs: discovered container instance ID: %#v", *result.ContainerInstances[res].Ec2InstanceId)
			var t descECSInstOutput
			t.iid = *result.ContainerInstances[res].Ec2InstanceId
			t.iarn = *result.ContainerInstances[res].ContainerInstanceArn
			iid = append(iid, t)
		}
		// step: return instance ids
	} else {
		for res := range result.ContainerInstances {
			var t descECSInstOutput
			t.iid = *result.ContainerInstances[res].Ec2InstanceId
			t.iarn = *result.ContainerInstances[res].ContainerInstanceArn
			iid = append(iid, t)
		}
	}

	// partial failures test
	if len(iidFailures) == 0 {
		return iid, []ecs.Failure{}, nil
	}

	return iid, iidFailures, nil
}

// describeEC2Inst interogates the DescribeInstances AWS EC2 API endpoint to retrieve the container instance private ips
//
// IN
//
//  []string of ECS instance ids
//
// OUT
//  []descEC2InstOutput of instance priv ips
//  error
func (ec AwsEcsInput) describeEC2Inst(iid []string) ([]descEC2InstOutput, error) {

	var iprivip []descEC2InstOutput

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
		return []descEC2InstOutput{}, errors.New(errm)
	}

	// step: create a svc session
	svc := ec2.New(sess, aws.NewConfig().WithRegion(ec.Region))

	// step: prepare inputs

	var ii []*string
	for i := range iid {
		ii = append(ii, &iid[i])
	}

	// we want only the running instances
	filn := "instance-state-name"
	ist := "running"
	filter := ec2.Filter{
		Name: &filn,
		Values: []*string{
			&ist,
		},
	}
	f := []*ec2.Filter{
		&filter,
	}

	input := &ec2.DescribeInstancesInput{
		Filters:     f,
		InstanceIds: ii,
	}

	result, err := svc.DescribeInstances(input)

	// complete failure cases
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				// fmt.Println(aerr.Error())
				return []descEC2InstOutput{}, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			// fmt.Println(err.Error())
			return []descEC2InstOutput{}, aerr
		}
	}

	if dbgEcsPkg {
		if dbgAwsResp {
			// extra debug
			log.Printf("ec2 DescribeInstances: %v", result.Reservations)
		}
		for res := range result.Reservations {
			for i := range result.Reservations[res].Instances {
				for j := range result.Reservations[res].Instances[i].NetworkInterfaces {
					if dbgAwsResp {
						// extra debug
						log.Printf("ecs: discovered container instance private ips: %#v", *result.Reservations[res].Instances[i].NetworkInterfaces[j].PrivateIpAddresses[0])
						var t descEC2InstOutput
						t.iprivip = *result.Reservations[res].Instances[i].NetworkInterfaces[j].PrivateIpAddresses[0].PrivateIpAddress
						t.iid = *result.Reservations[res].Instances[i].InstanceId
						iprivip = append(iprivip, t)
					} else {
						log.Printf("ecs: discovered container instance private ips: %#v", *result.Reservations[res].Instances[i].NetworkInterfaces[j].PrivateIpAddresses[0].PrivateIpAddress)
						var t descEC2InstOutput
						t.iprivip = *result.Reservations[res].Instances[i].NetworkInterfaces[j].PrivateIpAddresses[0].PrivateIpAddress
						t.iid = *result.Reservations[res].Instances[i].InstanceId
						iprivip = append(iprivip, t)
					}
				}
			}
		}
	} else {
		for res := range result.Reservations {
			for i := range result.Reservations[res].Instances {
				for j := range result.Reservations[res].Instances[i].NetworkInterfaces {
					var t descEC2InstOutput
					t.iprivip = *result.Reservations[res].Instances[i].NetworkInterfaces[j].PrivateIpAddresses[0].PrivateIpAddress
					t.iid = *result.Reservations[res].Instances[i].InstanceId
					iprivip = append(iprivip, t)
				}
			}
		}
	}

	// step: return instance priv ips
	return iprivip, nil
}

// Discover is used as a way to discover vault endpoints in ECS starting from a cluster name and region
//
// IN
//
// []AwsEcsInput of cluster and region names
//
// OUT
//
// []AwsEcsOutput of formatted vault endpoints and cluster names
//
func Discover(ec []AwsEcsInput) []AwsEcsOutput {

	var rdve []AwsEcsOutput

	for i := range ec {
		var dve AwsEcsOutput
		log.Printf("ecs: listing task definitions for cluster: %v", ec[i].Cluster)
		dve.Cluster = ec[i].Cluster

		// step: get task arns
		td, err := ec[i].listTaskDef()
		if err != nil {
			errm := fmt.Sprintf("ecs: error received when listing task definitions: %v", err)
			dve.Fault = append(dve.Fault, errors.New(errm))
		}

		if dbgEcsPkg {
			log.Printf("ecs: listing ECS instance ARNs for cluster: %v", ec[i].Cluster)
		}
		// step: get ECS instance arns
		ia, iaf, err := ec[i].describeTasks(td)
		if err != nil {
			errm := fmt.Sprintf("ecs: error received when describing instance ARNs: %v", err)
			dve.Fault = append(dve.Fault, errors.New(errm))
		}
		if len(iaf) != 0 {
			log.Printf("ecs: partial failures when running DescribeTasks(): %v", iaf)
		}

		// step: get instance ids
		var tia []string
		for i := range ia {
			tia = append(tia, ia[i].iarn)
		}
		iid, iipsf, err := ec[i].describeContInst(tia)
		if err != nil {
			errm := fmt.Sprintf("ecs: error received when describing container instances: %v", err)
			dve.Fault = append(dve.Fault, errors.New(errm))
		}
		if len(iipsf) != 0 {
			log.Printf("ecs: partial failures when running DescribeContainerInstances(): %v", iipsf)
		}

		// step: get instance privips
		var tiid []string
		for i := range iid {
			tiid = append(tiid, iid[i].iid)
		}
		iprivi, err := ec[i].describeEC2Inst(tiid)
		if err != nil {
			errm := fmt.Sprintf("ecs: error received when describing ec2 instances: %v", err)
			dve.Fault = append(dve.Fault, errors.New(errm))
		}

		for i := range iprivi {
			var ts VaultSrvOutput
			ts.IP = iprivi[i].iprivip
			for j := range iid {
				if iprivi[i].iid == iid[j].iid {
					for k := range ia {
						if iid[j].iarn == ia[k].iarn {
							ts.Port = strconv.FormatInt(ia[k].port, 10)
							dve.VaultServers = append(dve.VaultServers, ts)
						}
					}
				}
			}
		}

		// this should return the discovered vault servers
		if dbgEcsPkg {
			log.Printf("all discovered vault endpoints: %v", dve)
		}
		rdve = append(rdve, dve)
	}

	return rdve
}

// PropagateDebug propagates the debug flag from main into this pkg, when explicitly called
func PropagateDebug(dbg bool, confDbg bool) {
	dbgEcsPkg = dbg
	dbgEcsConf = confDbg
	// compile time flag for now
	dbgAwsResp = false
}

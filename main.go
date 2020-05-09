package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func runInstances(ec2client *ec2.EC2) (*ec2.Reservation, error) {
	instancesinput := &ec2.RunInstancesInput{
		TagSpecifications: []*ec2.TagSpecification{
			&ec2.TagSpecification{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					&ec2.Tag{
						Key:   aws.String("Name"),
						Value: aws.String("test"),
					},
				},
			},
		},
		ImageId:  aws.String("ami-0323c3dd2da7fb37d"),
		MinCount: aws.Int64(1),
		MaxCount: aws.Int64(1),
	}
	instances, err := ec2client.RunInstances(instancesinput)
	return instances, err
}

func getInstances(r *ec2.DescribeInstancesOutput) []*string {
	instances := []*string{}
	for i := 0; i < len(r.Reservations); i++ {
		id := *r.Reservations[i].Instances[0].InstanceId
		instances = append(instances, aws.String(id))
	}
	return instances
}

func printInstances(s []*string) {
	if len(s) < 1 {
		fmt.Println("No instances are running.")
	} else {
		fmt.Println("Instances running:", aws.StringValueSlice(s))
	}
}

func main() {
	sess, err := session.NewSession(
		&aws.Config{Region: aws.String("us-east-1")},
	)
	if err != nil {
		fmt.Println("Error creating session ", err)
		return
	}
	ec2client := ec2.New(sess)

	d := &ec2.DescribeInstancesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("pending"),
					aws.String("running"),
				},
			},
			&ec2.Filter{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String("test"),
				},
			},
		},
	}
	reservations, err := ec2client.DescribeInstances(d)
	instanceIds := getInstances(reservations)
	printInstances(instanceIds)

	//////////////////////////
	//  Creating instances  //
	//////////////////////////

	fmt.Println("Creating new EC2 instance...")
	_, err = runInstances(ec2client)
	if err != nil {
		fmt.Println("Error creating instance ", err)
		return
	}

	// Refresh instances list
	time.Sleep(10 * time.Second)
	reservations, err = ec2client.DescribeInstances(d)
	instanceIds = getInstances(reservations)
	printInstances(instanceIds)
	time.Sleep(10 * time.Second)

	/////////////////////////////
	//  Terminating instances  //
	/////////////////////////////

	fmt.Println("Terminating test EC2 instances: ", aws.StringValueSlice(instanceIds))
	terminateinstancesinput := &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds,
	}
	_, err = ec2client.TerminateInstances(terminateinstancesinput)
	if err != nil {
		fmt.Println("Error terminating instances", err)
		return
	}
}

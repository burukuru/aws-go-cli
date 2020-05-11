package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/urfave/cli/v2"
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
		fmt.Println(id)
		instances = append(instances, aws.String(id))

	}
	return instances
}

func printInstances(s []*string, tagkv string) {
	if len(s) < 1 {
		log.Print("No running instances with specified tag \"", tagkv, "\" found.")
	} else {
		fmt.Println("Instances running:", aws.StringValueSlice(s))
	}
}

func createClient() *ec2.EC2 {
	var region = "us-east-1"
	sess, err := session.NewSession(
		&aws.Config{Region: aws.String(region)},
	)
	if err != nil {
		fmt.Println("Error creating session ", err)
		panic(err)
	}
	return ec2.New(sess)
}

func describeInstances(ec2client *ec2.EC2, tagkv string) ([]*string, error) {
	t := strings.Split(tagkv, "=")
	tagkey := strings.Join([]string{"tag:", t[0]}, "")
	tagvalue := t[1]
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
				Name: aws.String(tagkey),
				Values: []*string{
					aws.String(tagvalue),
				},
			},
		},
	}
	reservations, err := ec2client.DescribeInstances(d)
	if err != nil {
		log.Fatal(err)
	}
	instanceIds := getInstances(reservations)
	printInstances(instanceIds, tagkv)

	return instanceIds, err
}

func createInstance(ec2client *ec2.EC2) error {
	fmt.Println("Creating new EC2 instance...")
	_, err := runInstances(ec2client)
	if err != nil {
		log.Fatal("Error creating instance ", err)
	}
	return err
}

func terminateinstances(ec2client *ec2.EC2, instanceIds []*string) {
	fmt.Println("Terminating test EC2 instances: ", aws.StringValueSlice(instanceIds))
	terminateinstancesinput := &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds,
	}
	_, err := ec2client.TerminateInstances(terminateinstancesinput)
	if err != nil {
		fmt.Println("Error terminating instances", err)
		panic(err)
	}
}

func createDestroyInstance(ec2client *ec2.EC2) error {
	describeInstances(ec2client, "Name=test")
	//////////////////////////
	//  Creating instances  //
	//////////////////////////
	createInstance(ec2client)

	// Refresh instances list
	time.Sleep(10 * time.Second)
	instanceIds, err := describeInstances(ec2client, "Name=test")
	if err != nil {
		log.Fatal(err)
	}

	/////////////////////////////
	//  Terminating instances  //
	/////////////////////////////
	time.Sleep(10 * time.Second)
	terminateinstances(ec2client, instanceIds)

	return err
}

func main() {
	ec2client := createClient()

	app := &cli.App{
		Name:  "aws-go-cli",
		Usage: "AWS CLI wrapper in Go",
		Commands: []*cli.Command{
			{
				Name:    "describe-instances",
				Aliases: []string{"di"},
				Usage:   "List EC2 instances in selected region",
				Action: func(c *cli.Context) error {
					_, err := describeInstances(ec2client, "Name=test")
					if err != nil {
						log.Fatal(err)
					}
					return err
				},
			},
			{
				Name:    "create-instance",
				Aliases: []string{"ci"},
				Usage:   "Create test EC2 instance",
				Action: func(c *cli.Context) error {
					err := createInstance(ec2client)
					if err != nil {
						log.Fatal(err)
					}
					return err
				},
			},
			{
				Name:    "terminate-instances",
				Aliases: []string{"ti"},
				Usage:   "Terminate test EC2 instances",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "tag",
						Value: "Name=test",
						Usage: "Filter tag of EC2 instances to terminate in format: `TagName=TagValue`",
					},
				},
				Action: func(c *cli.Context) error {
					instanceIds, err := describeInstances(ec2client, c.String("tag"))
					if err != nil {
						log.Fatal(err)
					}
					if len(instanceIds) < 1 {
						log.Fatal("No running instances with specified tag \"", c.String("tag"), "\" to terminate.")
					}
					terminateinstances(ec2client, instanceIds)
					return err
				},
			},
			{
				Name:    "run-test",
				Aliases: []string{"rt"},
				Usage:   "Run full test cycle: create, list, terminate test instance",
				Action: func(c *cli.Context) error {
					err := createDestroyInstance(ec2client)
					if err != nil {
						log.Fatal(err)
					}
					return err
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

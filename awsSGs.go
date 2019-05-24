package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getSecurityGroups(sgIds []*string) []*ec2.SecurityGroup {
	svc := ec2.New(session.New(&aws.Config{
		Region: aws.String("us-east-1")},
	))
	_, err := svc.Config.Credentials.Get()
	if err != nil {
		log.Fatalln(err)
	}
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: sgIds,
	}
	result, err := svc.DescribeSecurityGroups(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				panic(err.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			panic("Error getting Security Group with the provided Id.")
		}
	} else {
		return result.SecurityGroups
	}
}

func dropRuleFromSg(ipsToDrop []string, groupID string) {
	svc := ec2.New(session.New(&aws.Config{
		Region: aws.String("us-east-1")},
	))
	for _, cidr := range ipsToDrop {
		_, err := svc.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
			GroupId: aws.String(groupID),
			IpPermissions: []*ec2.IpPermission{
				// Can use setters to simplify seting multiple values without the
				// needing to use aws.String or associated helper utilities.
				(&ec2.IpPermission{}).
					SetIpProtocol("tcp").
					SetFromPort(443).
					SetToPort(443).
					SetIpRanges([]*ec2.IpRange{
						(&ec2.IpRange{}).
							SetCidrIp(cidr),
					}),
			},
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
	fmt.Printf("Successfully removed invalid CIDRs: %s", ipsToDrop)
}

func addRuleToSg(ipRangesToAdd []string, groupID string) {
	svc := ec2.New(session.New(&aws.Config{
		Region: aws.String("us-east-1")},
	))
	for _, cidr := range ipRangesToAdd {
		fmt.Printf("Attempting to add the following CIDR to '%s': %s\n", groupID, cidr)
		_, err := svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(groupID),
			IpPermissions: []*ec2.IpPermission{
				// Can use setters to simplify seting multiple values without the
				// needing to use aws.String or associated helper utilities.
				(&ec2.IpPermission{}).
					SetIpProtocol("tcp").
					SetFromPort(443).
					SetToPort(443).
					SetIpRanges([]*ec2.IpRange{
						(&ec2.IpRange{}).
							SetCidrIp(cidr),
					}),
			},
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
	fmt.Printf("Successfully updated Security Group with additional CIDRs:      %s\n\n", ipRangesToAdd)
}

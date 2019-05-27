package job

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var svc = ec2.New(session.New(&aws.Config{
	Region: aws.String("us-east-1")},
))

func getSecurityGroups(sgIds []*string) ([]*ec2.SecurityGroup, error) {
	_, err := svc.Config.Credentials.Get()
	if err != nil {
		fmt.Println("Encountered error while checking for aws credentials")
		return nil, err
	}
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: sgIds,
	}
	result, err := svc.DescribeSecurityGroups(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			fmt.Println("Error getting Security Group(s) with the provided input.")
			return nil, err
		}
	} else {
		return result.SecurityGroups, nil
	}
}

func dropRuleFromSg(ipsToDrop []string, groupID string) error {
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
			return err
		}
	}
	fmt.Printf("Successfully removed invalid CIDRs: %s\n", ipsToDrop)
	return nil
}

func addRuleToSg(ipRangesToAdd []string, groupID string) error {
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
			return err
		}
	}
	fmt.Printf("Successfully updated Security Group with additional CIDRs:      %s\n\n", ipRangesToAdd)
	return nil
}

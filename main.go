package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func githubHookCIDRs() []string {
	resp, err := http.Get("https://api.github.com/meta")
	if err != nil {
		log.Fatalln("Error loading Github CIDRs")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var data struct {
		Hooks []string //we only really care about the Hooks
	}
	json.Unmarshal(body, &data)
	fmt.Printf("Valid 'Hooks' CIDRs response from https://api.github.com/meta:\n	%s\n\n\n", data.Hooks)
	return data.Hooks
}

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

// Checks if the provided string "item" exists in a provided "slice"
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// Converts given string of security group ids to a slice
// Handles both comma separated and comma-space separated
func regSplitEnv(envVar string) []*string {
	re := regexp.MustCompile("(, )|(,)")
	split := re.Split(envVar, -1)
	set := []*string{}
	for i := range split {
		set = append(set, &split[i])
	}
	return set
}

func main() {
	hooks := githubHookCIDRs()
	sgsToCheck := regSplitEnv(os.Getenv("SG_IDS"))
	for _, sg := range getSecurityGroups(sgsToCheck) {
		currentRules := sg.IpPermissions

		//currentRules := getSecurityGroups().IpPermissions
		var activeSgCIDRs []string
		var invalidSgCIDRs []string
		for _, rule := range currentRules {
			for _, ipRange := range rule.IpRanges {
				activeSgCIDRs = append(activeSgCIDRs, *ipRange.CidrIp)

				// Checks if rule exists in sg but not a valid github ip
				if contains(hooks, *ipRange.CidrIp) != true {
					invalidSgCIDRs = append(invalidSgCIDRs, *ipRange.CidrIp)
				}
			}
		}

		// Checks if there are any missing github ip's in the security group
		var cidrsToAdd []string
		for _, v := range hooks {
			if contains(activeSgCIDRs, v) != true {
				cidrsToAdd = append(cidrsToAdd, v)
			}
		}
		fmt.Println("Checking for missing Github 'Hooks' CIDRs...\n")
		if len(cidrsToAdd) > 0 {
			fmt.Printf("Currently Active Security Group CIDRS: %s\n", activeSgCIDRs)
			fmt.Printf("Valid Github CIDRs:                    %s\n", hooks)
			fmt.Printf("Missing Github 'Hooks' CIDRs:          %s\n\n", cidrsToAdd)
			addRuleToSg(cidrsToAdd, *sg.GroupId)
			for _, v := range cidrsToAdd {
				activeSgCIDRs = append(activeSgCIDRs, v)
			}
		} else {
			fmt.Println("No extra CIDRs to add, all good ༼つ▀̿_▀̿ ༽つ")
		}

		if len(invalidSgCIDRs) > 0 {
			fmt.Printf("Currently Active Security Group CIDRS: %s\n", activeSgCIDRs)
			fmt.Printf("Valid Github 'Hooks' CIDRs:            %s\n", hooks)
			fmt.Printf("Invalid CIDRs on Security Group %s:\n  %s\n", *sg.GroupId, invalidSgCIDRs)
			dropRuleFromSg(invalidSgCIDRs, *sg.GroupId)
		} else {
			fmt.Println("No invalid CIDRs to drop, all good ༼つ▀̿_▀̿ ༽つ")
		}
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

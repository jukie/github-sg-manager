package job

import (
	"fmt"
	"log"
	"os"
	"regexp"
)

// Checks if the provided string "item" exists in the provided "slice"
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
func Execute() error {
	hooks, err := githubHookCIDRs()
	if err != nil {
		log.Println("Encounted error while making request for Github 'Hooks' CIDRs")
		return err
	}

	if len(os.Getenv("SECURITY_GROUP_IDS")) == 0 {
		log.Println("No security group ids provided, exiting")
	}
	sgsToCheck := regSplitEnv(os.Getenv("SECURITY_GROUP_IDS"))
	sgResults, err := getSecurityGroups(sgsToCheck)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, sg := range sgResults {
		currentRules := sg.IpPermissions
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
			err = addRuleToSg(cidrsToAdd, *sg.GroupId)
			if err != nil {
				fmt.Println("Error adding rules to Security Group")
				return err
			}
			for _, v := range cidrsToAdd {
				activeSgCIDRs = append(activeSgCIDRs, v)
			}
		} else {
			fmt.Println("No extra CIDRs to add, all good ༼つ▀̿_▀̿ ༽つ")
		}
		fmt.Println("Checking for invalid Security Group CIDRs...\n")
		if len(invalidSgCIDRs) > 0 {
			fmt.Printf("Currently Active Security Group CIDRS: %s\n", activeSgCIDRs)
			fmt.Printf("Valid Github 'Hooks' CIDRs:            %s\n", hooks)
			fmt.Printf("Invalid CIDRs on Security Group %s:\n  %s\n", *sg.GroupId, invalidSgCIDRs)
			err = dropRuleFromSg(invalidSgCIDRs, *sg.GroupId)
			if err != nil {
				fmt.Println("Error adding rules to Security Group")
				return err
			}
		} else {
			fmt.Println("No invalid CIDRs to drop, all good ༼つ▀̿_▀̿ ༽つ")
		}
	}
	return nil
}

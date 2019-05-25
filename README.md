# github-sg-manager
AWS Lambda function to allow inbound Github Webhooks via security group

Lambda function to automate updating a Security Group of GitHub's hooks IP addresses.
I use this function to allow inbound Github webhooks to my ci/cd environments


## Environment variables

* `SECURITY_GROUP_IDS` - Comma separated list of Security Groups to make changes to 
  *  Valid values: 
     * `sg-example0123456789, sg-example1234567890`
     * `sg-example0123456789,sg-example1234567890`

## Code walkthrogh
Top-level files:
* `lambda.go` - Meant for AWS Lambda Function, contains lambda execution call to job.Execute() 
* `local.go` - Used to execute code locally and calls job.Execute() directly

Execution flow:
1. Performs an http GET call to https://api.github.com/meta, obtains the 'hooks' CIDRs, and prints them to STDOUT
     ```
     Valid 'Hooks' CIDRs response from https://api.github.com/meta:
          [192.30.252.0/22 185.199.108.0/22 140.82.112.0/20]
     ```
2. For the given Security Group(s), lookup the active rules and check for invalid ones(i.e ones that don't exist in the 'hooks' CIDRs)
3. Checks for missing rules by comparing the active rules to the 'hooks' CIDRs
4. * Add any missing CIDRs
     ```
     Checking for missing Github 'Hooks' CIDRs...

     Currently Active Security Group CIDRS: [192.30.252.0/22 185.199.108.0/22 0.0.0.0/0]
     Valid Github CIDRs:                    [192.30.252.0/22 185.199.108.0/22 140.82.112.0/20]
     Missing Github 'Hooks' CIDRs:          [140.82.112.0/20]

     Attempting to add the following CIDR to 'sg-example0123456789': 140.82.112.0/20
     Successfully updated Security Group with additional CIDRs:      [140.82.112.0/20]
     ```
   * Or else take no action and print a cool emoticon
     ```
     Checking for missing Github 'Hooks' CIDRs...

     No extra CIDRs to add, all good ༼つ▀̿_▀̿ ༽つ
     ```
5. * Remove any invalid CIDRs
     ```
     Checking for invalid Security Group CIDRs...

     Currently Active Security Group CIDRS: [192.30.252.0/22 185.199.108.0/22 0.0.0.0/0 140.82.112.0/20]
     Valid Github 'Hooks' CIDRs:            [192.30.252.0/22 185.199.108.0/22 140.82.112.0/20]
     Invalid CIDRs on Security Group sg-example0123456789:
     [0.0.0.0/0]
     Successfully removed invalid CIDRs: [0.0.0.0/0]
     ```
   * Or take no action and print another cool emoticon
     ```
     Checking for invalid Security Group CIDRs...

     No invalid CIDRs to drop, all good ༼つ▀̿_▀̿ ༽つ
     ``` 
Build lambda deployment package:
* `make build` - Produces a binary and zip package within `./build/` named `hookCIDRs` and `hookCIDRs.zip`. This can then be used in the lambda function with a handler value of `hookCIDRs`.

Cleanup build artifacts:
* `make clean` - Removes the referenced compiled binary and zip file from `./build/`
## To do list:

* More docs
* Add Terraform examples for initial deployment
* Add support for any port (currently defaults 443)
* Integration tests
* Code review? Tell me what sucks

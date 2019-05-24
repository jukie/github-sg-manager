# github-hooks-cidrs-lambda
AWS Lambda function to allow inbound Github Webhooks via security group

Lambda function to automate updating a Security Group of GitHub's hooks IP addresses.
I use this function to allow inbound Github webhooks to my ci/cd environments


## Environment variables

* `SECURITY_GROUP_IDS` - Comma separated list of Security Groups to make changes to 
  *  Valid values: 
     * `sg-example0123456789, sg-example1234567890`
     * `sg-example0123456789,sg-example1234567890`


## To do list:
* More docs
* Add Terraform examples for initial deployment
* Add support for any port (currently defaults 443)
* Code review? Tell me what sucks

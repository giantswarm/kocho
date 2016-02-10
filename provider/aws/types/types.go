// Package types provides some general types used by the api clients of the provider/aws package for internal usage.
package types

// Tag represents a key value pair.
type Tag struct {
	Key   string
	Value string
}

// StackResource represents a resource in a CloudFormation stack.
type StackResource struct {
	Type       string `json:"ResourceType"`
	Status     string `json:"ResourceStatus"`
	PhysicalId string `json:"PhysicalResourceId"`
	LogicalId  string `json:"LogicalResourceId"`
}

// Instance represents an instance on AWS.
type Instance struct {
	InstanceId       string
	ImageId          string
	InstanceType     string
	PublicIPAddress  string
	PublicDNSName    string
	PrivateIPAddress string
	PrivateDNSName   string
}

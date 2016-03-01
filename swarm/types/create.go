package swarmtypes

// CreateFlags describes flags for creating a swarm.
type CreateFlags struct {
	// Template type that should be used
	Type string

	// Comma separated list of key=value tag pairs that are applied to all nodes in the fleet cluster
	Tags string

	// The version tag of Yochu to be deployed
	YochuVersion string

	ClusterSize      int
	EtcdPeers        string
	EtcdVersion      string
	FleetVersion     string
	DockerVersion    string
	EtcdDiscoveryURL string
	K8sVersion       string
	RktVersion       string
	TemplateDir      string

	// MachineType provides some identifier for the provider to know which type of machine should be used.
	// for AWS these are the EC2 types, e.g. t2.nano, m3.large etc.
	MachineType string

	// URI for the OS image that should be used for the nodes in the cluster. Must be understood by the provider.
	ImageURI string

	// An URI for the certificate you want to deploy for the swarm. Must be understood by the provider.
	CertificateURI string

	// Provider Specific Structs
	*AWSCreateFlags
}

// AWSCreateFlags describes AWS specific flags for creating a swarm.
type AWSCreateFlags struct {
	KeypairName      string
	VPC              string
	VPCCIDR          string
	Subnet           string
	AvailabilityZone string
}

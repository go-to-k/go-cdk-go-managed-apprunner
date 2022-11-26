package input

type AppRunnerStackInputProps struct {
	StackEnv                         *StackEnv
	VpcConnectorProps                *VpcConnectorProps
	SourceConfigurationProps         *SourceConfigurationProps
	InstanceConfigurationProps       *InstanceConfigurationProps
	AutoScalingConfigurationArnProps *AutoScalingConfigurationArnProps
}

type StackEnv struct {
	Account string
	Region  string
}

type VpcConnectorProps struct {
	VpcID     string
	SubnetID1 string
	SubnetID2 string
}

type SourceConfigurationProps struct {
	RepositoryUrl  string
	BranchName     string
	BuildCommand   string
	StartCommand   string
	Port           int
	ConnectionName string
}

type InstanceConfigurationProps struct {
	Cpu    string
	Memory string
}

type AutoScalingConfigurationArnProps struct {
	MaxConcurrency int
	MaxSize        int
	MinSize        int
}

func NewAppRunnerStackInputProps() *AppRunnerStackInputProps {
	return &AppRunnerStackInputProps{
		StackEnv: &StackEnv{
			Account: "123456789012", // Your Account ID
			Region:  "ap-northeast-1",
		},
		VpcConnectorProps: &VpcConnectorProps{
			VpcID:     "vpc-xxxxxxxxxxxxxxx",    // Your VPC ID
			SubnetID1: "subnet-xxxxxxxxxxxxxxx", // Your Subnet ID
			SubnetID2: "subnet-xxxxxxxxxxxxxxx", // Your Subnet ID
		},
		SourceConfigurationProps: &SourceConfigurationProps{
			RepositoryUrl:  "https://github.com/go-to-k/go-cdk-go-managed-apprunner",
			BranchName:     "master",
			BuildCommand:   "go install ./app/...",
			StartCommand:   "go run app/main.go",
			Port:           8080,
			ConnectionName: "AppRunnerConnection",
		},
		InstanceConfigurationProps: &InstanceConfigurationProps{
			Cpu:    "1 vCPU",
			Memory: "2 GB",
		},
		AutoScalingConfigurationArnProps: &AutoScalingConfigurationArnProps{
			MaxConcurrency: 50,
			MaxSize:        3,
			MinSize:        1,
		},
	}
}

package input

type AppRunnerStackInputProps struct {
	SourceConfigurationProps         *SourceConfigurationProps
	InstanceConfigurationProps       *InstanceConfigurationProps
	AutoScalingConfigurationArnProps *AutoScalingConfigurationArnProps
	VpcConnectorProps                *VpcConnectorProps
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

type VpcConnectorProps struct {
	VpcID     string
	SubnetID1 string
	SubnetID2 string
}

func NewAppRunnerStackInputProps() *AppRunnerStackInputProps {
	return &AppRunnerStackInputProps{
		&SourceConfigurationProps{
			RepositoryUrl:  "https://github.com/go-to-k/go-cdk-go-managed-apprunner",
			BranchName:     "master",
			BuildCommand:   "go install ./app/...",
			StartCommand:   "go run app/main.go",
			Port:           8080,
			ConnectionName: "AppRunnerConnection",
		},
		&InstanceConfigurationProps{
			Cpu:    "1 vCPU",
			Memory: "2 GB",
		},
		&AutoScalingConfigurationArnProps{
			MaxConcurrency: 50,
			MaxSize:        3,
			MinSize:        1,
		},
		&VpcConnectorProps{
			VpcID:     "vpc-xxxxxxxxxxxxxxxxxx",
			SubnetID1: "subnet-xxxxxxxxxxxxxxxxxx",
			SubnetID2: "subnet-xxxxxxxxxxxxxxxxxx",
		},
	}
}

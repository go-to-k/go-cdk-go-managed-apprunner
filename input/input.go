package input

type AppRunnerStackInputProps struct {
	SourceConfigurationProps         *SourceConfigurationProps
	InstanceConfigurationProps       *InstanceConfigurationProps
	AutoScalingConfigurationArnProps *AutoScalingConfigurationArnProps
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
		&SourceConfigurationProps{
			"https://github.com/go-to-k/go-cdk-go-managed-apprunner",
			"master",
			"go install",
			"go run app/main.go",
			8080,
			"AppRunnerConnection",
		},
		&InstanceConfigurationProps{
			"1 vCPU",
			"2 GB",
		},
		&AutoScalingConfigurationArnProps{
			50,
			3,
			1,
		},
	}
}

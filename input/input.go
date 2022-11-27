package input

type AppRunnerStackInputs struct {
	SourceConfigurationInputs         *SourceConfigurationInputs
	InstanceConfigurationInputs       *InstanceConfigurationInputs
	AutoScalingConfigurationArnInputs *AutoScalingConfigurationArnInputs
}

type SourceConfigurationInputs struct {
	RepositoryUrl  string
	BranchName     string
	BuildCommand   string
	StartCommand   string
	Port           int
	ConnectionName string
}

type InstanceConfigurationInputs struct {
	Cpu    string
	Memory string
}

type AutoScalingConfigurationArnInputs struct {
	MaxConcurrency int
	MaxSize        int
	MinSize        int
}

func NewAppRunnerStackInputs() *AppRunnerStackInputs {
	return &AppRunnerStackInputs{
		&SourceConfigurationInputs{
			"https://github.com/go-to-k/go-cdk-go-managed-apprunner",
			"master",
			"go install",
			"go run app/main.go",
			8080,
			"AppRunnerConnection",
		},
		&InstanceConfigurationInputs{
			"1 vCPU",
			"2 GB",
		},
		&AutoScalingConfigurationArnInputs{
			50,
			3,
			1,
		},
	}
}

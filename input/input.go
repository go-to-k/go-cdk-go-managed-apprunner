package input

type AppRunnerStackInputs struct {
	RepositoryUrl  string
	BranchName     string
	BuildCommand   string
	StartCommand   string
	Cpu            string
	Memory         string
	Port           int
	MaxConcurrency int
	MaxSize        int
	MinSize        int
	ConnectionName string
}

func NewAppRunnerStackInputs() *AppRunnerStackInputs {
	return &AppRunnerStackInputs{
		"https://github.com/go-to-k/go-cdk-go-managed-apprunner",
		"master",
		"go install",
		"go run app/main.go",
		"1 vCPU",
		"2 GB",
		8080,
		50,
		3,
		1,
		"AppRunnerConnection",
	}
}

package input

type AppRunnerStackInputs struct {
	RepositoryUrl  string
	BranchName     string
	BuildCommand   string
	StartCommand   string
	ConnectionName string
}

func NewAppRunnerStackInputs() *AppRunnerStackInputs {
	return &AppRunnerStackInputs{
		"https://github.com/go-to-k/go-cdk-go-managed-apprunner",
		"master",
		"go install",
		"go run app/main.go",
		"AppRunnerConnection",
	}
}

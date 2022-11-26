package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapprunner"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AppRunnerStackProps struct {
	awscdk.StackProps
	repositoryUrl string
	branchName    string
	buildCommand  string
	startCommand  string
	connectionArn string
}

func NewAppRunnerStack(scope constructs.Construct, id string, props *AppRunnerStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	awsapprunner.NewCfnService(stack, jsii.String("AppRunnerService"), &awsapprunner.CfnServiceProps{
		SourceConfiguration: &awsapprunner.CfnService_SourceConfigurationProperty{
			AutoDeploymentsEnabled: jsii.Bool(true),
			AuthenticationConfiguration: &awsapprunner.CfnService_AuthenticationConfigurationProperty{
				ConnectionArn: jsii.String(props.connectionArn),
			},
			CodeRepository: &awsapprunner.CfnService_CodeRepositoryProperty{
				RepositoryUrl: jsii.String(props.repositoryUrl),
				SourceCodeVersion: &awsapprunner.CfnService_SourceCodeVersionProperty{
					Type:  jsii.String("BRANCH"),
					Value: jsii.String(props.branchName),
				},
				CodeConfiguration: &awsapprunner.CfnService_CodeConfigurationProperty{
					ConfigurationSource: jsii.String("API"),
					CodeConfigurationValues: &awsapprunner.CfnService_CodeConfigurationValuesProperty{
						Runtime:      jsii.String("GO_1"),
						BuildCommand: jsii.String(props.buildCommand),
						Port:         jsii.String("8080"),
						RuntimeEnvironmentVariables: []interface{}{
							&awsapprunner.CfnService_KeyValuePairProperty{
								Name:  jsii.String("ENV1"),
								Value: jsii.String("Test"),
							},
						},
						StartCommand: jsii.String(props.startCommand),
					},
				},
			},
		},
		HealthCheckConfiguration: &awsapprunner.CfnService_HealthCheckConfigurationProperty{
			Path:     jsii.String("/"),
			Protocol: jsii.String("HTTP"),
		},
		InstanceConfiguration: &awsapprunner.CfnService_InstanceConfigurationProperty{
			Cpu:    jsii.String("1 vCPU"),
			Memory: jsii.String("2 GB"),
		},
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewAppRunnerStack(app, "AppRunnerStack", &AppRunnerStackProps{
		awscdk.StackProps{
			Env: env(),
		},
		"https://github.com/go-to-k/go-cdk-go-managed-apprunner",
		"main",
		"go install",
		"go run app/main.go",
		os.Getenv("CONNECTION_ARN"),
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}

package main

import (
	"context"
	"fmt"
	"go-cdk-go-managed-apprunner/input"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapprunner"
	"github.com/aws/aws-cdk-go/awscdk/v2/customresources"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AppRunnerStackProps struct {
	awscdk.StackProps
	AppRunnerStackInputProps *input.AppRunnerStackInputProps
}

func NewAppRunnerStack(scope constructs.Construct, id string, props *AppRunnerStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// You must create a connection at the AWS AppRunner console before deploy.
	connectionArn, err := getConnectionArn(props.AppRunnerStackInputProps.SourceConfigurationProps.ConnectionName, *props.Env.Region)
	if err != nil {
		panic(err)
	}

	// TODO: Use API rather than custom resource ?(Because needs arn on delete but it is only generated on create)
	autoScalingConfigurationResult := customresources.NewAwsCustomResource(stack, jsii.String("AutoScalingConfiguration"), &customresources.AwsCustomResourceProps{
		Policy: customresources.AwsCustomResourcePolicy_FromSdkCalls(&customresources.SdkCallsPolicyOptions{
			Resources: customresources.AwsCustomResourcePolicy_ANY_RESOURCE(),
		}),
		OnCreate: &customresources.AwsSdkCall{
			Service: jsii.String("AppRunner"),
			Action:  jsii.String("createAutoScalingConfiguration"),
			Parameters: map[string]interface{}{
				"AutoScalingConfigurationName": jsii.String(*stack.StackName()),
				"MaxConcurrency":               jsii.String(strconv.Itoa(props.AppRunnerStackInputProps.AutoScalingConfigurationArnProps.MaxConcurrency)),
				"MaxSize":                      jsii.String(strconv.Itoa(props.AppRunnerStackInputProps.AutoScalingConfigurationArnProps.MaxSize)),
				"MinSize":                      jsii.String(strconv.Itoa(props.AppRunnerStackInputProps.AutoScalingConfigurationArnProps.MinSize)),
			},
			PhysicalResourceId: customresources.PhysicalResourceId_Of(jsii.String("AutoScalingConfiguration")),
		},
		// OnDelete: &customresources.AwsSdkCall{
		// 	Service: jsii.String("AppRunner"),
		// 	Action:  jsii.String("deleteAutoScalingConfiguration"),
		// 	Parameters: map[string]interface{}{
		// 		"AutoScalingConfigurationArn": jsii.String(""),
		// 	},
		// },
	})
	autoScalingConfigurationArn := autoScalingConfigurationResult.GetResponseField(jsii.String("AutoScalingConfiguration.AutoScalingConfigurationArn"))

	// There is an L2 construct if it is an alpha version.
	awsapprunner.NewCfnService(stack, jsii.String("AppRunnerService"), &awsapprunner.CfnServiceProps{
		SourceConfiguration: &awsapprunner.CfnService_SourceConfigurationProperty{
			AutoDeploymentsEnabled: jsii.Bool(true),
			AuthenticationConfiguration: &awsapprunner.CfnService_AuthenticationConfigurationProperty{
				ConnectionArn: jsii.String(connectionArn),
			},
			CodeRepository: &awsapprunner.CfnService_CodeRepositoryProperty{
				RepositoryUrl: jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.RepositoryUrl),
				SourceCodeVersion: &awsapprunner.CfnService_SourceCodeVersionProperty{
					Type:  jsii.String("BRANCH"),
					Value: jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.BranchName),
				},
				CodeConfiguration: &awsapprunner.CfnService_CodeConfigurationProperty{
					ConfigurationSource: jsii.String("API"),
					CodeConfigurationValues: &awsapprunner.CfnService_CodeConfigurationValuesProperty{
						Runtime:      jsii.String("GO_1"),
						BuildCommand: jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.BuildCommand),
						Port:         jsii.String(strconv.Itoa(props.AppRunnerStackInputProps.SourceConfigurationProps.Port)),
						RuntimeEnvironmentVariables: []interface{}{
							&awsapprunner.CfnService_KeyValuePairProperty{
								Name:  jsii.String("ENV1"),
								Value: jsii.String("Test"),
							},
						},
						StartCommand: jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.StartCommand),
					},
				},
			},
		},
		HealthCheckConfiguration: &awsapprunner.CfnService_HealthCheckConfigurationProperty{
			Path:     jsii.String("/"),
			Protocol: jsii.String("HTTP"),
		},
		InstanceConfiguration: &awsapprunner.CfnService_InstanceConfigurationProperty{
			Cpu:    jsii.String(props.AppRunnerStackInputProps.InstanceConfigurationProps.Cpu),
			Memory: jsii.String(props.AppRunnerStackInputProps.InstanceConfigurationProps.Memory),
		},
		AutoScalingConfigurationArn: autoScalingConfigurationArn,
	})

	return stack
}

func getConnectionArn(connectionName string, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", err
	}

	client := apprunner.NewFromConfig(cfg)

	input := &apprunner.ListConnectionsInput{
		ConnectionName: aws.String(connectionName),
	}

	output, err := client.ListConnections(context.TODO(), input)
	if err != nil {
		return "", err
	}

	if len(output.ConnectionSummaryList) > 0 {
		connectionArn := aws.ToString(output.ConnectionSummaryList[0].ConnectionArn)
		return connectionArn, nil
	}

	return "", fmt.Errorf("connection not found.")
}

func main() {
	app := awscdk.NewApp(nil)

	appRunnerStackInputProps := input.NewAppRunnerStackInputProps()

	appRunnerStackProps := &AppRunnerStackProps{
		awscdk.StackProps{
			Env: env(),
		},
		appRunnerStackInputProps,
	}

	NewAppRunnerStack(app, "AppRunnerStack", appRunnerStackProps)

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Region: jsii.String("ap-northeast-1"),
	}
}

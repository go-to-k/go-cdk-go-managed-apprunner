package main

import (
	"go-cdk-go-managed-apprunner/cdk/input"
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	assertions "github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/bradleyjkemp/cupaloy/v2"
)

func TestAppRunnerStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	appRunnerStackInputProps := input.NewAppRunnerStackInputProps()

	appRunnerStackProps := &AppRunnerStackProps{
		awscdk.StackProps{
			Env: env(
				appRunnerStackInputProps.StackEnv.Account,
				appRunnerStackInputProps.StackEnv.Region,
			),
		},
		appRunnerStackInputProps,
	}

	// WHEN
	stack := NewAppRunnerStack(app, "AppRunnerStack", appRunnerStackProps)

	// THEN
	template := assertions.Template_FromStack(stack, nil)
	templateJson := convertSnapshot(template.ToJSON())

	t.Run("Snapshot Test", func(t *testing.T) {
		cupaloy.SnapshotT(t, templateJson)
	})

	t.Run("CustomResourceLambda created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::Lambda::Function"), jsii.Number(1))
	})

	t.Run("AutoScalingConfiguration created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("Custom::AutoScalingConfiguration"), jsii.Number(1))
	})

	t.Run("IAMRole created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::IAM::Role"), jsii.Number(2))
	})

	t.Run("IAMPolicy created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::IAM::Policy"), jsii.Number(1))
	})

	t.Run("SecurityGroup created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::EC2::SecurityGroup"), jsii.Number(2))
	})

	t.Run("VpcConnector created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::AppRunner::VpcConnector"), jsii.Number(2))
	})

	t.Run("AppRunner Service created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::AppRunner::Service"), jsii.Number(2))
	})

}

func convertSnapshot(templateJson *map[string]interface{}) map[string]interface{} {
	resources := (*templateJson)["Resources"].(map[string]interface{})
	for key := range resources {
		if valProperties, ok := resources[key].(map[string]interface{})["Properties"]; ok {
			properties := valProperties.(map[string]interface{})
			if valCode, ok := properties["Code"]; ok {
				code := valCode.(map[string]interface{})
				if _, ok := code["S3Key"]; ok {
					(*templateJson)["Resources"].(map[string]interface{})[key].(map[string]interface{})["Properties"].(map[string]interface{})["Code"].(map[string]interface{})["S3Key"] = ""
				}
			}
		}
	}
	return resources
}

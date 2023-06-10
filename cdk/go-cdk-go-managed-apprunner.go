package main

import (
	"bufio"
	"context"
	"fmt"
	"go-cdk-go-managed-apprunner/cdk/input"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapprunner"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	apprunner "github.com/aws/aws-cdk-go/awscdkapprunneralpha/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	apprunnerClient "github.com/aws/aws-sdk-go-v2/service/apprunner"
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

	/*
		Custom Resource Lambda for creation of AutoScalingConfiguration
	*/
	customResourceLambda := awslambda.NewFunction(stack, jsii.String("CustomResourceLambda"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_GO_1_X(),
		Handler: jsii.String("main"),
		Code: awslambda.AssetCode_FromAsset(jsii.String("../"), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image:   awslambda.Runtime_GO_1_X().BundlingImage(),
				Command: jsii.Strings("bash", "-c", "GOOS=linux GOARCH=amd64 go build -o /asset-output/main custom/custom.go"),
				User:    jsii.String("root"),
			},
		}),
		Timeout: awscdk.Duration_Seconds(jsii.Number(900)),
		InitialPolicy: &[]awsiam.PolicyStatement{
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Actions: &[]*string{
					jsii.String("apprunner:*AutoScalingConfiguration*"),
					jsii.String("apprunner:UpdateService"),
					jsii.String("apprunner:ListOperations"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Actions: &[]*string{
					jsii.String("cloudformation:DescribeStacks"),
				},
				Resources: &[]*string{
					stack.StackId(),
				},
			}),
		},
	})

	/*
		AutoScalingConfiguration
	*/
	autoScalingConfiguration := awscdk.NewCustomResource(stack, jsii.String("AutoScalingConfiguration"), &awscdk.CustomResourceProps{
		ResourceType: jsii.String("Custom::AutoScalingConfiguration"),
		Properties: &map[string]interface{}{
			"AutoScalingConfigurationName": *stack.StackName(),
			"MaxConcurrency":               strconv.Itoa(props.AppRunnerStackInputProps.AutoScalingConfigurationArnProps.MaxConcurrency),
			"MaxSize":                      strconv.Itoa(props.AppRunnerStackInputProps.AutoScalingConfigurationArnProps.MaxSize),
			"MinSize":                      strconv.Itoa(props.AppRunnerStackInputProps.AutoScalingConfigurationArnProps.MinSize),
			"StackName":                    *stack.StackName(),
		},
		ServiceToken: customResourceLambda.FunctionArn(),
	})
	autoScalingConfigurationArn := autoScalingConfiguration.GetAttString(jsii.String("AutoScalingConfigurationArn"))

	/*
		ConnectionArn for GitHub Connection
	*/
	connectionArn, err := createConnection(props.AppRunnerStackInputProps.SourceConfigurationProps.ConnectionName, *props.Env.Region)
	if err != nil {
		panic(err)
	}

	/*
		InstanceRole for AppRunner Service
	*/
	appRunnerInstanceRole := awsiam.NewRole(stack, jsii.String("AppRunnerInstanceRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("tasks.apprunner.amazonaws.com"), nil),
	})

	/*
		L2 Construct(alpha version) for VPC Connector
	*/
	securityGroupForVpcConnectorL2 := awsec2.NewSecurityGroup(stack, jsii.String("SecurityGroupForVpcConnectorL2"), &awsec2.SecurityGroupProps{
		Vpc: awsec2.Vpc_FromLookup(stack, jsii.String("VPCForSecurityGroupForVpcConnectorL2"), &awsec2.VpcLookupOptions{
			VpcId: jsii.String(props.AppRunnerStackInputProps.VpcConnectorProps.VpcID),
		}),
		Description: jsii.String("for AppRunner VPC Connector L2"),
	})

	vpcConnectorL2 := apprunner.NewVpcConnector(stack, jsii.String("VpcConnectorL2"), &apprunner.VpcConnectorProps{
		Vpc: awsec2.Vpc_FromLookup(stack, jsii.String("VPCForVpcConnectorL2"), &awsec2.VpcLookupOptions{
			VpcId: jsii.String(props.AppRunnerStackInputProps.VpcConnectorProps.VpcID),
		}),
		SecurityGroups: &[]awsec2.ISecurityGroup{securityGroupForVpcConnectorL2},
		VpcSubnets: &awsec2.SubnetSelection{
			Subnets: &[]awsec2.ISubnet{
				awsec2.Subnet_FromSubnetId(stack, jsii.String("Subnet1"), jsii.String(props.AppRunnerStackInputProps.VpcConnectorProps.SubnetID1)),
				awsec2.Subnet_FromSubnetId(stack, jsii.String("Subnet2"), jsii.String(props.AppRunnerStackInputProps.VpcConnectorProps.SubnetID2)),
			},
		},
	})

	/*
		L1 Construct for VPC Connector
	*/
	securityGroupForVpcConnectorL1 := awsec2.NewSecurityGroup(stack, jsii.String("SecurityGroupForVpcConnectorL1"), &awsec2.SecurityGroupProps{
		Vpc: awsec2.Vpc_FromLookup(stack, jsii.String("VPCForVpcConnectorL1"), &awsec2.VpcLookupOptions{
			VpcId: jsii.String(props.AppRunnerStackInputProps.VpcConnectorProps.VpcID),
		}),
		Description: jsii.String("for AppRunner VPC Connector L1"),
	})

	vpcConnectorL1 := awsapprunner.NewCfnVpcConnector(stack, jsii.String("VpcConnectorL1"), &awsapprunner.CfnVpcConnectorProps{
		SecurityGroups: jsii.Strings(*securityGroupForVpcConnectorL1.SecurityGroupId()),
		Subnets: jsii.Strings(
			props.AppRunnerStackInputProps.VpcConnectorProps.SubnetID1,
			props.AppRunnerStackInputProps.VpcConnectorProps.SubnetID2,
		),
	})

	/*
		L2 Construct(alpha version) for AppRunner Service
	*/
	apprunnerServiceL2 := apprunner.NewService(stack, jsii.String("AppRunnerServiceL2"), &apprunner.ServiceProps{
		InstanceRole: appRunnerInstanceRole,
		Source: apprunner.Source_FromGitHub(&apprunner.GithubRepositoryProps{
			RepositoryUrl:       jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.RepositoryUrl),
			Branch:              jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.BranchName),
			ConfigurationSource: apprunner.ConfigurationSourceType_API,
			CodeConfigurationValues: &apprunner.CodeConfigurationValues{
				Runtime:      apprunner.Runtime_GO_1(),
				Port:         jsii.String(strconv.Itoa(props.AppRunnerStackInputProps.SourceConfigurationProps.Port)),
				StartCommand: jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.StartCommand),
				BuildCommand: jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.BuildCommand),
				Environment: &map[string]*string{
					"ENV1": jsii.String("L2"),
				},
			},
			Connection: apprunner.GitHubConnection_FromConnectionArn(jsii.String(connectionArn)),
		}),
		Cpu:                    apprunner.Cpu_Of(jsii.String(props.AppRunnerStackInputProps.InstanceConfigurationProps.Cpu)),
		Memory:                 apprunner.Memory_Of(jsii.String(props.AppRunnerStackInputProps.InstanceConfigurationProps.Memory)),
		VpcConnector:           vpcConnectorL2,
		AutoDeploymentsEnabled: jsii.Bool(true),
	})

	var cfnAppRunner awsapprunner.CfnService
	//lint:ignore SA1019 This is deprecated, but Go does not support escape hatches yet.
	jsii.Get(apprunnerServiceL2.Node(), "defaultChild", &cfnAppRunner)
	cfnAppRunner.SetAutoScalingConfigurationArn(autoScalingConfigurationArn)
	cfnAppRunner.SetHealthCheckConfiguration(&awsapprunner.CfnService_HealthCheckConfigurationProperty{
		Path:     jsii.String("/"),
		Protocol: jsii.String("HTTP"),
	})

	/*
		L1 Construct for AppRunner Service
	*/
	apprunnerServiceL1 := awsapprunner.NewCfnService(stack, jsii.String("AppRunnerServiceL1"), &awsapprunner.CfnServiceProps{
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
						Port:         jsii.String(strconv.Itoa(props.AppRunnerStackInputProps.SourceConfigurationProps.Port)),
						StartCommand: jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.StartCommand),
						BuildCommand: jsii.String(props.AppRunnerStackInputProps.SourceConfigurationProps.BuildCommand),
						RuntimeEnvironmentVariables: []interface{}{
							&awsapprunner.CfnService_KeyValuePairProperty{
								Name:  jsii.String("ENV1"),
								Value: jsii.String("L1"),
							},
						},
					},
				},
			},
		},
		HealthCheckConfiguration: &awsapprunner.CfnService_HealthCheckConfigurationProperty{
			Path:     jsii.String("/"),
			Protocol: jsii.String("HTTP"),
		},
		InstanceConfiguration: &awsapprunner.CfnService_InstanceConfigurationProperty{
			Cpu:             jsii.String(props.AppRunnerStackInputProps.InstanceConfigurationProps.Cpu),
			Memory:          jsii.String(props.AppRunnerStackInputProps.InstanceConfigurationProps.Memory),
			InstanceRoleArn: appRunnerInstanceRole.RoleArn(),
		},
		NetworkConfiguration: &awsapprunner.CfnService_NetworkConfigurationProperty{
			EgressConfiguration: awsapprunner.CfnService_EgressConfigurationProperty{
				EgressType:      jsii.String("VPC"),
				VpcConnectorArn: vpcConnectorL1.AttrVpcConnectorArn(),
			},
		},
		AutoScalingConfigurationArn: autoScalingConfigurationArn,
	})

	awscdk.NewCfnOutput(stack, jsii.String("AppRunnerServiceL2ServiceArn"), &awscdk.CfnOutputProps{
		Value:      apprunnerServiceL2.ServiceArn(),
		ExportName: jsii.String(*stack.StackName() + "AppRunnerServiceL2ServiceArn"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("AppRunnerServiceL1ServiceArn"), &awscdk.CfnOutputProps{
		Value:      apprunnerServiceL1.AttrServiceArn(),
		ExportName: jsii.String(*stack.StackName() + "AppRunnerServiceL1ServiceArn"),
	})

	return stack
}

func createConnection(connectionName string, region string) (string, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", err
	}

	client := apprunnerClient.NewFromConfig(cfg)

	listConnectionsInput := &apprunnerClient.ListConnectionsInput{
		ConnectionName: aws.String(connectionName),
	}

	listConnectionsOutput, err := client.ListConnections(ctx, listConnectionsInput)
	if err != nil {
		return "", err
	}

	// If there is already a connection, return the connection ARN
	if len(listConnectionsOutput.ConnectionSummaryList) > 0 {
		if listConnectionsOutput.ConnectionSummaryList[0].Status == "PENDING_HANDSHAKE" {
			confirmCompleteHandshake()
		}
		connectionArn := aws.ToString(listConnectionsOutput.ConnectionSummaryList[0].ConnectionArn)
		return connectionArn, nil
	}

	// Otherwise, create a connection
	createConnectionInput := &apprunnerClient.CreateConnectionInput{
		ConnectionName: aws.String(connectionName),
		ProviderType:   "GITHUB",
	}

	createConnectionOutput, err := client.CreateConnection(ctx, createConnectionInput)
	if err != nil {
		return "", err
	}

	confirmCompleteHandshake()
	return *createConnectionOutput.Connection.ConnectionArn, nil
}

func confirmCompleteHandshake() {
	for {
		fmt.Println("Now, click the \"Complete handshake\" button at the AWS App Runner console.")
		if ok := getYesNo("Did you click the button?"); ok {
			return
		}
		continue
	}
}

func getYesNo(label string) bool {
	choices := "Y/n"
	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		fmt.Fprintln(os.Stderr)

		s = strings.TrimSpace(s)
		if s == "" {
			return true
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}

func main() {
	defer jsii.Close()

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

	NewAppRunnerStack(app, "AppRunnerGoStack", appRunnerStackProps)

	app.Synth(nil)
}

func env(account string, region string) *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"golang.org/x/sync/errgroup"
)

type InputProps struct {
	autoScalingConfigurationName string
	maxConcurrency               int
	maxSize                      int
	minSize                      int
	stackName                    string
}

func HandleRequest(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	physicalResourceID = "AutoScalingConfiguration"
	requestType := event.RequestType
	data = make(map[string]interface{})

	inputProps, err := convertInputParameters(event.ResourceProperties)
	if err != nil {
		return "", nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return "", nil, err
	}

	apprunnerClient := apprunner.NewFromConfig(cfg)
	cfnClient := cloudformation.NewFromConfig(cfg)

	if requestType == "Create" {
		autoScalingConfigurationArn, err := createAutoScalingConfiguration(ctx, apprunnerClient, inputProps)
		if err != nil {
			return "", nil, err
		}

		data["AutoScalingConfigurationArn"] = autoScalingConfigurationArn
	} else if requestType == "Update" {
		autoScalingConfigurationList, err := listAutoScalingConfiguration(ctx, apprunnerClient, inputProps.autoScalingConfigurationName)
		if err != nil {
			return "", nil, err
		}

		if len(autoScalingConfigurationList) > 0 {
			err := changeAutoScalingConfigurationToDefault(ctx, apprunnerClient, cfnClient, inputProps.stackName)
			if err != nil {
				return "", nil, err
			}

			err = deleteAutoScalingConfiguration(ctx, apprunnerClient, *autoScalingConfigurationList[0].AutoScalingConfigurationArn)
			if err != nil {
				return "", nil, err
			}
		}

		autoScalingConfigurationArn, err := createAutoScalingConfiguration(ctx, apprunnerClient, inputProps)
		if err != nil {
			return "", nil, err
		}

		data["AutoScalingConfigurationArn"] = autoScalingConfigurationArn
	} else if requestType == "Delete" {
		autoScalingConfigurationList, err := listAutoScalingConfiguration(ctx, apprunnerClient, inputProps.autoScalingConfigurationName)
		if err != nil {
			return "", nil, err
		}

		if len(autoScalingConfigurationList) > 0 {
			err := deleteAutoScalingConfiguration(ctx, apprunnerClient, *autoScalingConfigurationList[0].AutoScalingConfigurationArn)
			if err != nil {
				return "", nil, err
			}
		}
	}

	return
}

func listAutoScalingConfiguration(ctx context.Context, client *apprunner.Client, autoScalingConfigurationName string) ([]types.AutoScalingConfigurationSummary, error) {
	output, err := client.ListAutoScalingConfigurations(ctx, &apprunner.ListAutoScalingConfigurationsInput{
		AutoScalingConfigurationName: &autoScalingConfigurationName,
	})
	if err != nil {
		return nil, err
	}

	return output.AutoScalingConfigurationSummaryList, nil
}

func createAutoScalingConfiguration(ctx context.Context, client *apprunner.Client, inputProps *InputProps) (string, error) {
	output, err := client.CreateAutoScalingConfiguration(ctx, &apprunner.CreateAutoScalingConfigurationInput{
		AutoScalingConfigurationName: aws.String(inputProps.autoScalingConfigurationName),
		MaxConcurrency:               aws.Int32(int32(inputProps.maxConcurrency)),
		MaxSize:                      aws.Int32(int32(inputProps.maxSize)),
		MinSize:                      aws.Int32(int32(inputProps.minSize)),
	})
	if err != nil {
		return "", err
	}

	return *output.AutoScalingConfiguration.AutoScalingConfigurationArn, nil
}

func deleteAutoScalingConfiguration(ctx context.Context, client *apprunner.Client, autoScalingConfigurationArn string) error {
	_, err := client.DeleteAutoScalingConfiguration(ctx, &apprunner.DeleteAutoScalingConfigurationInput{
		AutoScalingConfigurationArn: aws.String(autoScalingConfigurationArn),
	})

	return err
}

func getServiceArns(ctx context.Context, client *cloudformation.Client, stackName string) ([]string, error) {
	stacks, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, err
	}

	arns := []string{}
	for _, output := range stacks.Stacks[0].Outputs {
		if *output.ExportName == stackName+"AppRunnerServiceL1ServiceArn" || *output.ExportName == stackName+"AppRunnerServiceL2ServiceArn" {
			arns = append(arns, *output.OutputValue)
		}
	}

	return arns, nil
}

func waitOperation(ctx context.Context, apprunnerClient *apprunner.Client, operationId string, serviceArn string) error {
	if operationId == "" {
		return fmt.Errorf("OperationId is empty")
	}

	for {
		output, err := apprunnerClient.ListOperations(ctx, &apprunner.ListOperationsInput{
			ServiceArn: aws.String(serviceArn),
		})
		if err != nil {
			return err
		}
		if len(output.OperationSummaryList) == 0 {
			return fmt.Errorf("OperationSummaryList is empty")
		}

		for _, operationSummary := range output.OperationSummaryList {
			if *operationSummary.Id == operationId {
				if operationSummary.Status == types.OperationStatusSucceeded {
					return nil
				} else if operationSummary.Status == types.OperationStatusInProgress || operationSummary.Status == types.OperationStatusPending {
					time.Sleep(time.Second * 10)
				} else {
					return fmt.Errorf("OperationError status:" + string(operationSummary.Status))
				}
			}
		}
	}
}

func updateServiceForAutoScalingConfiguration(
	ctx context.Context,
	apprunnerClient *apprunner.Client,
	cfnClient *cloudformation.Client,
	stackName string,
	autoScalingConfigurationArn string,
) error {
	serviceArns, err := getServiceArns(ctx, cfnClient, stackName)
	if err != nil {
		return err
	}
	if len(serviceArns) == 0 {
		return fmt.Errorf("Service Arns not found")
	}

	eg, ctx := errgroup.WithContext(ctx)
	for _, serviceArn := range serviceArns {
		serviceArn := serviceArn
		eg.Go(func() error {
			output, err := apprunnerClient.UpdateService(ctx, &apprunner.UpdateServiceInput{
				ServiceArn:                  aws.String(serviceArn),
				AutoScalingConfigurationArn: aws.String(autoScalingConfigurationArn),
			})
			if err != nil {
				return err
			}

			if err = waitOperation(ctx, apprunnerClient, *output.OperationId, serviceArn); err != nil {
				return err
			}

			return nil
		})
	}

	return eg.Wait()
}

func changeAutoScalingConfigurationToDefault(ctx context.Context, apprunnerClient *apprunner.Client, cfnClient *cloudformation.Client, stackName string) error {
	defaultAutoScalingConfigurationName := "DefaultConfiguration"
	defaultAutoScalingConfiguration, err := listAutoScalingConfiguration(ctx, apprunnerClient, defaultAutoScalingConfigurationName)
	if err != nil {
		return err
	}

	if len(defaultAutoScalingConfiguration) > 0 {
		autoScalingConfigurationArn := *defaultAutoScalingConfiguration[0].AutoScalingConfigurationArn
		if err = updateServiceForAutoScalingConfiguration(ctx, apprunnerClient, cfnClient, stackName, autoScalingConfigurationArn); err != nil {
			return err
		}
	}

	return nil
}

func convertInputParameters(resourceProperties map[string]interface{}) (*InputProps, error) {
	autoScalingConfigurationName, ok := resourceProperties["AutoScalingConfigurationName"].(string)
	if !ok {
		return nil, fmt.Errorf("AutoScalingConfigurationName Assertion Error: %v", resourceProperties["AutoScaling"])
	}

	maxConcurrencyInput, ok := resourceProperties["MaxConcurrency"].(string)
	if !ok {
		return nil, fmt.Errorf("MaxConcurrency Assertion Error: %v", resourceProperties["MaxConcurrency"])
	}
	maxConcurrency, err := strconv.Atoi(maxConcurrencyInput)
	if err != nil {
		return nil, fmt.Errorf("MaxConcurrency Convert Error: %v", maxConcurrency)
	}

	maxSizeInput, ok := resourceProperties["MaxSize"].(string)
	if !ok {
		return nil, fmt.Errorf("MaxSize Assertion Error: %v", resourceProperties["MaxSize"])
	}
	maxSize, err := strconv.Atoi(maxSizeInput)
	if err != nil {
		return nil, fmt.Errorf("MaxSize Convert Error: %v", maxSize)
	}

	minSizeInput, ok := resourceProperties["MinSize"].(string)
	if !ok {
		return nil, fmt.Errorf("MinSize Assertion Error: %v", resourceProperties["MinSize"])
	}
	minSize, err := strconv.Atoi(minSizeInput)
	if err != nil {
		return nil, fmt.Errorf("MinSize Convert Error: %v", minSize)
	}

	stackName, ok := resourceProperties["StackName"].(string)
	if !ok {
		return nil, fmt.Errorf("StackName Assertion Error: %v", resourceProperties["StackName"])
	}

	return &InputProps{
		autoScalingConfigurationName: autoScalingConfigurationName,
		maxConcurrency:               maxConcurrency,
		maxSize:                      maxSize,
		minSize:                      minSize,
		stackName:                    stackName,
	}, nil
}

func main() {
	lambda.Start(cfn.LambdaWrap(HandleRequest))
}

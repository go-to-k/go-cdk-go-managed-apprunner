package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
)

type InputProps struct {
	autoScalingConfigurationName string
	maxConcurrency               int
	maxSize                      int
	minSize                      int
}

func HandleRequest(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	requestType := event.RequestType
	data = make(map[string]interface{})

	inputProps, err := convertInputParameters(event.ResourceProperties)
	if err != nil {
		return "", nil, err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return "", nil, err
	}

	client := apprunner.NewFromConfig(cfg)

	if requestType == "Create" {
		createInput := &apprunner.CreateAutoScalingConfigurationInput{
			AutoScalingConfigurationName: aws.String(inputProps.autoScalingConfigurationName),
			MaxConcurrency:               aws.Int32(int32(inputProps.maxConcurrency)),
			MaxSize:                      aws.Int32(int32(inputProps.maxSize)),
			MinSize:                      aws.Int32(int32(inputProps.minSize)),
		}

		output, err := client.CreateAutoScalingConfiguration(context.TODO(), createInput)
		if err != nil {
			return "", nil, err
		}

		data["AutoScalingConfigurationArn"] = output.AutoScalingConfiguration.AutoScalingConfigurationArn
	} else if requestType == "Delete" {
		listInput := &apprunner.ListAutoScalingConfigurationsInput{
			AutoScalingConfigurationName: aws.String(inputProps.autoScalingConfigurationName),
		}

		output, err := client.ListAutoScalingConfigurations(context.TODO(), listInput)
		if err != nil {
			return "", nil, err
		}

		if len(output.AutoScalingConfigurationSummaryList) > 0 {
			autoScalingConfigurationArn := output.AutoScalingConfigurationSummaryList[0].AutoScalingConfigurationArn

			deleteInput := &apprunner.DeleteAutoScalingConfigurationInput{
				AutoScalingConfigurationArn: autoScalingConfigurationArn,
			}

			_, err := client.DeleteAutoScalingConfiguration(context.TODO(), deleteInput)
			if err != nil {
				return "", nil, err
			}
		}
	}

	physicalResourceID = "AutoScalingConfiguration"

	return
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

	return &InputProps{
		autoScalingConfigurationName: autoScalingConfigurationName,
		maxConcurrency:               maxConcurrency,
		maxSize:                      maxSize,
		minSize:                      minSize,
	}, nil
}

func main() {
	lambda.Start(cfn.LambdaWrap(HandleRequest))
}

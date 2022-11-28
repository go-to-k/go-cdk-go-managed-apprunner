package main

import (
	"context"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
)

func HandleRequest(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	requestType := event.RequestType

	autoScalingConfigurationName, _ := event.ResourceProperties["AutoScalingConfigurationName"].(string)
	maxConcurrency, _ := strconv.Atoi(event.ResourceProperties["MaxConcurrency"].(string))
	maxSize, _ := strconv.Atoi(event.ResourceProperties["MaxSize"].(string))
	minSize, _ := strconv.Atoi(event.ResourceProperties["MinSize"].(string))

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return "", nil, err
	}
	client := apprunner.NewFromConfig(cfg)

	if requestType == "Create" {
		createInput := &apprunner.CreateAutoScalingConfigurationInput{
			AutoScalingConfigurationName: aws.String(autoScalingConfigurationName),
			MaxConcurrency:               aws.Int32(int32(maxConcurrency)),
			MaxSize:                      aws.Int32(int32(maxSize)),
			MinSize:                      aws.Int32(int32(minSize)),
		}
		output, err := client.CreateAutoScalingConfiguration(context.TODO(), createInput)
		if err != nil {
			return "", nil, err
		}

		data = make(map[string]interface{})
		data["AutoScalingConfigurationArn"] = output.AutoScalingConfiguration.AutoScalingConfigurationArn
	} else if requestType == "Delete" {
		listInput := &apprunner.ListAutoScalingConfigurationsInput{
			AutoScalingConfigurationName: aws.String(autoScalingConfigurationName),
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

func main() {
	lambda.Start(cfn.LambdaWrap(HandleRequest))
}

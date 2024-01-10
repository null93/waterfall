package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	StackNotFoundErr = errors.New("stack not found")
)

func GetConfig(profile string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}

func CheckAuth(cfg aws.Config) error {
	stsClient := sts.NewFromConfig(cfg)
	_, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	return err
}

func GetStackArnFromStackName(cfg aws.Config, stackName string) (string, error) {
	cfnClient := cloudformation.NewFromConfig(cfg)
	params := cloudformation.ListStacksInput{}
	paginator := cloudformation.NewListStacksPaginator(cfnClient, &params)
	for paginator.HasMorePages() {
		response, err := paginator.NextPage(context.Background())
		if err != nil {
			return "", err
		}
		for _, stack := range response.StackSummaries {
			if aws.ToString(stack.StackName) == stackName {
				return aws.ToString(stack.StackId), nil
			}
		}
	}
	return "", StackNotFoundErr
}

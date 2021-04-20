package eks

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awseks"
	"github.com/aws/jsii-runtime-go"
	stackHelper "github.com/faryne/go-cdk-example/libs/stack_helper"
)

type Result struct {
	Stack awscdk.Stack
}

var stack *stackHelper.MyCDKStack

func Init(parentStack awscdk.Stack, stackName *string, props *awscdk.StackProps) Result {
	stack := awscdk.NewStack(parentStack, stackName, props)

	// 建立 EKS cluster
	stack.AddResource([]awscdk.CfnResource{}, createEksCluster)
	// 建立 EKS Fargate Profile
	stack.AddResource([]awscdk.CfnResource{}, createEksFargateProfile)

	return stack
}
func createEksCluster(stack *stackHelper.MyCDKStack, dependencies []awscdk.CfnResource) awscdk.CfnResource {
	resource := awseks.NewCfnCluster(stack.Stack, jsii.String("aaaa"), &awseks.CfnClusterProps{
		ResourcesVpcConfig:      nil,
		RoleArn:                 nil,
		EncryptionConfig:        nil,
		KubernetesNetworkConfig: nil,
		Name:                    nil,
		Version:                 nil,
	})

	return resource
}
func createEksFargateProfile(stack *stackHelper.MyCDKStack, dependencies []awscdk.CfnResource) awscdk.CfnResource {
	resource := awseks.NewCfnFargateProfile(stack.Stack, jsii.String("xbbb"), &awseks.CfnFargateProfileProps{
		ClusterName:         stack.GetOutputResource(jsii.String("aaaa")).Ref(),
		PodExecutionRoleArn: nil,
		Selectors:           nil,
		FargateProfileName:  nil,
		Subnets:             nil,
		Tags:                nil,
	})

	return resource
}

package vpc

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/jsii-runtime-go"
	stackHelper "github.com/faryne/go-cdk-example/libs/stack_helper"
)

func Init(parentStack awscdk.CfnStack, stackName *string, props *awscdk.CfnStackProps) *stackHelper.MyCDKStack {
	stack := stackHelper.New(parentStack, stackName, props)

	// 建立 VPC
	stack.AddResource([]awscdk.CfnResource{}, newVpc)

	return stack
}
func newVpc(stack *stackHelper.MyCDKStack, dependencies []awscdk.CfnResource) awscdk.CfnResource {
	resource := awsec2.NewCfnVPC(stack.Stack, jsii.String("newVpcPreview"), &awsec2.CfnVPCProps{
		CidrBlock:          jsii.String("10.101.0.0/16"),
		EnableDnsHostnames: true,
		EnableDnsSupport:   true,
	})
	return resource
}

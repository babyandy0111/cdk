package vpc

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/jsii-runtime-go"
	stackHelper "github.com/faryne/go-cdk-example/libs/stack_helper"
)

var (
	newVpcSourceId = "newVpcSourceId"
)

func Init(parentStack awscdk.Stack, stackName *string, props *awscdk.StackProps) *stackHelper.MyCDKStack {
	stack := stackHelper.New(parentStack, stackName, props)

	// 建立 VPC
	stack.AddResource([]awscdk.Resource{}, newVpc)

	return stack
}

func newVpc(stack *stackHelper.MyCDKStack, dependencies []awscdk.Resource) awscdk.Resource {
	// subnet 設定，這會自動根據該 region 的 AZ 各自產生出 Subnet
	// 以 ap-northeast-1 而言，這會在 ap-northeast-1a / ap-northeast-1c / ap-northeast-1d 各自產生一組 public /private subnet
	var subnets = []*awsec2.SubnetConfiguration{
		&awsec2.SubnetConfiguration{
			Name:       jsii.String("PublicSubnet"),
			SubnetType: awsec2.SubnetType_PUBLIC,
			CidrMask:   jsii.Number(float64(19)),
		},
		&awsec2.SubnetConfiguration{
			Name:       jsii.String("PrivateSubnet"),
			SubnetType: awsec2.SubnetType_PRIVATE,
			CidrMask:   jsii.Number(float64(20)),
		},
	}
	resource := awsec2.NewVpc(stack.Stack, jsii.String("newVpcPreview"), &awsec2.VpcProps{
		Cidr:                jsii.String("10.101.0.0/16"),
		EnableDnsHostnames:  jsii.Bool(true),
		EnableDnsSupport:    jsii.Bool(true),
		SubnetConfiguration: &subnets,
	})

	return resource
}

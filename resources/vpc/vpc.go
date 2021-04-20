package vpc

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/jsii-runtime-go"
)

var (
	newVpcSourceId = "newVpcSourceId"
)

type Result struct {
	Stack awscdk.Stack
	Vpc   awsec2.Vpc
}

func Init(parentStack awscdk.Stack, stackName *string, props *awscdk.StackProps) Result {
	stack := awscdk.NewStack(parentStack, stackName, props)

	// 建立 VPC
	vpc := newVpc(stack, jsii.String("10.101.0.0/16"), jsii.Number(float64(20)))

	return Result{
		Stack: stack,
		Vpc:   vpc,
	}
}

func newVpc(stack awscdk.Stack, CidrBlock *string, CidrMask *float64) awsec2.Vpc {
	// subnet 設定，這會自動根據該 region 的 AZ 各自產生出 Subnet
	// 以 ap-northeast-1 而言，這會在 ap-northeast-1a / ap-northeast-1c / ap-northeast-1d 各自產生一組 public /private subnet
	var subnets = []*awsec2.SubnetConfiguration{
		&awsec2.SubnetConfiguration{
			Name:       jsii.String("PublicSubnet"),
			SubnetType: awsec2.SubnetType_PUBLIC,
			CidrMask:   CidrMask,
		},
		&awsec2.SubnetConfiguration{
			Name:       jsii.String("PrivateSubnet"),
			SubnetType: awsec2.SubnetType_PRIVATE,
			CidrMask:   CidrMask,
		},
	}
	resource := awsec2.NewVpc(stack, jsii.String("newVpcPreview"), &awsec2.VpcProps{
		Cidr:                CidrBlock,
		EnableDnsHostnames:  jsii.Bool(true),
		EnableDnsSupport:    jsii.Bool(true),
		SubnetConfiguration: &subnets,
	})

	return resource
}

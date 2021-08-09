package vpc

import (
	"github.com/andy-demo/gocdk/libs/stack_helper"
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/jsii-runtime-go"
	"os"
)

var (
	newVpcSourceId = "newVpcSourceId"
)

type VpcResult struct {
	Stack awscdk.Stack
	Vpc   awsec2.Vpc
}

func Init(parentStack awscdk.Stack, stackName *string, props *awscdk.StackProps) VpcResult {
	stack := awscdk.NewStack(parentStack, stackName, props)

	// 建立 VPC
	vpc := newVpc(stack, jsii.String(os.Getenv("VPC_IP_RANGE")), jsii.Number(float64(20)))

	return VpcResult{
		Stack: stack,
		Vpc:   vpc,
	}
}

func GetSubnet(vpc awsec2.Vpc, subnetType awsec2.SubnetType) awsec2.SubnetSelection {
	switch subnetType {
	case awsec2.SubnetType_PRIVATE:
		return awsec2.SubnetSelection{
			Subnets: vpc.PrivateSubnets(),
		}
	case awsec2.SubnetType_PUBLIC:
		return awsec2.SubnetSelection{
			Subnets: vpc.PublicSubnets(),
		}
	case awsec2.SubnetType_ISOLATED:
	default:
		return awsec2.SubnetSelection{
			Subnets: vpc.IsolatedSubnets(),
		}
	}
	return awsec2.SubnetSelection{
		Subnets: vpc.PublicSubnets(),
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
	}
	resource := awsec2.NewVpc(stack, jsii.String(stack_helper.GenerateNameForResource("vpc-no-private")), &awsec2.VpcProps{
		Cidr:                CidrBlock,
		EnableDnsHostnames:  jsii.Bool(true),
		EnableDnsSupport:    jsii.Bool(true),
		SubnetConfiguration: &subnets,
	})
	awscdk.Tags_Of(resource).Add(
		jsii.String(os.Getenv("TAG_ENVTYPE_NAME")),
		jsii.String(os.Getenv("ENVTYPE")),
		&awscdk.TagProps{
			IncludeResourceTypes: &[]*string{
				jsii.String("AWS::EC2::Subnet"),
				jsii.String("AWS::EC2::VPC"),
			},
		},
	)
	awscdk.Tags_Of(resource).Add(
		jsii.String(os.Getenv("TAG_SERVICETYPE_NAME")),
		jsii.String("SUBNET"),
		&awscdk.TagProps{
			IncludeResourceTypes: &[]*string{
				jsii.String("AWS::EC2::Subnet"),
				jsii.String("AWS::EC2::VPC"),
			},
		},
	)
	awscdk.Tags_Of(resource).Add(
		jsii.String(os.Getenv("TAG_SERVICETYPE_NAME")),
		jsii.String("VPC"),
		&awscdk.TagProps{
			IncludeResourceTypes: &[]*string{
				jsii.String("AWS::EC2::VPC"),
			},
		},
	)

	return resource
}

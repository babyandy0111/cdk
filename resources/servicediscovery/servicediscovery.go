package servicediscovery

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/awsservicediscovery"
	"github.com/aws/jsii-runtime-go"
	"os"
)

type servicediscoveryStack struct {
	Stack awscdk.Stack
	Vpc   awsec2.Vpc
}

func NewServiceDiscovery(parentStack awscdk.Stack, stackName *string, vpc awsec2.Vpc, props *awscdk.StackProps) *servicediscoveryStack {
	stack := awscdk.NewStack(parentStack, stackName, props)

	return &servicediscoveryStack{
		Stack: stack,
		Vpc:   vpc,
	}
}

// 建立給核心系統使用的 namespace
func (stack *servicediscoveryStack) NewInternalPrivateDnsNamespace(name, description string) awsservicediscovery.PrivateDnsNamespace {
	namespace := awsservicediscovery.NewPrivateDnsNamespace(stack.Stack, jsii.String(name), &awsservicediscovery.PrivateDnsNamespaceProps{
		Name:        jsii.String(name),
		Description: jsii.String(description),
		Vpc:         stack.Vpc,
	})
	return namespace
}

// 建立給客戶使用的 namespace
func (stack *servicediscoveryStack) NewInternalClientDnsNamespace(name, description string) awsservicediscovery.PrivateDnsNamespace {
	namespace := awsservicediscovery.NewPrivateDnsNamespace(stack.Stack, jsii.String(name), &awsservicediscovery.PrivateDnsNamespaceProps{
		Name:        jsii.String(name),
		Description: jsii.String(description),
		Vpc:         stack.Vpc,
	})
	awscdk.Tags_Of(namespace).Add(jsii.String(os.Getenv("TAG_ENVTYPE_NAME")), jsii.String(os.Getenv("ENVTYPE")), &awscdk.TagProps{
		IncludeResourceTypes: &[]*string{
			jsii.String("AWS::ServiceDiscovery::PrivateDnsNamespace"),
		},
	})
	awscdk.Tags_Of(namespace).Add(jsii.String(os.Getenv("TAG_SERVICETYPE_NAME")), jsii.String("PRIVATE_DNS_NAMESPACE"), &awscdk.TagProps{
		IncludeResourceTypes: &[]*string{
			jsii.String("AWS::ServiceDiscovery::PrivateDnsNamespace"),
		},
	})
	return namespace
}

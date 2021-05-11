package ec2

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/awselasticloadbalancingv2"
	"github.com/aws/jsii-runtime-go"
)

func NewLoadBalancer(parentstack awscdk.Stack, name *string, props *awscdk.StackProps, vpc awsec2.Vpc, subnet *awsec2.SubnetSelection, acm awscertificatemanager.Certificate) (awscdk.Stack, awselasticloadbalancingv2.ApplicationLoadBalancer, awselasticloadbalancingv2.ApplicationListener) {
	stack := awscdk.NewStack(parentstack, name, props)

	resourceLoadBalancer := awselasticloadbalancingv2.NewApplicationLoadBalancer(stack, jsii.String("Preview-EKS-LoadBalancer"), &awselasticloadbalancingv2.ApplicationLoadBalancerProps{
		Vpc:                vpc,
		DeletionProtection: jsii.Bool(false),
		InternetFacing:     jsii.Bool(true),
		LoadBalancerName:   jsii.String("Preview-EKS-LoadBalancer"),
		VpcSubnets:         subnet,
		Http2Enabled:       jsii.Bool(true),
		IdleTimeout:        awscdk.Duration_Minutes(jsii.Number(float64(5))),
		IpAddressType:      awselasticloadbalancingv2.IpAddressType_IPV4,
		//SecurityGroup:      nil,
	})
	// 建立 Target Group
	targetGroup := awselasticloadbalancingv2.NewApplicationTargetGroup(stack, jsii.String("Preview-EKS-Main-TargetGroup"), &awselasticloadbalancingv2.ApplicationTargetGroupProps{
		DeregistrationDelay: awscdk.Duration_Seconds(jsii.Number(float64(60))),
		HealthCheck: &awselasticloadbalancingv2.HealthCheck{
			Enabled:                 jsii.Bool(true),
			HealthyGrpcCodes:        nil,
			HealthyHttpCodes:        nil,
			HealthyThresholdCount:   jsii.Number(float64(5)),
			Interval:                awscdk.Duration_Seconds(jsii.Number(float64(30))),
			Path:                    jsii.String("/"),
			Port:                    jsii.String("80"),
			Protocol:                awselasticloadbalancingv2.Protocol_HTTP,
			Timeout:                 awscdk.Duration_Seconds(jsii.Number(float64(5))),
			UnhealthyThresholdCount: jsii.Number(float64(3)),
		},
		TargetGroupName: jsii.String("Preview-EKS-Main-TargetGroup"),
		TargetType:      awselasticloadbalancingv2.TargetType_IP,
		Vpc:             vpc,
		Port:            jsii.Number(float64(80)),
		Protocol:        awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		ProtocolVersion: awselasticloadbalancingv2.ApplicationProtocolVersion_HTTP2,
		SlowStart:       awscdk.Duration_Seconds(jsii.Number(float64(60))),
	})
	// 替 Target Group 加上 tag
	awscdk.Tags_Of(targetGroup).Add(jsii.String("Preview"), jsii.String("true"), nil)

	// 建立 Listener Certificate
	listenerCertificate := awselasticloadbalancingv2.ListenerCertificate_FromCertificateManager(acm)
	// 建立 Listener Action
	//defaultListenerAction := awselasticloadbalancingv2.NewListenerAction(&awselasticloadbalancingv2.CfnListener_ActionProperty{
	//	Type:           jsii.String("forward"),
	//	TargetGroupArn: targetGroup.TargetGroupArn(),
	//}, nil)

	// 建立 443 Listener
	resourceListener := awselasticloadbalancingv2.NewApplicationListener(stack, jsii.String("Preview-EKS-LoadBalancer-Listener"), &awselasticloadbalancingv2.ApplicationListenerProps{
		Certificates: &[]awselasticloadbalancingv2.IListenerCertificate{listenerCertificate},
		//DefaultAction: defaultListenerAction,
		Open:         jsii.Bool(true),
		Port:         jsii.Number(float64(443)),
		Protocol:     awselasticloadbalancingv2.ApplicationProtocol_HTTPS,
		SslPolicy:    awselasticloadbalancingv2.SslPolicy_FORWARD_SECRECY_TLS12,
		LoadBalancer: resourceLoadBalancer,
		DefaultTargetGroups: &[]awselasticloadbalancingv2.IApplicationTargetGroup{
			targetGroup,
		},
	})
	return stack, resourceLoadBalancer, resourceListener
}

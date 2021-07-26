package ecs

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/awsservicediscovery"
	"github.com/aws/jsii-runtime-go"
	"os"
)

type ECSStack struct {
	Stack                     awscdk.Stack
	Role                      awsiam.Role
	Cluster                   awsecs.Cluster
	Vpc                       awsec2.Vpc
	CloudMapNamespace         awsservicediscovery.PrivateDnsNamespace
	LB                        awselasticloadbalancingv2.ApplicationLoadBalancer
	LBSecurityGroup           awsec2.SecurityGroup
	ContainerSecurityGroup    awsec2.SecurityGroup
	Listener80                awselasticloadbalancingv2.ApplicationListener
	Listener443               awselasticloadbalancingv2.ApplicationListener
	DefaultTargetGroup        awselasticloadbalancingv2.ApplicationTargetGroup
	CloudMapNamespacesMapping map[string]awsservicediscovery.PrivateDnsNamespace
}

func NewECS(parentStack awscdk.Stack, stackName *string, vpc awsec2.Vpc, cert awscertificatemanager.Certificate, props *awscdk.StackProps) *ECSStack {
	stack := awscdk.NewStack(parentStack, stackName, props)
	// 建立 ECS 專用 Role
	role := CreateECSGeneralRole(stack, "preview-general-role", "General Role For ECS Task")
	// 建立專用 SecurityGroup for loadbalancer
	sg := awsec2.NewSecurityGroup(stack, jsii.String("preview-api-man-sg"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(true),
		Description:       jsii.String("Main Load Balancer For Service"),
		SecurityGroupName: jsii.String("preview-api-lb-sg"),
	})
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(80)), jsii.String("HTTP Entry"), jsii.Bool(false))
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(443)), jsii.String("HTTPS Entry"), jsii.Bool(false))

	// 建立security group for ecs container
	containerSG := awsec2.NewSecurityGroup(stack, jsii.String("preview-api-container-sg"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(true),
		Description:       jsii.String("Default Security Group For ECS Container"),
		SecurityGroupName: jsii.String("preview-api-container-sg"),
	})
	containerSG.AddIngressRule(awsec2.Peer_Ipv4(jsii.String("10.0.0.0/8")), awsec2.Port_Tcp(jsii.Number(80)), jsii.String("For HTTP Frontend"), jsii.Bool(false))
	containerSG.AddIngressRule(awsec2.Peer_Ipv4(jsii.String("10.0.0.0/8")), awsec2.Port_Tcp(jsii.Number(5001)), jsii.String("For GRPC Webpackager"), jsii.Bool(false))
	containerSG.AddIngressRule(awsec2.Peer_Ipv4(jsii.String("10.0.0.0/8")), awsec2.Port_Tcp(jsii.Number(6379)), jsii.String("For Redis Usage"), jsii.Bool(false))
	containerSG.AddIngressRule(awsec2.Peer_Ipv4(jsii.String("10.0.0.0/8")), awsec2.Port_Tcp(jsii.Number(8080)), jsii.String("For HTTP Backend"), jsii.Bool(false))
	awscdk.Tags_Of(containerSG).Add(jsii.String(os.Getenv("TAG_ENVTYPE_NAME")), jsii.String(os.Getenv("ENVTYPE")), &awscdk.TagProps{})
	awscdk.Tags_Of(containerSG).Add(jsii.String(os.Getenv("TAG_SERVICETYPE_NAME")), jsii.String("CONTAINER_SECURITY_GROUP"), &awscdk.TagProps{})
	// 建立專用 Application Load Balancer
	lb := awselasticloadbalancingv2.NewApplicationLoadBalancer(stack, jsii.String("preview-api-main-lb"), &awselasticloadbalancingv2.ApplicationLoadBalancerProps{
		Vpc:                vpc,
		DeletionProtection: jsii.Bool(false),
		InternetFacing:     jsii.Bool(true),
		LoadBalancerName:   jsii.String("preview-api-main-lb"),
		Http2Enabled:       jsii.Bool(true),
		IdleTimeout:        awscdk.Duration_Seconds(jsii.Number(60)),
		IpAddressType:      awselasticloadbalancingv2.IpAddressType_IPV4,
		SecurityGroup:      sg,
	})
	// 加上標籤
	awscdk.Tags_Of(lb).Add(jsii.String(os.Getenv("TAG_ENVTYPE_NAME")), jsii.String(os.Getenv("ENVTYPE")), &awscdk.TagProps{})
	awscdk.Tags_Of(lb).Add(jsii.String(os.Getenv("TAG_SERVICETYPE_NAME")), jsii.String("LOADBALANCER"), &awscdk.TagProps{})
	// 建立預設 TargetGroup ，只要不符合規則的都丟進這裡
	defaultTargetGroup := awselasticloadbalancingv2.NewApplicationTargetGroup(stack, jsii.String("preview-main-apigateway"), &awselasticloadbalancingv2.ApplicationTargetGroupProps{
		DeregistrationDelay: awscdk.Duration_Seconds(jsii.Number(10)),
		HealthCheck: &awselasticloadbalancingv2.HealthCheck{
			Enabled:                 jsii.Bool(true),
			HealthyHttpCodes:        jsii.String("200-499"),
			HealthyThresholdCount:   jsii.Number(5),
			Interval:                awscdk.Duration_Seconds(jsii.Number(10)),
			Path:                    jsii.String("/"),
			Port:                    jsii.String("8080"),
			Protocol:                awselasticloadbalancingv2.Protocol_HTTP,
			Timeout:                 awscdk.Duration_Seconds(jsii.Number(7)),
			UnhealthyThresholdCount: jsii.Number(3),
		},
		TargetGroupName: jsii.String("preview-main-apigateway"),
		TargetType:      awselasticloadbalancingv2.TargetType_IP,
		Vpc:             vpc,
		Port:            jsii.Number(8080),
		Protocol:        awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		ProtocolVersion: awselasticloadbalancingv2.ApplicationProtocolVersion_HTTP1,
	})
	listener80 := awselasticloadbalancingv2.NewApplicationListener(stack, jsii.String("Listener80"), &awselasticloadbalancingv2.ApplicationListenerProps{
		DefaultAction: awselasticloadbalancingv2.ListenerAction_Redirect(&awselasticloadbalancingv2.RedirectOptions{
			Host:      jsii.String("#{host}"),
			Path:      jsii.String("/#{path}"),
			Permanent: jsii.Bool(true),
			Port:      jsii.String("443"),
			Protocol:  jsii.String(string(awselasticloadbalancingv2.ApplicationProtocol_HTTPS)),
			Query:     jsii.String("#{query}"),
		}),
		Port:         jsii.Number(80),
		Protocol:     awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		LoadBalancer: lb,
	})
	listener443 := awselasticloadbalancingv2.NewApplicationListener(stack, jsii.String("Listener443"), &awselasticloadbalancingv2.ApplicationListenerProps{
		Certificates: &[]awselasticloadbalancingv2.IListenerCertificate{
			cert,
		},
		DefaultTargetGroups: &[]awselasticloadbalancingv2.IApplicationTargetGroup{
			defaultTargetGroup,
		},
		Port:         jsii.Number(443),
		Protocol:     awselasticloadbalancingv2.ApplicationProtocol_HTTPS,
		SslPolicy:    awselasticloadbalancingv2.SslPolicy_RECOMMENDED,
		LoadBalancer: lb,
	})
	return &ECSStack{
		Stack:                     stack,
		Role:                      role,
		Vpc:                       vpc,
		LB:                        lb,
		LBSecurityGroup:           sg,
		ContainerSecurityGroup:    containerSG,
		Listener80:                listener80,
		Listener443:               listener443,
		DefaultTargetGroup:        defaultTargetGroup,
		CloudMapNamespacesMapping: make(map[string]awsservicediscovery.PrivateDnsNamespace, 0),
	}
}

func (stack *ECSStack) SetCloudmapDnsNamespace(namespace awsservicediscovery.PrivateDnsNamespace) {
	stack.CloudMapNamespace = namespace
}
func (stack *ECSStack) SetCloudmapDnsNamespacesMapping(mapName string, namespace awsservicediscovery.PrivateDnsNamespace) {
	stack.CloudMapNamespacesMapping[mapName] = namespace
}

func CreateECSGeneralRole(stack awscdk.Stack, roleName string, description string) awsiam.Role {
	role := awsiam.NewRole(stack, jsii.String("roleName"), &awsiam.RoleProps{
		AssumedBy:   awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		Description: jsii.String(description),
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"DEFAULT": awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
				AssignSids: jsii.Bool(true),
				Statements: &[]awsiam.PolicyStatement{
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Actions: &[]*string{
							jsii.String("ecr:*"),
							jsii.String("ecs:*"),
							jsii.String("logs:CreateLogStream"),
							jsii.String("logs:PutLogEvents"),
							jsii.String("application-autoscaling:*"),
							jsii.String("cloudwatch:DescribeAlarms"),
							jsii.String("sns:Publish"),
							jsii.String("ec2:DescribeSecurityGroups"),
							jsii.String("ec2:DescribeSubnets"),
						},
						Effect: awsiam.Effect_ALLOW,
						Resources: &[]*string{
							jsii.String("*"),
						},
						Sid: jsii.String("PreviewEcsRole"),
					}),
				},
			}),
		},
		Path:     jsii.String("/"),
		RoleName: jsii.String(roleName),
	})
	return role
}

func (stack *ECSStack) CreateCluster(name string, vpc awsec2.IVpc) {
	resource := awsecs.NewCluster(stack.Stack, jsii.String(name), &awsecs.ClusterProps{
		ClusterName: jsii.String(name),
		Vpc:         vpc,
	})
	awscdk.Tags_Of(resource).Add(
		jsii.String(os.Getenv("TAG_ENVTYPE_NAME")),
		jsii.String(os.Getenv("ENVTYPE")),
		&awscdk.TagProps{
			IncludeResourceTypes: &[]*string{
				jsii.String("AWS::ECS::Cluster"),
			},
		},
	)
	// 將 cluster 資訊加入到物件成員
	stack.Cluster = resource
	awscdk.Tags_Of(resource).Add(
		jsii.String(os.Getenv("TAG_SERVICETYPE_NAME")),
		jsii.String("ECSCLUSTER"),
		&awscdk.TagProps{
			IncludeResourceTypes: &[]*string{
				jsii.String("AWS::ECS::Cluster"),
			},
		},
	)
}

func (stack *ECSStack) generateMapPointer(env map[string]string) map[string]*string {
	var returns = make(map[string]*string, 0)
	if len(env) > 0 {
		for k, v := range env {
			returns[k] = jsii.String(v)
		}
	}
	return returns
}

// 設定 Backend 所使用的Task Definition
func (stack *ECSStack) RegisterTaskDefinitionAPIManagementBackend(name string, env map[string]string) awsecs.TaskDefinition {
	backendLogGroup := awslogs.NewLogGroup(stack.Stack, jsii.String("preview-backend-api"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("preview-backend-api"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     awslogs.RetentionDays_ONE_MONTH,
	})
	backendLogDriver := awsecs.NewAwsLogDriver(&awsecs.AwsLogDriverProps{
		StreamPrefix: jsii.String("log-message-"),
		LogGroup:     backendLogGroup,
	})
	redisLogGroup := awslogs.NewLogGroup(stack.Stack, jsii.String("preview-backend-redis"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("preview-backend-redis"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     awslogs.RetentionDays_ONE_MONTH,
	})
	redisLogDriver := awsecs.NewAwsLogDriver(&awsecs.AwsLogDriverProps{
		StreamPrefix: jsii.String("log-message-"),
		LogGroup:     redisLogGroup,
	})
	def := awsecs.NewTaskDefinition(stack.Stack, jsii.String(name), &awsecs.TaskDefinitionProps{
		ExecutionRole: stack.Role,
		Family:        jsii.String(name),
		TaskRole:      stack.Role,
		Compatibility: awsecs.Compatibility_FARGATE,
		Cpu:           jsii.String("256"),
		MemoryMiB:     jsii.String("512"),
	})
	envContent := stack.generateMapPointer(env)
	// 以下環境變數直接參照 cloudformation 內容，不會設置在 deployment.yaml
	envContent["AWS_LB_DOMAIN"] = stack.LB.LoadBalancerDnsName()
	envContent["AWS_VPC_ID"] = stack.Vpc.VpcId()
	envContent["AWS_ECS_CLUSTER"] = stack.Cluster.ClusterName()
	envContent["AWS_ECS_TASK_ROLE_ARN"] = stack.Role.RoleArn()
	envContent["AWS_ECS_TASK_EXEC_ARN"] = stack.Role.RoleArn()
	envContent["AWS_CM_CLIENT_PRIVATE_DOMAIN"] = stack.CloudMapNamespacesMapping["client"].NamespaceName()
	envContent["AWS_CM_MANAGEMENT_PRIVATE_DOMAIN"] = stack.CloudMapNamespacesMapping["management"].NamespaceName()
	envContent["AWS_CLIENT_INTERNAL_DOMAIN_ID"] = stack.CloudMapNamespacesMapping["client"].NamespaceId()
	envContent["AWS_MANAGEMENT_INTERNAL_DOMAIN_ID"] = stack.CloudMapNamespacesMapping["management"].NamespaceId()
	backendContainer := awsecs.ContainerImage_FromRegistry(jsii.String("babyandy0111/api-automation-backend:latest"), &awsecs.RepositoryImageProps{})
	def.AddContainer(jsii.String("api-backend"), &awsecs.ContainerDefinitionOptions{
		Image:                backendContainer,
		Cpu:                  jsii.Number(128),
		DisableNetworking:    jsii.Bool(false),
		Environment:          &envContent,
		Essential:            jsii.Bool(true),
		Logging:              backendLogDriver,
		MemoryLimitMiB:       jsii.Number(256),
		MemoryReservationMiB: jsii.Number(256),
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(8080),
				HostPort:      jsii.Number(8080),
				Protocol:      awsecs.Protocol_TCP,
			},
		},
		StartTimeout: awscdk.Duration_Seconds(jsii.Number(10)),
		StopTimeout:  awscdk.Duration_Seconds(jsii.Number(10)),
	})
	redisImage := awsecs.ContainerImage_FromRegistry(jsii.String("redis:6.2.3-alpine"), &awsecs.RepositoryImageProps{})
	redisContainer := def.AddContainer(jsii.String("api-backend-redis"), &awsecs.ContainerDefinitionOptions{
		Image:                redisImage,
		Cpu:                  jsii.Number(128),
		DisableNetworking:    jsii.Bool(false),
		Environment:          &envContent,
		Essential:            jsii.Bool(true),
		Logging:              redisLogDriver,
		MemoryLimitMiB:       jsii.Number(256),
		MemoryReservationMiB: jsii.Number(256),
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(6379),
				HostPort:      jsii.Number(6379),
				Protocol:      awsecs.Protocol_TCP,
			},
		},
		StartTimeout: awscdk.Duration_Seconds(jsii.Number(10)),
		StopTimeout:  awscdk.Duration_Seconds(jsii.Number(10)),
	})

	// 建立 TargetGroup
	targetgroup := awselasticloadbalancingv2.NewApplicationTargetGroup(stack.Stack, jsii.String("api-backend-targetgroup"), &awselasticloadbalancingv2.ApplicationTargetGroupProps{
		DeregistrationDelay: awscdk.Duration_Seconds(jsii.Number(10)),
		HealthCheck: &awselasticloadbalancingv2.HealthCheck{
			Enabled:                 jsii.Bool(true),
			HealthyHttpCodes:        jsii.String("200-499"),
			HealthyThresholdCount:   jsii.Number(5),
			Interval:                awscdk.Duration_Seconds(jsii.Number(10)),
			Path:                    jsii.String("/"),
			Port:                    jsii.String("8080"),
			Protocol:                awselasticloadbalancingv2.Protocol_HTTP,
			Timeout:                 awscdk.Duration_Seconds(jsii.Number(7)),
			UnhealthyThresholdCount: jsii.Number(3),
		},
		TargetGroupName: jsii.String("preview-main-backend"),
		TargetType:      awselasticloadbalancingv2.TargetType_IP,
		Vpc:             stack.Vpc,
		Port:            jsii.Number(8080),
		Protocol:        awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		ProtocolVersion: awselasticloadbalancingv2.ApplicationProtocolVersion_HTTP1,
	})
	// 建立 ListenerRule
	awselasticloadbalancingv2.NewApplicationListenerRule(stack.Stack, jsii.String("api-backend-listenerrule"), &awselasticloadbalancingv2.ApplicationListenerRuleProps{
		Priority: jsii.Number(1),
		Conditions: &[]awselasticloadbalancingv2.ListenerCondition{
			awselasticloadbalancingv2.ListenerCondition_HostHeaders(&[]*string{
				jsii.String(fmt.Sprintf("my-api.%s", os.Getenv("ACM_MAIN_DOMAIN"))),
			}),
		},
		TargetGroups: &[]awselasticloadbalancingv2.IApplicationTargetGroup{
			targetgroup,
		},
		Listener: stack.Listener443,
	})

	// 建立 Service
	service := awsecs.NewFargateService(stack.Stack, jsii.String("preview-ecs-backend"), &awsecs.FargateServiceProps{
		Cluster: stack.Cluster,
		CloudMapOptions: &awsecs.CloudMapOptions{
			CloudMapNamespace: stack.CloudMapNamespace,
			Container:         redisContainer,
			ContainerPort:     jsii.Number(6379),
			DnsRecordType:     "A",
			DnsTtl:            awscdk.Duration_Seconds(jsii.Number(300)),
			Name:              jsii.String("redis-backend"),
		},
		DesiredCount:           jsii.Number(1),
		HealthCheckGracePeriod: awscdk.Duration_Seconds(jsii.Number(10)),
		MaxHealthyPercent:      jsii.Number(200),
		MinHealthyPercent:      jsii.Number(100),
		ServiceName:            jsii.String("preview-api-backend"),
		TaskDefinition:         def,
		AssignPublicIp:         jsii.Bool(false),
		SecurityGroups: &[]awsec2.ISecurityGroup{
			stack.ContainerSecurityGroup,
		},
		VpcSubnets: &awsec2.SubnetSelection{Subnets: stack.Vpc.PrivateSubnets()},
	})
	// 註冊進targetgroup
	targetgroup.AddTarget([]awselasticloadbalancingv2.IApplicationLoadBalancerTarget{
		service,
	}...)

	return def
}
func (stack *ECSStack) RegisterTaskDefinitionAPIManagementFrontend(name string) awsecs.TaskDefinition {
	frontendLogGroup := awslogs.NewLogGroup(stack.Stack, jsii.String("preview-frontend-api"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("preview-frontend-api"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     awslogs.RetentionDays_ONE_MONTH,
	})
	frontendLogDriver := awsecs.NewAwsLogDriver(&awsecs.AwsLogDriverProps{
		StreamPrefix: jsii.String("log-message-"),
		LogGroup:     frontendLogGroup,
	})
	webpushPackagerLogGroup := awslogs.NewLogGroup(stack.Stack, jsii.String("preview-webpush-packager"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("preview-webpush-packager"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     awslogs.RetentionDays_ONE_MONTH,
	})
	webpushLogDriver := awsecs.NewAwsLogDriver(&awsecs.AwsLogDriverProps{
		StreamPrefix: jsii.String("log-message-"),
		LogGroup:     webpushPackagerLogGroup,
	})
	def := awsecs.NewTaskDefinition(stack.Stack, jsii.String(name), &awsecs.TaskDefinitionProps{
		ExecutionRole: stack.Role,
		Family:        jsii.String(name),
		TaskRole:      stack.Role,
		Compatibility: awsecs.Compatibility_FARGATE,
		Cpu:           jsii.String("256"),
		MemoryMiB:     jsii.String("512"),
	})
	frontendImage := awsecs.ContainerImage_FromRegistry(jsii.String("babyandy0111/api-automation-frontend:latest"), &awsecs.RepositoryImageProps{})
	webpushImage := awsecs.ContainerImage_FromRegistry(jsii.String("babyandy0111/grpc-webpush:latest"), &awsecs.RepositoryImageProps{})
	def.AddContainer(jsii.String("api-frontend"), &awsecs.ContainerDefinitionOptions{
		Image:                frontendImage,
		Cpu:                  jsii.Number(128),
		DisableNetworking:    jsii.Bool(false),
		Essential:            jsii.Bool(true),
		Logging:              frontendLogDriver,
		MemoryLimitMiB:       jsii.Number(256),
		MemoryReservationMiB: jsii.Number(256),
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(80),
				HostPort:      jsii.Number(80),
				Protocol:      awsecs.Protocol_TCP,
			},
		},
		StartTimeout: awscdk.Duration_Seconds(jsii.Number(10)),
		StopTimeout:  awscdk.Duration_Seconds(jsii.Number(10)),
	})
	webpushContainer := def.AddContainer(jsii.String("webpush-packager"), &awsecs.ContainerDefinitionOptions{
		Image:                webpushImage,
		Cpu:                  jsii.Number(128),
		DisableNetworking:    jsii.Bool(false),
		Essential:            jsii.Bool(true),
		Logging:              webpushLogDriver,
		MemoryLimitMiB:       jsii.Number(256),
		MemoryReservationMiB: jsii.Number(256),
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(5001),
				HostPort:      jsii.Number(5001),
				Protocol:      awsecs.Protocol_TCP,
			},
		},
		Environment: &map[string]*string{
			"PORT":                  jsii.String("5001"),
			"AWS_ACCESS_KEY_ID":     jsii.String(os.Getenv("ACCESS_KEY")),
			"AWS_SECRET_ACCESS_KEY": jsii.String(os.Getenv("SECRET_KEY")),
		},
		StartTimeout: awscdk.Duration_Seconds(jsii.Number(10)),
		StopTimeout:  awscdk.Duration_Seconds(jsii.Number(10)),
	})

	// 建立 TargetGroup
	targetgroup := awselasticloadbalancingv2.NewApplicationTargetGroup(stack.Stack, jsii.String("api-frontend-targetgroup"), &awselasticloadbalancingv2.ApplicationTargetGroupProps{
		DeregistrationDelay: awscdk.Duration_Seconds(jsii.Number(10)),
		HealthCheck: &awselasticloadbalancingv2.HealthCheck{
			Enabled:                 jsii.Bool(true),
			HealthyHttpCodes:        jsii.String("200-499"),
			HealthyThresholdCount:   jsii.Number(5),
			Interval:                awscdk.Duration_Seconds(jsii.Number(10)),
			Path:                    jsii.String("/"),
			Port:                    jsii.String("80"),
			Protocol:                awselasticloadbalancingv2.Protocol_HTTP,
			Timeout:                 awscdk.Duration_Seconds(jsii.Number(3)),
			UnhealthyThresholdCount: jsii.Number(3),
		},
		TargetGroupName: jsii.String("preview-main-frontend"),
		TargetType:      awselasticloadbalancingv2.TargetType_IP,
		Vpc:             stack.Vpc,
		Port:            jsii.Number(80),
		Protocol:        awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		ProtocolVersion: awselasticloadbalancingv2.ApplicationProtocolVersion_HTTP1,
	})
	// 建立 ListenerRule
	awselasticloadbalancingv2.NewApplicationListenerRule(stack.Stack, jsii.String("api-frontend-listenerrule"), &awselasticloadbalancingv2.ApplicationListenerRuleProps{
		Priority: jsii.Number(2),
		Conditions: &[]awselasticloadbalancingv2.ListenerCondition{
			awselasticloadbalancingv2.ListenerCondition_HostHeaders(&[]*string{
				jsii.String(fmt.Sprintf("my-api-management.%s", os.Getenv("ACM_MAIN_DOMAIN"))),
			}),
		},
		TargetGroups: &[]awselasticloadbalancingv2.IApplicationTargetGroup{
			targetgroup,
		},
		Listener: stack.Listener443,
	})

	// 建立 ECS Service
	service := awsecs.NewFargateService(stack.Stack, jsii.String("preview-ecs-frontend"), &awsecs.FargateServiceProps{
		Cluster: stack.Cluster,
		CloudMapOptions: &awsecs.CloudMapOptions{
			CloudMapNamespace: stack.CloudMapNamespace,
			Container:         webpushContainer,
			ContainerPort:     jsii.Number(5001),
			DnsRecordType:     "A",
			DnsTtl:            awscdk.Duration_Seconds(jsii.Number(300)),
			Name:              jsii.String("webpush"),
		},
		DesiredCount:           jsii.Number(1),
		HealthCheckGracePeriod: awscdk.Duration_Seconds(jsii.Number(10)),
		MaxHealthyPercent:      jsii.Number(200),
		MinHealthyPercent:      jsii.Number(100),
		ServiceName:            jsii.String("preview-api-frontend"),
		TaskDefinition:         def,
		AssignPublicIp:         jsii.Bool(false),
		SecurityGroups: &[]awsec2.ISecurityGroup{
			stack.ContainerSecurityGroup,
		},
		VpcSubnets: &awsec2.SubnetSelection{Subnets: stack.Vpc.PrivateSubnets()},
	})
	// 註冊進targetgroup
	targetgroup.AddTarget([]awselasticloadbalancingv2.IApplicationLoadBalancerTarget{
		service,
	}...)
	return def
}
func (stack *ECSStack) RegisterTaskDefinitionAPIGateway(name string, env map[string]string) awsecs.TaskDefinition {
	apigatewayLogGroup := awslogs.NewLogGroup(stack.Stack, jsii.String("preview-backend-apigateway"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("preview-backend-apigateway"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     awslogs.RetentionDays_ONE_MONTH,
	})
	apigatewayLogDriver := awsecs.NewAwsLogDriver(&awsecs.AwsLogDriverProps{
		StreamPrefix: jsii.String("log-message-"),
		LogGroup:     apigatewayLogGroup,
	})
	redisLogGroup := awslogs.NewLogGroup(stack.Stack, jsii.String("preview-apigateway-redis"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("preview-apigateway-redis"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     awslogs.RetentionDays_ONE_MONTH,
	})
	redisLogDriver := awsecs.NewAwsLogDriver(&awsecs.AwsLogDriverProps{
		StreamPrefix: jsii.String("log-message-"),
		LogGroup:     redisLogGroup,
	})
	def := awsecs.NewTaskDefinition(stack.Stack, jsii.String(name), &awsecs.TaskDefinitionProps{
		ExecutionRole: stack.Role,
		Family:        jsii.String(name),
		TaskRole:      stack.Role,
		Compatibility: awsecs.Compatibility_FARGATE,
		Cpu:           jsii.String("256"),
		MemoryMiB:     jsii.String("512"),
	})
	envContent := stack.generateMapPointer(env)
	backendContainer := awsecs.ContainerImage_FromRegistry(jsii.String("babyandy0111/apigateway:latest"), &awsecs.RepositoryImageProps{})
	def.AddContainer(jsii.String("api-backend"), &awsecs.ContainerDefinitionOptions{
		Image:                backendContainer,
		Cpu:                  jsii.Number(128),
		DisableNetworking:    jsii.Bool(false),
		Environment:          &envContent,
		Essential:            jsii.Bool(true),
		Logging:              apigatewayLogDriver,
		MemoryLimitMiB:       jsii.Number(256),
		MemoryReservationMiB: jsii.Number(256),
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(8080),
				HostPort:      jsii.Number(8080),
				Protocol:      awsecs.Protocol_TCP,
			},
		},
		StartTimeout: awscdk.Duration_Seconds(jsii.Number(10)),
		StopTimeout:  awscdk.Duration_Seconds(jsii.Number(10)),
	})
	redisImage := awsecs.ContainerImage_FromRegistry(jsii.String("redis:6.2.3-alpine"), &awsecs.RepositoryImageProps{})
	redisContainer := def.AddContainer(jsii.String("api-backend-redis"), &awsecs.ContainerDefinitionOptions{
		Image:                redisImage,
		Cpu:                  jsii.Number(128),
		DisableNetworking:    jsii.Bool(false),
		Environment:          &envContent,
		Essential:            jsii.Bool(true),
		Logging:              redisLogDriver,
		MemoryLimitMiB:       jsii.Number(256),
		MemoryReservationMiB: jsii.Number(256),
		PortMappings: &[]*awsecs.PortMapping{
			{
				ContainerPort: jsii.Number(6379),
				HostPort:      jsii.Number(6379),
				Protocol:      awsecs.Protocol_TCP,
			},
		},
		StartTimeout: awscdk.Duration_Seconds(jsii.Number(10)),
		StopTimeout:  awscdk.Duration_Seconds(jsii.Number(10)),
	})
	// 建立 Service
	service := awsecs.NewFargateService(stack.Stack, jsii.String("preview-ecs-apigateway"), &awsecs.FargateServiceProps{
		Cluster: stack.Cluster,
		CloudMapOptions: &awsecs.CloudMapOptions{
			CloudMapNamespace: stack.CloudMapNamespace,
			Container:         redisContainer,
			ContainerPort:     jsii.Number(6379),
			DnsRecordType:     "A",
			DnsTtl:            awscdk.Duration_Seconds(jsii.Number(300)),
			Name:              jsii.String("redis-apigateway"),
		},
		DesiredCount:           jsii.Number(1),
		HealthCheckGracePeriod: awscdk.Duration_Seconds(jsii.Number(10)),
		MaxHealthyPercent:      jsii.Number(200),
		MinHealthyPercent:      jsii.Number(100),
		ServiceName:            jsii.String("preview-apigateway"),
		TaskDefinition:         def,
		AssignPublicIp:         jsii.Bool(false),
		SecurityGroups: &[]awsec2.ISecurityGroup{
			stack.ContainerSecurityGroup,
		},
		VpcSubnets: &awsec2.SubnetSelection{Subnets: stack.Vpc.PrivateSubnets()},
	})
	// 註冊進targetgroup
	stack.DefaultTargetGroup.AddTarget([]awselasticloadbalancingv2.IApplicationLoadBalancerTarget{
		service,
	}...)

	return def
}

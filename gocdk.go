package main

import (
	"fmt"
	"github.com/andy-demo/gocdk/libs/stack_helper"
	"github.com/andy-demo/gocdk/resources/acm"
	"github.com/andy-demo/gocdk/resources/ecs"
	"github.com/andy-demo/gocdk/resources/s3"
	"github.com/andy-demo/gocdk/resources/servicediscovery"
	"github.com/andy-demo/gocdk/resources/vpc"
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	// 解析 .env 作為部署依據
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 解析 deployment.yaml 檔作為 ECS 環境變數依據
	e, err := stack_helper.ParseYAML("./deployment.yaml")
	fmt.Printf("\n%+v\n", e)

	app := awscdk.NewApp(nil)

	accountId := os.Getenv("ACCOUNT_ID")
	region := os.Getenv("REGION")

	props := awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(accountId),
			Region:  jsii.String(region),
		},
		StackName: jsii.String(stack_helper.GenerateNameForResource("RootStack")),
	}

	// 建立 root stack
	rootStack := awscdk.NewStack(app, jsii.String(stack_helper.GenerateNameForResource("RootStack")), &props)
	// 設定部分變數
	awscdk.NewCfnOutput(rootStack, jsii.String("AWS_ACCESS_KEY"), &awscdk.CfnOutputProps{
		Value:       jsii.String(os.Getenv("ACCESS_KEY")),
		Description: jsii.String("Default AWS Access Key"),
		ExportName:  jsii.String("AWS:ACCESS:KEY"),
	})
	awscdk.NewCfnOutput(rootStack, jsii.String("AWS_REGION"), &awscdk.CfnOutputProps{
		Value:       jsii.String(os.Getenv("REGION")),
		Description: jsii.String("Default AWS Region"),
		ExportName:  jsii.String("AWS:REGION"),
	})
	awscdk.NewCfnOutput(rootStack, jsii.String("AWS_SECRET_KEY"), &awscdk.CfnOutputProps{
		Value:       jsii.String(os.Getenv("SECRET_KEY")),
		Description: jsii.String("Default AWS Secret Key"),
		ExportName:  jsii.String("AWS:SECRET:KEY"),
	})
	awscdk.NewCfnOutput(rootStack, jsii.String("CI_USER_TOKEN"), &awscdk.CfnOutputProps{
		Value:       jsii.String(os.Getenv("GITHUB_TOKEN")),
		Description: jsii.String("Default github token"),
		ExportName:  jsii.String("CI:USER:TOKEN"),
	})
	awscdk.NewCfnOutput(rootStack, jsii.String("DOCKER_PASSWORD"), &awscdk.CfnOutputProps{
		Value:       jsii.String(os.Getenv("DOCKER_PASSWORD")),
		Description: jsii.String("Default Dockerhub password"),
		ExportName:  jsii.String("DOCKER:PASSWORD"),
	})
	awscdk.NewCfnOutput(rootStack, jsii.String("DOCKER_USERNAME"), &awscdk.CfnOutputProps{
		Value:       jsii.String(os.Getenv("DOCKER_USERNAME")),
		Description: jsii.String("Default Dockerhub username"),
		ExportName:  jsii.String("DOCKER:USERNAME"),
	})
	awscdk.NewCfnOutput(rootStack, jsii.String("MYSQL_HOST"), &awscdk.CfnOutputProps{
		Value:       jsii.String(e["PRIMARY_ECSTASK_ENV"]["DB_HOST"]),
		Description: jsii.String("Default MySQL HOST"),
		ExportName:  jsii.String("MYSQL:HOST"),
	})
	awscdk.NewCfnOutput(rootStack, jsii.String("MYSQL_PASSWORD"), &awscdk.CfnOutputProps{
		Value:       jsii.String(e["PRIMARY_ECSTASK_ENV"]["DB_PASSWORD"]),
		Description: jsii.String("Default MySQL Password"),
		ExportName:  jsii.String("MYSQL:PASSWORD"),
	})
	awscdk.NewCfnOutput(rootStack, jsii.String("MYSQL_USER"), &awscdk.CfnOutputProps{
		Value:       jsii.String(e["PRIMARY_ECSTASK_ENV"]["DB_USER"]),
		Description: jsii.String("Default MySQL Username"),
		ExportName:  jsii.String("MYSQL:USER"),
	})
	// 建立 VPC Stack
	props.StackName = jsii.String(stack_helper.GenerateNameForResource("VPCStack"))
	Vpc := vpc.Init(rootStack, jsii.String(stack_helper.GenerateNameForResource("VPCStack")), &props)

	// 建立 ACM Stack
	props.StackName = jsii.String(stack_helper.GenerateNameForResource("ACMStack"))
	acmStack, acmResource := acm.NewACM(rootStack, jsii.String(stack_helper.GenerateNameForResource("ACMStack")), &props)
	fmt.Println(*acmStack.StackName())
	fmt.Println(*acmResource.CertificateArn())

	// 建立 cloudmap 相關服務（for service discovery）
	props.StackName = jsii.String(stack_helper.GenerateNameForResource("ServiceDiscoveryStack"))
	serviceDiscoveryStack := servicediscovery.NewServiceDiscovery(rootStack, jsii.String(stack_helper.GenerateNameForResource("ServiceDiscoveryStack")), Vpc.Vpc, &props)
	internalDomain := fmt.Sprintf("%s.management.internal", stack_helper.GetEnv())
	clientDomain := fmt.Sprintf("%s.client.internal", stack_helper.GetEnv())
	internalNamespace := serviceDiscoveryStack.NewInternalPrivateDnsNamespace(internalDomain, "internal service for core system")
	fmt.Println(internalNamespace)
	clientNamespace := serviceDiscoveryStack.NewInternalClientDnsNamespace(clientDomain, "client service for client system")
	fmt.Println(clientNamespace)

	// 建立 S3/ Cloudfront
	props.StackName = jsii.String(stack_helper.GenerateNameForResource("S3Stack"))
	s3Stack := s3.New(rootStack, jsii.String(stack_helper.GenerateNameForResource("S3Stack")), &props)
	s3Stack.Stack.AddDependency(acmStack, jsii.String("Waiting ACM Updated"))
	fsBucket, _, publicKey, defaultDomain := s3Stack.CreateStorageBucket(acmResource, Vpc.Vpc, e["PRIMARY_ECSTASK_ENV"])

	// 建立 ECS 相關服務 (ECS Task Definition / Service / Cloudmap / Load Balancer)
	props.StackName = jsii.String(stack_helper.GenerateNameForResource("ECSStack"))
	ecsStack := ecs.NewECS(rootStack, jsii.String(stack_helper.GenerateNameForResource("ECSStack")), Vpc.Vpc, acmResource, &props)
	ecsStack.Stack.AddDependency(Vpc.Stack, jsii.String("Waiting for VPCStack update"))
	ecsStack.Stack.AddDependency(serviceDiscoveryStack.Stack, jsii.String("Waiting for ServiceDiscoveryStack update"))
	ecsStack.Stack.AddDependency(s3Stack.Stack, jsii.String("Waiting for S3Stack"))
	ecsStack.SetCloudmapDnsNamespace(internalNamespace)
	ecsStack.SetCloudmapDnsNamespacesMapping("management", internalNamespace)
	ecsStack.SetCloudmapDnsNamespacesMapping("client", clientNamespace)
	cluster := ecsStack.CreateCluster(stack_helper.GenerateNameForResource("ECS-Cluster"), Vpc.Vpc)

	// 建立 ECS TaskDefinition
	ecsStack.RegisterTaskDefinitionAPIManagementBackend(stack_helper.GenerateNameForResource("api-main-backend"), cluster, e["PRIMARY_ECSTASK_ENV"], fsBucket)
	ecsStack.RegisterTaskDefinitionAPIManagementFrontend(cluster)
	ecsStack.RegisterTaskDefinitionAPIGateway(cluster, e["APIGATEWAY_ECSTASK_ENV"], publicKey, s3Stack.PackageBucket, fsBucket)
	ecsStack.RegisterTaskDefinitionFixedMySQLGrpcService(cluster)
	ecsStack.RegisterTaskDefinitionNginxThumbService(cluster, *defaultDomain)

	app.Synth(nil)
}

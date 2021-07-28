package main

import (
	"fmt"
	"github.com/andy-demo/gocdk/libs/stack_helper"
	"github.com/andy-demo/gocdk/resources/acm"
	"github.com/andy-demo/gocdk/resources/ecs"
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
		StackName: jsii.String("TestPreviewStackRoot"),
	}

	// 建立 root stack
	rootStack := awscdk.NewStack(app, jsii.String("TestPreviewStackRoot"), &props)
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
	props.StackName = jsii.String("VPCStack")
	Vpc := vpc.Init(rootStack, jsii.String("VPCStack"), &props)

	// 建立 ACM Stack
	props.StackName = jsii.String("ACMStack")
	acmStack, acmResource := acm.NewACM(rootStack, jsii.String("ACMStack"), &props)
	fmt.Println(*acmStack.StackName())
	fmt.Println(*acmResource.CertificateArn())

	// 建立 cloudmap 相關服務（for service discovery）
	props.StackName = jsii.String("ServiceDiscoveryStack")
	serviceDiscoveryStack := servicediscovery.NewServiceDiscovery(rootStack, jsii.String("ServiceDiscoveryStack"), Vpc.Vpc, &props)
	internalNamespace := serviceDiscoveryStack.NewInternalPrivateDnsNamespace("management.internal", "internal service for core system")
	clientNamespace := serviceDiscoveryStack.NewInternalClientDnsNamespace("client.internal", "client service for client system")

	// 建立 ECS 相關服務 (ECS Task Definition / Service / Cloudmap / Load Balancer)
	props.StackName = jsii.String("ECSStack")
	ecsStack := ecs.NewECS(rootStack, jsii.String("ECSStack"), Vpc.Vpc, acmResource, &props)
	ecsStack.Stack.AddDependency(Vpc.Stack, jsii.String("Waiting for VPCStack update"))
	ecsStack.Stack.AddDependency(serviceDiscoveryStack.Stack, jsii.String("Waiting for ServiceDiscoveryStack update"))
	ecsStack.SetCloudmapDnsNamespace(internalNamespace)
	ecsStack.SetCloudmapDnsNamespacesMapping("management", internalNamespace)
	ecsStack.SetCloudmapDnsNamespacesMapping("client", clientNamespace)
	ecsStack.CreateCluster("Preview-ECS-Cluster", Vpc.Vpc)

	// 建立 ECS TaskDefinition
	ecsStack.RegisterTaskDefinitionAPIManagementBackend("preview-api-man-backend", e["PRIMARY_ECSTASK_ENV"])
	ecsStack.RegisterTaskDefinitionAPIManagementFrontend("preview-api-man-frontend")
	ecsStack.RegisterTaskDefinitionAPIGateway("preview-api-man-apigateway", e["APIGATEWAY_ECSTASK_ENV"])

	app.Synth(nil)
}

package main

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/jsii-runtime-go"
	"github.com/faryne/go-cdk-example/libs/stack_helper"
	"github.com/faryne/go-cdk-example/resources/acm"
	"github.com/faryne/go-cdk-example/resources/ecs"
	"github.com/faryne/go-cdk-example/resources/eks"
	"github.com/faryne/go-cdk-example/resources/servicediscovery"
	"github.com/faryne/go-cdk-example/resources/vpc"
	"github.com/joho/godotenv"
	"os"
)

var envType = os.Getenv("ENVTYPE")

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
	// 建立 VPC Stack
	props.StackName = jsii.String("VPCStack")
	Vpc := vpc.Init(rootStack, jsii.String("VPCStack"), &props)

	// 建立 EKS Stack
	props.StackName = jsii.String("EKSStack")
	eksStack := eks.Init(rootStack, jsii.String("EKSStack"), &props, Vpc.Vpc)
	eksStack.Stack.AddDependency(Vpc.Stack, jsii.String("Waiting for VPCStack Done"))

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

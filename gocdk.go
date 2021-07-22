package main

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/jsii-runtime-go"
	"github.com/faryne/go-cdk-example/libs/stack_helper"
	"github.com/faryne/go-cdk-example/resources/acm"
	"github.com/faryne/go-cdk-example/resources/ecs"
	"github.com/faryne/go-cdk-example/resources/eks"
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

	// 建立 ECS Cluster
	props.StackName = jsii.String("ECSStack")
	ecsStack := ecs.NewECS(rootStack, jsii.String("ECSStack"), Vpc.Vpc, acmResource, &props)
	ecsStack.Stack.AddDependency(Vpc.Stack, jsii.String("Waiting for VPCStack update"))
	ecsStack.CreateCluster("Preview-ECS-Cluster", Vpc.Vpc)

	// 建立 ECS TaskDefinition
	ecsStack.RegisterTaskDefinitionAPIManagementBackend("preview-api-man-backend", e["PRIMARY_ECSTASK_ENV"])
	ecsStack.RegisterTaskDefinitionAPIManagementFrontend("preview-api-man-frontend")

	// 建立 Load Balancer Stack
	//subnet := vpc.GetSubnet(Vpc.Vpc, awsec2.SubnetType_PUBLIC)
	//props.StackName = jsii.String("LoadBalancerStack")
	//lbStack, _, _ := ec2.NewLoadBalancer(rootStack, jsii.String("LoadBalancerStack"), &props, Vpc.Vpc, &subnet, acmCertificate)
	//lbStack.AddDependency(Vpc.Stack, jsii.String("Load Balancer needs VPC is set"))
	//lbStack.AddDependency(acmStack, jsii.String("Load Balancer needs ACM is set"))

	app.Synth(nil)
}

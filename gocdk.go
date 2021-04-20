package main

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/jsii-runtime-go"
	"github.com/faryne/go-cdk-example/resources/vpc"
	"os"
)

var envType = os.Getenv("ENVTYPE")

func main() {
	app := awscdk.NewApp(nil)

	accountId := os.Getenv("ACCOUNT_ID")
	region := os.Getenv("REGION")

	props := awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(accountId),
			Region:  jsii.String(region),
		},
	}

	// 建立 root stack
	rootStack := awscdk.NewStack(app, jsii.String("TEST-PreviewRootStack"), &props)
	// 建立 VPC Stack
	vpc.Init(rootStack, jsii.String("VPCStack"), &props)

	// 建立 EKS Stack
	//eksStack := eks.Init(rootStack, jsii.String("EKSStack"), &props)
	//eksStack.AddStackDependency(vpcStack, jsii.String("Reason"))

	app.Synth(nil)
}

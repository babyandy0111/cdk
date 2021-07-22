package eks

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/awseks"
	"github.com/aws/jsii-runtime-go"
	stackHelper "github.com/faryne/go-cdk-example/libs/stack_helper"
	"os"
	"strings"
)

type EKSResult struct {
	Stack awscdk.Stack
}

var stack *stackHelper.MyCDKStack

func Init(parentStack awscdk.Stack, stackName *string, props *awscdk.StackProps, vpc awsec2.Vpc) EKSResult {
	stack := awscdk.NewStack(parentStack, stackName, props)

	// 建立 EKS Fargate Cluster
	//newEksFargetCluster(stack, vpc)
	// 建立 EKS EC2 Cluster
	newEksEC2Cluster(stack, vpc)

	return EKSResult{
		Stack: stack,
	}
}

func GetSubnetSelection(vpc awsec2.Vpc) *[]*awsec2.SubnetSelection {
	return &[]*awsec2.SubnetSelection{
		&awsec2.SubnetSelection{
			Subnets: vpc.PrivateSubnets(),
		},
		//&awsec2.SubnetSelection{
		//	Subnets: vpc.PublicSubnets(),
		//},
	}
}

func newEksFargetCluster(stack awscdk.Stack, vpc awsec2.Vpc) awseks.FargateCluster {
	subnet := GetSubnetSelection(vpc)
	var allowedIPS = awseks.EndpointAccess_PUBLIC_AND_PRIVATE().OnlyFrom(
		jsii.String(os.Getenv("ALLOWED_IP1")),
		jsii.String(os.Getenv("ALLOWED_IP2")),
	)
	resource := awseks.NewFargateCluster(stack, jsii.String("EKSFargateCluster"), &awseks.FargateClusterProps{
		Version:             awseks.KubernetesVersion_V1_19(),
		ClusterName:         jsii.String("PreviewFargateCluster"),
		OutputClusterName:   jsii.Bool(true),
		OutputConfigCommand: jsii.Bool(true),
		//Role:                      nil,
		//SecurityGroup:             nil,
		Vpc:        vpc,
		VpcSubnets: subnet,
		//ClusterHandlerEnvironment: nil,
		CoreDnsComputeType: awseks.CoreDnsComputeType_FARGATE,
		EndpointAccess:     allowedIPS,
		//KubectlEnvironment:        nil,
		//KubectlLayer:              nil,
		KubectlMemory: awscdk.Size_Gibibytes(jsii.Number(float64(1))),
		//MastersRole:               nil,
		OutputMastersRoleArn:     jsii.Bool(true),
		PlaceClusterHandlerInVpc: jsii.Bool(true),
		Prune:                    nil,
		SecretsEncryptionKey:     nil,
		//DefaultProfile: &awseks.FargateProfileOptions{
		//	FargateProfileName: jsii.String("FargateProfile"),
		//	Selectors: &[]*awseks.Selector{
		//		&awseks.Selector{
		//			Namespace: jsii.String("PreviewFargateClusterNamespace"),
		//		},
		//	},
		//},
	})

	// 建立 Argo CD 的 Profile
	labels := make(map[string]*string)
	labels["appName"] = jsii.String("argocd")
	resource.AddFargateProfile(jsii.String("Preview-argocd"), &awseks.FargateProfileOptions{
		Selectors: &[]*awseks.Selector{
			&awseks.Selector{
				Namespace: jsii.String("argocd"),
				Labels:    &labels,
			},
		},
		FargateProfileName: jsii.String("argocd"),
		SubnetSelection: &awsec2.SubnetSelection{
			Subnets: vpc.PrivateSubnets(),
		},
		Vpc: vpc,
	})

	// 再建立 Traefik namespace
	labels["appName"] = jsii.String("traefik")
	resource.AddFargateProfile(jsii.String("Preview-Traefik"), &awseks.FargateProfileOptions{
		Selectors: &[]*awseks.Selector{
			&awseks.Selector{
				Namespace: jsii.String("traefik"),
				Labels:    &labels,
			},
		},
		FargateProfileName: jsii.String("traefik"),
		SubnetSelection: &awsec2.SubnetSelection{
			Subnets: vpc.PrivateSubnets(),
		},
		Vpc: vpc,
	})

	// 建立 OIDC 服務
	awseks.NewOpenIdConnectProvider(stack, jsii.String("PreviewEKSStack-OIDC"), &awseks.OpenIdConnectProviderProps{
		Url: resource.ClusterOpenIdConnectIssuerUrl(),
	})

	return resource
}
func newEksEC2Cluster(stack awscdk.Stack, vpc awsec2.Vpc) awseks.Cluster {
	subnet := GetSubnetSelection(vpc)
	inputIPS := strings.Split(os.Getenv("ALLOWED_IPS"), ",")
	var allowedIpsArray = make([]*string, 0)
	for _, v := range inputIPS {
		allowedIpsArray = append(allowedIpsArray, jsii.String(v))
	}

	resource := awseks.NewCluster(stack, jsii.String("EKSEC2Cluster"), &awseks.ClusterProps{
		Version:             awseks.KubernetesVersion_V1_19(),
		ClusterName:         jsii.String("PreviewEC2Cluster"),
		OutputClusterName:   jsii.Bool(true),
		OutputConfigCommand: jsii.Bool(true),
		//Role:                      nil,
		//SecurityGroup:             nil,
		Vpc:        vpc,
		VpcSubnets: subnet,
		//ClusterHandlerEnvironment: nil,
		CoreDnsComputeType: awseks.CoreDnsComputeType_EC2,
		EndpointAccess:     awseks.EndpointAccess_PUBLIC_AND_PRIVATE().OnlyFrom(allowedIpsArray...),
		//KubectlEnvironment:       nil,
		//KubectlLayer:             nil,
		KubectlMemory: awscdk.Size_Gibibytes(jsii.Number(float64(1))),
		//MastersRole:              nil,
		OutputMastersRoleArn:     jsii.Bool(true),
		PlaceClusterHandlerInVpc: jsii.Bool(true),
		Prune:                    nil,
		SecretsEncryptionKey:     nil,
		DefaultCapacity:          jsii.Number(float64(2)),
		DefaultCapacityInstance:  awsec2.NewInstanceType(jsii.String("t3.small")),
		DefaultCapacityType:      awseks.DefaultCapacityType_EC2,
	})

	return resource
}
func newEKSOpenIdConnectProvider(stack awscdk.Stack) /*awseks.OpenIdConnectProvider*/ {
	//return awseks.NewOpenIdConnectProvider(stack, jsii.String("OpenIdProvider"), &awseks.OpenIdConnectProviderProps{
	//	Url: resource.ClusterOpenIdConnectIssuerUrl(),
	//})
}

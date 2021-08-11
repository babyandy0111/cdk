package s3

import (
	"fmt"
	"github.com/andy-demo/gocdk/libs/stack_helper"
	"github.com/andy-demo/gocdk/resources/route53"
	"github.com/aws/aws-cdk-go/awscdk/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/awss3"
	//"github.com/andy-demo/gocdk/resources/route53"
	"github.com/aws/aws-cdk-go/awscdk"
	//"github.com/aws/aws-cdk-go/awscdk/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/awscloudfront"
	//"github.com/aws/aws-cdk-go/awscdk/awscloudfrontorigins"
	//"github.com/aws/aws-cdk-go/awscdk/awss3"
	"github.com/aws/jsii-runtime-go"
	"os"
)

type S3Stack struct {
	Stack awscdk.Stack
}

func New(parentStack awscdk.Stack, stackName *string, props *awscdk.StackProps) *S3Stack {
	stack := awscdk.NewStack(parentStack, stackName, props)
	return &S3Stack{Stack: stack}
}

func (s *S3Stack) CreateStorageBucket(certificate awscertificatemanager.ICertificate) {
	bucketName := stack_helper.GenerateNameForResource("upload") + "." + os.Getenv("ACM_MAIN_DOMAIN")
	// 建立OAI
	oai := awscloudfront.NewOriginAccessIdentity(s.Stack, jsii.String("cf-oai"), &awscloudfront.OriginAccessIdentityProps{
		Comment: jsii.String("For Cloudfront usage"),
	})
	f, err := os.Open("./cfkey.pub")
	defer f.Close()
	if err != nil {
		panic(err)
	}
	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}
	keycontent := make([]byte, stat.Size())
	f.Read(keycontent)
	pk := awscloudfront.NewPublicKey(s.Stack, jsii.String("cf-public-key"), &awscloudfront.PublicKeyProps{
		EncodedKey:    jsii.String(string(keycontent)),
		Comment:       jsii.String(stack_helper.GetEnv() + " public key"),
		PublicKeyName: jsii.String(stack_helper.GetEnv()),
	})
	// 建立 keygroup
	kg := awscloudfront.NewKeyGroup(s.Stack, jsii.String("cf-key-group"), &awscloudfront.KeyGroupProps{
		Items: &[]awscloudfront.IPublicKey{
			pk,
		},
		Comment:      jsii.String(fmt.Sprintf("%s key group for cloudfront", stack_helper.GetEnv())),
		KeyGroupName: jsii.String(stack_helper.GetEnv()),
	})
	// 建立 s3
	bucket := awss3.NewBucket(s.Stack, jsii.String("upload-bucket"), &awss3.BucketProps{
		AccessControl:     awss3.BucketAccessControl_PRIVATE,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		BucketName:        jsii.String(bucketName),
		Versioned:         jsii.Bool(false),
	})
	bucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: &[]*string{
			jsii.String("s3:PutObject"),
			jsii.String("s3:GetObject"),
			jsii.String("s3:DeleteObject"),
		},
		Effect: awsiam.Effect_ALLOW,
		Principals: &[]awsiam.IPrincipal{
			awsiam.NewAccountPrincipal(os.Getenv("ACCOUNT_ID")),
			oai.GrantPrincipal(),
		},
		Resources: &[]*string{
			bucket.ArnForObjects(jsii.String("*")),
		},
		Sid: jsii.String("DefaultUploadBucketPolicy"),
	}))
	//policy := awss3.NewBucketPolicy(s.Stack, jsii.String(""), &awss3.BucketPolicyProps{
	//	Bucket:        bucket,
	//	RemovalPolicy: "",
	//})
	// 建立 cloudfront
	var subdomainName = stack_helper.GenerateNameForResource("upload")
	if stack_helper.GetEnv() == "production" {
		subdomainName = "upload"
	}
	var defaultDomain = jsii.String(fmt.Sprintf("%s.%s", subdomainName, os.Getenv("ACM_MAIN_DOMAIN")))
	cf := awscloudfront.NewDistribution(s.Stack, jsii.String("cf-distribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			AllowedMethods: awscloudfront.AllowedMethods_ALLOW_ALL(),
			CachedMethods:  awscloudfront.CachedMethods_CACHE_GET_HEAD(),
			Compress:       jsii.Bool(true),
			TrustedKeyGroups: &[]awscloudfront.IKeyGroup{
				kg,
			},
			Origin: awscloudfrontorigins.NewS3Origin(bucket, &awscloudfrontorigins.S3OriginProps{
				OriginAccessIdentity: oai,
			}),
		},
		Certificate:            certificate,
		Comment:                jsii.String("General CF Domain"),
		DefaultRootObject:      jsii.String("/"),
		DomainNames:            &[]*string{defaultDomain},
		Enabled:                jsii.Bool(true),
		EnableIpv6:             jsii.Bool(true),
		EnableLogging:          jsii.Bool(false),
		HttpVersion:            awscloudfront.HttpVersion_HTTP2,
		MinimumProtocolVersion: awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2019,
		PriceClass:             awscloudfront.PriceClass_PRICE_CLASS_100,
	})
	//  指定 domain
	hostedzone := route53.GetHostZoneByDomainName(s.Stack, jsii.String("find-cf-domain"), os.Getenv("ACM_MAIN_DOMAIN"))
	awsroute53.NewARecord(s.Stack, jsii.String("route53-cf-domain"), &awsroute53.ARecordProps{
		Zone:       hostedzone,
		Comment:    jsii.String(stack_helper.GetEnv() + " upload domain"),
		RecordName: jsii.String(subdomainName),
		Ttl:        awscdk.Duration_Seconds(jsii.Number(300)),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewCloudFrontTarget(cf)),
	})
}

package acm

import (
	"github.com/andy-demo/gocdk/libs/stack_helper"
	"github.com/andy-demo/gocdk/resources/route53"
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awscertificatemanager"
	"github.com/aws/jsii-runtime-go"
	"os"
	"strings"
)

func NewACM(parentStack awscdk.Stack, name *string, props *awscdk.StackProps) (awscdk.Stack, awscertificatemanager.Certificate) {
	stack := awscdk.NewStack(parentStack, name, props)
	otherDomains := strings.Split(os.Getenv("ACM_OTHER_DOMAIN"), ",")
	var inputOtherDomains = make([]*string, 0)
	for _, v := range otherDomains {
		inputOtherDomains = append(inputOtherDomains, jsii.String(v))
	}
	hostzone := route53.GetHostZoneByDomainName(stack, jsii.String(stack_helper.GenerateNameForResource("FindHostedZone")), os.Getenv("ACM_MAIN_DOMAIN"))
	resource := awscertificatemanager.NewCertificate(stack, jsii.String(stack_helper.GenerateNameForResource("NewACM")), &awscertificatemanager.CertificateProps{
		DomainName:              jsii.String(os.Getenv("ACM_MAIN_DOMAIN")),
		SubjectAlternativeNames: &inputOtherDomains,
		Validation:              awscertificatemanager.CertificateValidation_FromDns(hostzone),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ACM_CERTIFICATE_ARN"), &awscdk.CfnOutputProps{
		Value:       resource.CertificateArn(),
		Description: jsii.String("Default Certificate for Default Domain"),
		ExportName:  jsii.String("ACM:CERTIFICATE:ARN"),
	})
	return stack, resource
}

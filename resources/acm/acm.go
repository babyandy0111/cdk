package acm

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awscertificatemanager"
	"github.com/aws/jsii-runtime-go"
	"github.com/faryne/go-cdk-example/resources/route53"
)

func NewACM(parentStack awscdk.Stack, name *string, props *awscdk.StackProps) (awscdk.Stack, awscertificatemanager.Certificate) {

	stack := awscdk.NewStack(parentStack, name, props)

	hostzone := route53.GetHostZoneByDomainName(stack, jsii.String("XXX-Test-FindHostedZone"), "mydomain.xxx")
	resource := awscertificatemanager.NewCertificate(stack, jsii.String("XXX-Test-NewACM"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String("mydomain.xxx"),
		SubjectAlternativeNames: &[]*string{
			jsii.String("*.mydomain.xxx"),
			jsii.String("mydomain.xxx"),
			jsii.String("*.rpc.mydomain.xxx"),
			jsii.String("*.api.mydomain.xxx"),
		},
		Validation: awscertificatemanager.CertificateValidation_FromDns(hostzone),
	})
	return stack, resource
}

package route53

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsroute53"
	"github.com/aws/jsii-runtime-go"
)

/**
Get HostZone Info
*/
func getPublicHostZone(stack awscdk.Stack, name *string, domain string) awsroute53.IHostedZone {
	return awsroute53.PublicHostedZone_FromLookup(stack, name, &awsroute53.HostedZoneProviderProps{
		DomainName:  jsii.String(domain),
		PrivateZone: jsii.Bool(false),
	})
}

/**
Return Domain Resource By Domain Name
*/
func GetHostZoneByDomainName(stack awscdk.Stack, name *string, domain string) awsroute53.IHostedZone {
	return getPublicHostZone(stack, name, domain)
}

/**
Return Domain Resource ByHostedZone ID
*/
func GetHostZoneByHostZoneId(stack awscdk.Stack, name *string, hostedzoneId *string) awsroute53.IHostedZone {
	return awsroute53.HostedZone_FromHostedZoneId(stack, name, hostedzoneId)
}

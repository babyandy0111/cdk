package iam

import (
	"github.com/aws/aws-cdk-go/awscdk"
)

type IAMStack struct {
	Stack awscdk.Stack
}

func NewIAM(parentStack awscdk.Stack, stackName *string, props *awscdk.StackProps) *IAMStack {
	stack := awscdk.NewStack(parentStack, stackName, props)

	return &IAMStack{Stack: stack}
}

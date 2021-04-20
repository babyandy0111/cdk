package stack_helper

import (
	"github.com/aws/aws-cdk-go/awscdk"
)

type ResourceCallback func(*MyCDKStack, []awscdk.Resource) awscdk.Resource

type StackDependency struct {
	Resource awscdk.Stack
	Reason   *string
}

type MyCDKStackInterface interface {
	// 初始化物件
	New(awscdk.Stack, *string, awscdk.StackProps) *MyCDKStack
	// 建立資源，並回傳出一個 CfnResource
	AddResource(ResourceCallback) awscdk.Resource
	// 增加 Stack Dependency
	AddStackDependency(stack *MyCDKStack, reason *string)
	// 增加輸出
	AddOutputResource(key string, resource awscdk.Resource)
	// 取得輸出
	GetOutputResource(key string) awscdk.Resource
}

type MyCDKStack struct {
	Stack             awscdk.Stack
	OutputResources   map[string]*string
	StackDependencies map[string]*MyCDKStack
	MyCDKStackInterface
}

func New(parentStack awscdk.Stack, stackName *string, props *awscdk.StackProps) *MyCDKStack {
	return &MyCDKStack{
		Stack:           awscdk.NewStack(parentStack, stackName, props),
		OutputResources: make(map[string]*string, 0),
	}
}

func (s *MyCDKStack) GetStack() awscdk.Stack {
	return s.Stack
}

func (s *MyCDKStack) AddResource(dependencies []awscdk.Resource, callback ResourceCallback) awscdk.Resource {
	resource := callback(s, dependencies)
	return resource
}

func (s *MyCDKStack) AddStackDependency(key string, stack *MyCDKStack, reason *string) {
	target := stack.Stack
	s.StackDependencies[key] = stack
	s.Stack.AddDependency(target, reason)
}

func (s *MyCDKStack) GetStackDependency(key string) *MyCDKStack {
	return s.StackDependencies[key]
}

func (s *MyCDKStack) AddOutputResource(key string, resource *string) {
	s.OutputResources[key] = resource
}

func (s *MyCDKStack) GetOutputResource(key string) *string {
	return s.OutputResources[key]
}

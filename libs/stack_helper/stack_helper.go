package stack_helper

import "github.com/aws/aws-cdk-go/awscdk"

type ResourceCallback func(*MyCDKStack, []awscdk.CfnResource) awscdk.CfnResource

type StackDependency struct {
	Resource awscdk.CfnStack
	Reason   *string
}

type MyCDKStackInterface interface {
	// 初始化物件
	New(awscdk.Stack, *string, awscdk.CfnStackProps) *MyCDKStack
	// 建立資源，並回傳出一個 CfnResource
	AddResource(ResourceCallback) awscdk.CfnResource
	// 增加 Stack Dependency
	AddStackDependency(stack *MyCDKStack, reason *string)
	// 增加 Stack 輸出
	AddOut(key *string, value *string) *string
	// 取得 Stack 輸出
	GetOut(key *string) *string
	// 增加
	AddOutputResource(key *string, resource awscdk.CfnResource)
	// 取得
	GetOutputResource(key *string) awscdk.CfnResource
	// 設定條件式
	SetCondition()
	// 取得條件式
	GetCondition() bool
}

type MyCDKStack struct {
	Stack             awscdk.CfnStack
	Outputs           map[*string]*string
	OutputResources   map[*string]awscdk.CfnResource
	Condition         bool
	StackDependencies map[*string]*MyCDKStack
	MyCDKStackInterface
}

func New(parentStack awscdk.CfnStack, stackName *string, props *awscdk.CfnStackProps) *MyCDKStack {
	return &MyCDKStack{
		Stack:           awscdk.NewCfnStack(parentStack, stackName, props),
		Outputs:         make(map[*string]*string, 0),
		OutputResources: make(map[*string]awscdk.CfnResource, 0),
	}
}

func (s *MyCDKStack) GetStack() awscdk.CfnStack {
	return s.Stack
}

func (s *MyCDKStack) AddResource(dependencies []awscdk.CfnResource, callback ResourceCallback) awscdk.CfnResource {
	resource := callback(s, dependencies)
	s.AddOutputResource(resource.LogicalId(), resource)
	return resource
}

func (s *MyCDKStack) AddStackDependency(stack *MyCDKStack, reason *string) {
	target := stack.Stack.Stack()
	s.StackDependencies[target.StackId()] = stack
	s.Stack.Stack().AddDependency(target, reason)
}

func (s *MyCDKStack) GetStackDependency(key *string) *MyCDKStack {
	return s.StackDependencies[key]
}

func (s *MyCDKStack) AddOut(key *string, value *string) {
	s.Outputs[key] = value
}

func (s *MyCDKStack) GetOut(key *string) *string {
	return s.Outputs[key]
}

func (s *MyCDKStack) AddOutputResource(key *string, resource awscdk.CfnResource) {
	s.OutputResources[key] = resource
}

func (s *MyCDKStack) GetOutputResource(key *string) awscdk.CfnResource {
	return s.OutputResources[key]
}

func (s *MyCDKStack) SetCondition(value bool) {
	s.Condition = value
}
func (s *MyCDKStack) GetCondition() bool {
	return s.Condition
}

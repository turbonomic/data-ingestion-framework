package probe

import (
	"fmt"

	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type (
	TestProbe                   struct{}
	TestProbeDiscoveryClient    struct{}
	TestProbeRegistrationClient struct{}
	TestProbeActionClient       struct{}
)

func (handler *TestProbeDiscoveryClient) GetAccountValues() *TurboTargetInfo {
	return nil
}

func (handler *TestProbeDiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	return nil, fmt.Errorf("TestProbeDiscoveryClient Validate not implemented")
}

func (handler *TestProbeDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	return nil, fmt.Errorf("TestProbeDiscoveryClient Discover not implemented")
}

func (registrationClient *TestProbeRegistrationClient) GetSupplyChainDefinition() []*proto.TemplateDTO {
	return nil
}

func (registrationClient *TestProbeRegistrationClient) GetAccountDefinition() []*proto.AccountDefEntry {
	return nil
}

func (registrationClient *TestProbeRegistrationClient) GetIdentifyingFields() string {
	return ""
}

func (actionClient *TestProbeActionClient) ExecuteAction(actionExecutionDTO *proto.ActionExecutionDTO,
	accountValues []*proto.AccountValue,
	progressTracker ActionProgressTracker,
) (*proto.ActionResult, error) {
	return nil, fmt.Errorf("TestProbeDiscoveryClient ExecuteAction not implemented")
}

func (actionClient *TestProbeActionClient) ExecuteActionList(actionExecutionDTOs []*proto.ActionExecutionDTO,
	accountValues []*proto.AccountValue,
	progressTracker ActionListProgressTracker,
) (*proto.ActionListResponse, error) {
	return nil, fmt.Errorf("TestProbeDiscoveryClient ExecuteActionList not implemented")
}

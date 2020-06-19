package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"os"
)

func EntityType(difType data.DIFEntityType) *proto.EntityDTO_EntityType {
	entityTypeStr := data.DIFEntityTypeToTemplateEntityStringMap[difType]
	eType, exists := registration.TemplateEntityTypeMap[entityTypeStr]
	if !exists {
		return nil
	}

	return &eType
}

func getEntityId(entityType proto.EntityDTO_EntityType, entityName, scope string) string {
	eType := proto.EntityDTO_EntityType_name[int32(entityType)]

	return fmt.Sprintf("%s-%s-%s", eType, entityName, scope)
}

func getEntityPropertyNameValue(name, value string) *proto.EntityDTO_EntityProperty {
	attr := name
	ns := registration.DefaultPropertyNamespace

	return &proto.EntityDTO_EntityProperty{
		Namespace: &ns,
		Name:      &attr,
		Value:     &value,
	}
}

func logDebug(msg ...interface{}) {
	if os.Getenv("TURBODIF_LOCAL_DEBUG") == "1" && glog.V(4) {
		glog.Infof("%+v", msg)
	}
}

func logSupplyChainDetails(supplyChainNode *registration.SupplyChainNode) {
	if os.Getenv("TURBODIF_LOCAL_DEBUG") == "1" && glog.V(4) {

		var expectedSoldComms []string
		for comm := range supplyChainNode.SupportedComms {
			expectedSoldComms = append(expectedSoldComms, fmt.Sprintf("%v", comm))
		}
		glog.V(4).Infof("expectedSoldComms: %v", expectedSoldComms)

		var expectedSoldAccessComms []string
		for comm := range supplyChainNode.SupportedAccessComms {
			expectedSoldAccessComms = append(expectedSoldAccessComms, fmt.Sprintf("%v", comm))
		}
		glog.V(4).Infof("expectedSoldAccessComms: %v", expectedSoldAccessComms)

		expectedBought := make(map[string][]string)
		for provider, bought := range supplyChainNode.SupportedBoughtComms {
			var comms []string
			for comm := range bought {
				comms = append(comms, fmt.Sprintf("%v", comm))
			}
			expectedBought[fmt.Sprintf("%v", provider)] = comms
		}
		glog.V(4).Infof("expectedBought: %v", expectedBought)

		expectedAccessBought := make(map[string][]string)
		for provider, bought := range supplyChainNode.SupportedBoughtAccessComms {
			var comms []string
			for comm := range bought {
				comms = append(comms, fmt.Sprintf("%v", comm))
			}
			expectedAccessBought[fmt.Sprintf("%v", provider)] = comms
		}
		glog.V(4).Infof("expectedAccessBought: %v", expectedAccessBought)

		hostedByProviderType := supplyChainNode.ProviderByProviderType
		expectedHostedByProviderType := make(map[string]string)
		for provider, hostingType := range hostedByProviderType {
			expectedHostedByProviderType[fmt.Sprintf("%v", provider)] = hostingType
		}
		glog.V(4).Infof("expectedHostedByProviderType: %v", expectedHostedByProviderType)
	}
}

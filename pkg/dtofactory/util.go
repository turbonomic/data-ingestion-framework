package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
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

func createCommodityWithKey(accessCommType proto.CommodityDTO_CommodityType, key string) *proto.CommodityDTO {
	appCommodity, _ := builder.NewCommodityDTOBuilder(accessCommType).Key(key).Create()
	return appCommodity
}

func createCommodity(accessCommType proto.CommodityDTO_CommodityType) *proto.CommodityDTO {
	appCommodity, _ := builder.NewCommodityDTOBuilder(accessCommType).Create()
	return appCommodity
}

func logDebug(f func(format string, a ...interface{}) (int, error), msg ...interface{}) {
	if os.Getenv("TURBODIF_LOCAL_DEBUG") == "1" && glog.V(4) {
		f("%++v\n", msg)
	}
}

func logSupplyChainDetails(supplyChainNode *registration.SupplyChainNode) {
	if os.Getenv("TURBODIF_LOCAL_DEBUG") == "1" && glog.V(4) {

		var expectedSoldComms []string
		for comm, _ := range supplyChainNode.SupportedComms {
			expectedSoldComms = append(expectedSoldComms, fmt.Sprintf("%v", comm))
		}
		fmt.Printf("expectedSoldComms: %v\n", expectedSoldComms)

		var expectedSoldAccessComms []string
		for comm, _ := range supplyChainNode.SupportedAccessComms {
			expectedSoldAccessComms = append(expectedSoldAccessComms, fmt.Sprintf("%v", comm))
		}
		fmt.Printf("expectedSoldAccessComms: %v\n", expectedSoldAccessComms)

		expectedBought := make(map[string][]string)
		for provider, bought := range supplyChainNode.SupportedBoughtComms {
			var comms []string
			for comm, _ := range bought {
				comms = append(comms, fmt.Sprintf("%v", comm))
			}
			expectedBought[fmt.Sprintf("%v", provider)] = comms
		}
		fmt.Printf("expectedBought: %v\n", expectedBought)

		expectedAccessBought := make(map[string][]string)
		for provider, bought := range supplyChainNode.SupportedBoughtAccessComms {
			var comms []string
			for comm, _ := range bought {
				comms = append(comms, fmt.Sprintf("%v", comm))
			}
			expectedAccessBought[fmt.Sprintf("%v", provider)] = comms
		}
		fmt.Printf("expectedAccessBought: %v\n", expectedAccessBought)

		hostedByProviderType := supplyChainNode.ProviderByProviderType
		expectedHostedByProviderType := make(map[string]string)
		for provider, hostingType := range hostedByProviderType {
			expectedHostedByProviderType[fmt.Sprintf("%v", provider)] = hostingType
		}
		fmt.Printf("expectedHostedByProviderType: %v\n", expectedHostedByProviderType)
	}
}

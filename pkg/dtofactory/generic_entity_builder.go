package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type GenericEntityBuilder struct {
	entityType      proto.EntityDTO_EntityType
	difEntity       *data.BasicDIFEntity
	scope           string
	keepStandalone  bool
	supplyChainNode *registration.SupplyChainNode
}

func NewGenericEntityBuilder(entityType proto.EntityDTO_EntityType,
	difEntity *data.BasicDIFEntity, scope string, keepStandalone bool,
	supplyChainNode *registration.SupplyChainNode) *GenericEntityBuilder {
	return &GenericEntityBuilder{
		entityType:      entityType,
		keepStandalone:  keepStandalone,
		difEntity:       difEntity,
		scope:           scope,
		supplyChainNode: supplyChainNode,
	}
}

func (eb *GenericEntityBuilder) BuildEntity() (*proto.EntityDTO, error) {
	var dto *proto.EntityDTO

	id := getEntityId(eb.entityType, eb.difEntity.EntityId, eb.scope)
	glog.V(3).Infof("*** building ... %s", id)

	entityBuilder := builder.
		NewEntityDTOBuilder(eb.entityType, id).
		DisplayName(eb.difEntity.EntityId)

	mergePropertiesMap := make(map[string]string)
	commoditiesMap := make(map[proto.CommodityDTO_CommodityType][]*builder.CommodityDTOBuilder) //[]*proto.CommodityDTO)

	for _, difEntity := range eb.difEntity.GetDIFEntities() {
		// Entity Properties from matching identifiers
		if difEntity.MatchingIdentifiers != nil {
			matchingIdentifiers := difEntity.MatchingIdentifiers
			// currently we only have IP address as matching identifiers for merging with real entity in the server
			ip := matchingIdentifiers.IPAddress
			// saving ip address as key to avoid duplicates,
			//we will create a comma separated list before setting the entity properyy
			mergePropertiesMap[ip] = IPAttr

			// mark entity with MatchingId as PROXY  - TODO: need to change turbo-go-sdk to allow marking entity as proxy
			// without needing to set ReplacementMetadata which is not used in XL
			replacementEntityMetaDataBuilder := builder.NewReplacementEntityMetaDataBuilder()
			metaData := replacementEntityMetaDataBuilder.Matching(IPAttr).Build()
			entityBuilder.ReplacedBy(metaData)
		}
		// Build commodities corresponding the JSON metrics data
		cb := NewGenericCommodityBuilder(difEntity)
		commMap, _ := cb.BuildCommodity()

		// consolidate the metrics from this dif entity to the main commodities map
		for commType, commList := range commMap {
			commoditiesMap[commType] = commList
		}
	}
	logDebug(fmt.Printf, commoditiesMap)

	// Setting commodities and properties in the entity builder using the supply chain template
	supplyChainNode := eb.supplyChainNode
	logSupplyChainDetails(supplyChainNode)

	// Select sold commodities
	soldCommodities := eb.soldCommodities(commoditiesMap)
	entityBuilder.SellsCommodities(soldCommodities)

	// Bought commodities
	for pType, providerIds := range eb.difEntity.GetProviders() {

		// select commodities bought from the provider
		providerType, boughtCommodities := eb.boughtCommodities(pType, commoditiesMap)
		if providerType == nil {
			glog.Errorf("Invalid provider entity type %s", pType)
			continue
		}

		// Adding the provider and associated bought commodities to the entity builder
		for _, pId := range providerIds {
			providerId := getEntityId(*providerType, pId, eb.scope)
			glog.V(3).Infof("%s --> adding internal provider %s", id, providerId)
			entityBuilder.
				Provider(builder.CreateProvider(*providerType, providerId)).
				BuysCommodities(boughtCommodities)
		}
	}

	// External providers, commodities and metadata specified using IP or UUID
	scHostedByProviderType := supplyChainNode.ProviderByProviderType //map of external provider type and hosting relationship
	if eb.difEntity.GetExternalProviderByIP() != nil {
		// All external providers for this entity - set as provider in the DTO
		for pType, pMap := range eb.difEntity.GetExternalProviderByIP() {
			// select commodities bought from the provider
			//providerType, boughtCommodities := eb.boughtCommodities(pType, commoditiesMap)
			providerType, boughtCommodities := eb.externalBoughtCommodities(pType, commoditiesMap)

			if providerType == nil {
				glog.Errorf("Invalid hostedBy provider entity type %s", pType)
				continue
			}

			// Add the provider and associated bought commodities to the entity builder
			// Provider entity will be created by the proxy_provider_builder
			var providerIds []string
			for pId, _ := range pMap {
				providerIds = append(providerIds, pId)
				providerId := getEntityId(*providerType, pId, eb.scope)
				glog.V(3).Infof("%s --> adding external provider %s::%s", id, pType, providerId)
				entityBuilder.
					Provider(builder.CreateProvider(*providerType, providerId)).
					BuysCommodities(boughtCommodities)
			}

			if scHostedByProviderType[*providerType] == "HOSTING" && len(pMap) > 1 {
				// There should only be one of the hosting provider
				glog.Errorf("%s::%s Invalid number of external hostedBy providers %v",
					eb.entityType, eb.difEntity.EntityId, providerIds)
			}
		}
	}

	if eb.difEntity.GetExternalProviderByUUID() != nil {
		// All external providers for this entity - set as provider in the DTO
		for pType, pMap := range eb.difEntity.GetExternalProviderByUUID() {
			// select commodities bought from the provider
			providerType, boughtCommodities := eb.externalBoughtCommodities(pType, commoditiesMap)

			if providerType == nil {
				glog.Errorf("Invalid hostedBy provider entity type %s", pType)
				continue
			}

			// Add the provider and associated bought commodities to the entity builder
			var providerIds []string
			for pId, _ := range pMap {
				providerIds = append(providerIds, pId)
				providerId := getEntityId(*providerType, pId, eb.scope)
				glog.Infof("%s --> adding external provider %s::%s", id, pType, providerId)
				entityBuilder.
					Provider(builder.CreateProvider(*providerType, providerId)).
					BuysCommodities(boughtCommodities)
			}

			// Provider entity will be created by the proxy_provider_builder
			if scHostedByProviderType[*providerType] == "HOSTING" && len(pMap) > 1 {
				// There should only be one of the hosting provider
				glog.Errorf("%s::%s Invalid number of external hostedBy providers %v",
					eb.entityType, eb.difEntity.EntityId, providerIds)
			}
		}
	}

	// Adding merging metadata properties to the entity builder
	// create comma separated list if there are multiple values for the same property
	propMap := make(map[string]string)
	for propVal, propName := range mergePropertiesMap {
		if prop, exists := propMap[propName]; exists {
			prop = fmt.Sprintf(prop + "," + propVal)
			propMap[propName] = prop
		} else {
			propMap[propName] = propVal
		}
	}

	for propName, propVal := range propMap {
		entityBuilder.WithProperty(getEntityPropertyNameValue(propName, propVal))
	}

	dto, err := entityBuilder.Create()

	if err != nil {
		return nil, err
	}

	if eb.entityType == proto.EntityDTO_VIRTUAL_MACHINE {
		glog.Infof("NODE %++v\n", dto)
	}
	logDebug(fmt.Printf, protobuf.MarshalTextString(dto))
	return dto, nil
}

func (eb *GenericEntityBuilder) soldCommodities(
	commoditiesMap map[proto.CommodityDTO_CommodityType][]*builder.CommodityDTOBuilder) []*proto.CommodityDTO {
	var soldCommodities []*proto.CommodityDTO

	// SOLD COMM CONFIG
	scSupportedComms := eb.supplyChainNode.SupportedComms // map of associated comms
	scSupportedAccessComms := eb.supplyChainNode.SupportedAccessComms

	kb := NewCommodityKeyBuilder(eb.entityType, eb.difEntity)
	soldCommKey := kb.GetKey()

	for commType, commList := range commoditiesMap {
		_, ok := scSupportedComms[commType] // is the commodity type supported by the supply chain
		if !ok {                            //do no include commodity not specified in the supply chain
			glog.Warningf("%s:%s : unsupported sold commodity type %v",
				eb.entityType, eb.difEntity.EntityId, commType)
			continue
		}
		commTemplate := scSupportedComms[commType] //commodity template
		for _, cb := range commList {
			soldComm, _ := cb.Create()   //nothing to fail, so ignore the error
			if commTemplate.Key != nil { //commodity needs  key
				if soldComm.Key != nil {
					glog.V(3).Infof("Commodity Key is available in the json file : %++v\n", soldComm)
				} else if soldCommKey != nil {
					soldComm.Key = soldCommKey
				}
			}

			soldCommodities = append(soldCommodities, soldComm)
		}
	}

	// create access sold commodities
	accessCommKey := ""
	if soldCommKey != nil {
		accessCommKey = *soldCommKey
	}
	for commType, _ := range scSupportedAccessComms {
		soldCommodities = append(soldCommodities, createCommodityWithKey(commType, accessCommKey))
	}

	return soldCommodities
}

// Select the commodities from the metrics in the json file as commodities bought from the given provider.
// Commodity types are selected based on the supply chain specification for the entity type
func (eb *GenericEntityBuilder) boughtCommodities(pType data.DIFEntityType,
	commoditiesMap map[proto.CommodityDTO_CommodityType][]*builder.CommodityDTOBuilder,
) (*proto.EntityDTO_EntityType, []*proto.CommodityDTO) {

	var providerType proto.EntityDTO_EntityType
	var boughtCommodities []*proto.CommodityDTO

	// provider type
	eType := EntityType(pType)
	if eType == nil {
		return nil, boughtCommodities
	}
	providerType = *eType

	// supply chain specification for the entity providers and bought commodities
	scSupportedBoughtComms := eb.supplyChainNode.SupportedBoughtComms             // map of provider type and associated  commodities map
	scSupportedBoughtAccessComms := eb.supplyChainNode.SupportedBoughtAccessComms //map of provider type and associated access commodities map

	if _, exists := scSupportedBoughtComms[providerType]; !exists {
		glog.Errorf("Supply chain does not support provider %s for entity %s", providerType, eb.entityType)
		return &providerType, boughtCommodities
	}

	scProviderComms := scSupportedBoughtComms[providerType]
	scProviderAccessComms := scSupportedBoughtAccessComms[providerType]

	kb := NewCommodityKeyBuilder(eb.entityType, eb.difEntity)
	boughtCommKey := kb.GetKey()

	// Select commodities bought from the provider from the commodities map
	for commType, commList := range commoditiesMap {
		_, ok := scProviderComms[commType]
		if !ok { //dp not include commodity not specified in the supply chain?
			glog.Warningf("%s::%s: unsupported bought commodity type %v",
				eb.entityType, eb.difEntity.EntityId, commType)
			continue
		}
		commTemplate := scProviderComms[commType]
		for _, cb := range commList {
			boughtComm, _ := cb.Create() //nothing to fail, so ignore the error
			if commTemplate.Key != nil { //commodity needs  key
				if boughtComm.Key != nil {
					glog.V(3).Infof("Commodity Key is available in the json file : %++v\n", boughtComm)
				} else if boughtComm != nil {
					boughtComm.Key = boughtCommKey
				}
			}
			boughtCommodities = append(boughtCommodities, boughtComm)
		}
	}
	// create access bought commodities
	accessCommKey := ""
	if boughtCommKey != nil {
		accessCommKey = *boughtCommKey
	}
	for commType, _ := range scProviderAccessComms {
		boughtCommodities = append(boughtCommodities, createCommodityWithKey(commType, accessCommKey))
	}

	return &providerType, boughtCommodities
}

// Select the commodities from the metrics in the json file as commodities bought from the given external provider.
// Commodity types are selected based on the commodities bought section in the supply chain specification for the entity type
func (eb *GenericEntityBuilder) externalBoughtCommodities(pType data.DIFEntityType,
	commoditiesMap map[proto.CommodityDTO_CommodityType][]*builder.CommodityDTOBuilder,
) (*proto.EntityDTO_EntityType, []*proto.CommodityDTO) {
	var providerType proto.EntityDTO_EntityType
	var boughtCommodities []*proto.CommodityDTO

	// provider type
	eType := EntityType(pType)
	if eType == nil {
		return nil, boughtCommodities
	}
	providerType = *eType

	scHostedByBoughtComms := eb.supplyChainNode.SupportedBoughtComms
	// Get the commodities that should be created as per the supply chain for this proxy provider
	if _, exists := scHostedByBoughtComms[providerType]; !exists {
		glog.Errorf("Supply chain does not support hostedBy provider %s for entity %s",
			providerType, eb.entityType)
		return &providerType, boughtCommodities
	}
	// create the commodities bought from the external provider
	externalProviderComms := scHostedByBoughtComms[providerType]

	for commType, _ := range externalProviderComms {
		if commList, exists := commoditiesMap[commType]; exists {
			for _, cb := range commList {
				boughtComm, _ := cb.Create() //nothing to fail, so ignore the error
				boughtCommodities = append(boughtCommodities, boughtComm)
			}
		} else {
			glog.V(3).Infof("Creating fake bought commodity %v for provider %v", commType, providerType)
			boughtCommodities = append(boughtCommodities, createCommodity(commType))
		}
	}
	return &providerType, boughtCommodities
}

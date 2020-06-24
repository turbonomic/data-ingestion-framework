package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	sdkdata "github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
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

	entityID := getEntityId(eb.entityType, eb.difEntity.EntityId, eb.scope)
	glog.V(3).Infof("Building %s", entityID)

	entityBuilder := builder.
		NewEntityDTOBuilder(eb.entityType, entityID).
		DisplayName(eb.difEntity.DisplayName)

	mergePropertiesMap := make(map[string]string)
	commoditiesMap := make(map[proto.CommodityDTO_CommodityType][]*builder.CommodityDTOBuilder) //[]*proto.CommodityDTO)
	// Consolidate metrics from multiple dif entities into the main commodities map
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
			commoditiesMap[commType] = append(commoditiesMap[commType], commList...)
		}
	}
	logDebug(commoditiesMap)

	supplyChainNode := eb.supplyChainNode
	logSupplyChainDetails(supplyChainNode)

	// Setting commodities and properties in the entity builder using the supply chain template
	// Select sold commodities
	soldCommodities := eb.soldCommodities(commoditiesMap)
	entityBuilder.SellsCommodities(soldCommodities)
	// Select bought commodities from internal providers
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
			glog.V(3).Infof("%s --> adding internal provider %s", entityID, providerId)
			entityBuilder.
				Provider(builder.CreateProvider(*providerType, providerId)).
				BuysCommodities(boughtCommodities)
		}
	}
	// Select bought commodities from external providers
	eb.externalProviders(supplyChainNode, eb.difEntity.GetExternalProviderByIP(), entityID, entityBuilder, commoditiesMap)
	eb.externalProviders(supplyChainNode, eb.difEntity.GetExternalProviderByUUID(), entityID, entityBuilder, commoditiesMap)

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

	logDebug(protobuf.MarshalTextString(dto))
	return dto, nil
}

func (eb *GenericEntityBuilder) externalProviders(supplyChainNode *registration.SupplyChainNode,
	providers map[data.DIFEntityType]map[string][]*sdkdata.DIFEntity,
	entityID string, entityBuilder *builder.EntityDTOBuilder, commoditiesMap CommoditiesByType) {
	// Map of external provider type and hosting relationship
	scHostedByProviderType := supplyChainNode.ProviderByProviderType
	for pType, pMap := range providers {
		// Validate that the provider type is defined in the supply chain template
		eType := EntityType(pType)
		if eType == nil {
			glog.Errorf("%s::%s: Invalid hostedBy provider entity type %s",
				eb.entityType, eb.difEntity.EntityId, pType)
			continue
		}
		// Construct the IDs of the external providers
		var providerIds []string
		for pId := range pMap {
			providerIds = append(providerIds, getEntityId(*eType, pId, eb.scope))
		}
		// Make sure for HOSTING provider type, there is no more than 1 providers
		if scHostedByProviderType[*eType] == "HOSTING" && len(pMap) > 1 {
			glog.Errorf("%s::%s: Number of external hostedBy providers %v is larger than 1",
				eb.entityType, eb.difEntity.EntityId, providerIds)
			continue
		}
		// Select commodities bought from the external providers
		boughtCommodities := eb.externalBoughtCommodities(*eType, commoditiesMap)
		// Add the provider and associated bought commodities to the entity builder
		// Provider entity will be created by the proxy_provider_builder
		for _, providerId := range providerIds {
			glog.V(3).Infof("%s --> adding external provider %s::%s", entityID, pType, providerId)
			entityBuilder.
				Provider(builder.CreateProvider(*eType, providerId)).
				BuysCommodities(boughtCommodities)
		}
	}
}

func (eb *GenericEntityBuilder) soldCommodities(commoditiesMap CommoditiesByType) []*proto.CommodityDTO {
	var soldCommodities []*proto.CommodityDTO

	// SOLD COMM CONFIG
	scSupportedComms := eb.supplyChainNode.SupportedComms // map of associated comms
	scSupportedAccessComms := eb.supplyChainNode.SupportedAccessComms

	// Set resize of commodities
	setResizable(eb.entityType, commoditiesMap)

	for commType, commList := range commoditiesMap {
		_, ok := scSupportedComms[commType] // is the commodity type supported by the supply chain
		if !ok {                            // do not include commodity not specified in the supply chain
			glog.V(4).Infof("%s:%s does not sell %v",
				eb.entityType, eb.difEntity.EntityId, commType)
			continue
		}
		commTemplate := scSupportedComms[commType] //commodity template
		for _, cb := range commList {
			soldComm, err := cb.Create() //nothing to fail, so ignore the error
			if err != nil {
				glog.Warningf("Failed to create sold commodity %v: %v", commType, err)
				continue
			}
			if commTemplate.Key != nil && soldComm.Key == nil {
				glog.Warningf("Commodity key is required for %+v but not discovered in the JSON.", soldComm)
				continue
			} else if commTemplate.Key == nil {
				if soldComm.Key != nil {
					glog.V(4).Info("Commodity key is not defined in the template for %+v but is "+
						"discovered in the JSON. Ignore the key.", soldComm)
				}
				soldComm.Key = nil
			}
			soldCommodities = append(soldCommodities, soldComm)
		}
	}

	// create access sold commodities
	soldCommKeys := NewCommodityKeyBuilder(eb.entityType, eb.difEntity).GetSoldCommKey()
	for commType := range scSupportedAccessComms {
		for _, soldCommKey := range soldCommKeys {
			soldCommodity, err := builder.NewCommodityDTOBuilder(commType).Key(soldCommKey).Create()
			if err != nil {
				glog.Errorf("Failed to create sold commodity %v with key %v: %v", commType, soldCommKey, err)
				continue
			}
			soldCommodities = append(soldCommodities, soldCommodity)
		}
	}

	return soldCommodities
}

// Select the commodities from the metrics in the json file as commodities bought from the given provider.
// Commodity types are selected based on the supply chain specification for the entity type
func (eb *GenericEntityBuilder) boughtCommodities(
	pType data.DIFEntityType, commoditiesMap CommoditiesByType) (*proto.EntityDTO_EntityType, []*proto.CommodityDTO) {

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

	// Select commodities bought from the provider from the commodities map
	for commType, commList := range commoditiesMap {
		_, ok := scProviderComms[commType]
		if !ok {
			// Do not include commodity not specified in the supply chain
			glog.V(4).Infof("%s::%s: unsupported bought commodity type %v",
				eb.entityType, eb.difEntity.EntityId, commType)
			continue
		}
		commTemplate := scProviderComms[commType]
		for _, cb := range commList {
			boughtComm, err := cb.Create()
			if err != nil {
				glog.Warningf("Failed to create bought commodity %v: %v", commType, err)
				continue
			}
			if commTemplate.Key != nil && boughtComm.Key == nil {
				glog.Warningf("Commodity key is required for %+v but not discovered in the JSON", boughtComm)
				continue
			}
			boughtCommodities = append(boughtCommodities, boughtComm)
		}
	}
	// Create access bought commodities
	boughtCommKeys := NewCommodityKeyBuilder(eb.entityType, eb.difEntity).GetBoughtCommKey(true)
	for commType := range scProviderAccessComms {
		for _, boughtCommKey := range boughtCommKeys {
			boughtCommodity, err := builder.NewCommodityDTOBuilder(commType).Key(boughtCommKey).Create()
			if err != nil {
				glog.Errorf("Failed to create bought commodity %v with key %v: %v", commType, boughtCommKey, err)
				continue
			}
			boughtCommodities = append(boughtCommodities, boughtCommodity)
		}
	}
	return &providerType, boughtCommodities
}

// Select the commodities from the metrics in the json file as commodities bought from the given external provider.
// Commodity types are selected based on the commodities bought section in the supply chain specification for the entity type
func (eb *GenericEntityBuilder) externalBoughtCommodities(
	eType proto.EntityDTO_EntityType, commoditiesMap CommoditiesByType) []*proto.CommodityDTO {
	var boughtCommodities []*proto.CommodityDTO

	scHostedByBoughtComms := eb.supplyChainNode.SupportedBoughtComms
	scHostedByBoughtAccessComms := eb.supplyChainNode.SupportedBoughtAccessComms
	// Get the commodities that should be created as per the supply chain for this proxy provider
	if _, exists := scHostedByBoughtComms[eType]; !exists {
		glog.Errorf("Supply chain does not support hostedBy provider %s for entity %s",
			eType, eb.entityType)
		return boughtCommodities
	}
	// create the commodities bought from the external provider
	externalProviderComms := scHostedByBoughtComms[eType]
	externalProviderAccessComms := scHostedByBoughtAccessComms[eType]
	for commType := range externalProviderComms {
		if commList, exists := commoditiesMap[commType]; exists {
			for _, cb := range commList {
				boughtComm, err := cb.Create() //nothing to fail, so ignore the error
				if err != nil {
					continue
				}
				boughtCommodities = append(boughtCommodities, boughtComm)
			}
		}
	}

	// create access bought commodities
	boughtCommKeys := NewCommodityKeyBuilder(eb.entityType, eb.difEntity).SetScope(eb.scope).GetBoughtCommKey(false)
	for commType := range externalProviderAccessComms {
		for _, boughtCommKey := range boughtCommKeys {
			boughtCommodity, err := builder.NewCommodityDTOBuilder(commType).Key(boughtCommKey).Create()
			if err != nil {
				glog.Errorf("Failed to create access bought commodity %v with key %v: %v", commType, boughtCommKey, err)
				continue
			}
			boughtCommodities = append(boughtCommodities, boughtCommodity)
		}
	}

	return boughtCommodities
}

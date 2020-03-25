package conf

import (
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type SupplyChainConfig struct {
	Nodes []*NodeConfig `yaml:"supplyChainNode"`
}

type NodeConfig struct {
	TemplateClass        string                      `yaml:"templateClass"`
	TemplateType         *string                     `yaml:"templateType"`
	TemplatePriority     int32                       `yaml:"templatePriority"`
	CommodityBoughtList  []*CommodityBoughtConfig    `yaml:"commodityBought"`
	CommoditySoldList    []*CommodityConfig          `yaml:"commoditySold"`
	MergedEntityMetaData *MergedEntityMetaDataConfig `yaml:"mergedEntityMetaData"`
	ExternalLinkList     []*ExternalEntityLinkConfig `yaml:"externalLink"`
}

type CommodityBoughtConfig struct {
	Provider *ProviderConfig    `yaml:"key"`
	Comms    []*CommodityConfig `yaml:"value"`
}

type ProviderConfig struct {
	TemplateClass  *string `yaml:"templateClass"`
	ProviderType   *string `yaml:"providerType"`
	CardinalityMax int64   `yaml:"cardinalityMax"`
	CardinalityMin int64   `yaml:"cardinalityMin"`
}

type CommodityConfig struct {
	CommodityType *string `yaml:"commodityType"`
	Key           *string `yaml:"key"`
}

type ExternalEntityLinkConfig struct {
	Key   *string     `yaml:"key"`
	Value *LinkConfig `yaml:"value"`
}

type LinkConfig struct {
	BuyerRef                   *string                      `yaml:"buyerRef"`
	SellerRef                  *string                      `yaml:"sellerRef"`
	Relationship               *string                      `yaml:"relationship"`
	CommodityDefList           []*CommodityDef              `yaml:"commodityDefs"`
	ProbeEntityPropertyList    []*ProbeEntityPropertyDef    `yaml:"probeEntityPropertyDef"`
	ExternalEntityPropertyList []*ExternalEntityPropertyDef `yaml:"externalEntityPropertyDefs"`
}

type CommodityDef struct {
	CommType *string `yaml:"type"`
	HasKey   *bool   `yaml:"hasKey"`
}

type ProbeEntityPropertyDef struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type ExternalEntityPropertyDef struct {
	Entity      string                 `yaml:"entity"`
	Attribute   string                 `yaml:"attribute"`
	PropHandler *PropertyHandlerConfig `yaml:"propertyHandler"`
}

type PropertyHandlerConfig struct {
	MethodName    string `yaml:"methodName"`
	EntityType    string `yaml:"entityType"`
	DirectlyApply bool   `yaml:"directlyApply"`
}

type MergedEntityMetaDataConfig struct {
	KeepInTopology    bool                          `yaml:"keepStandalone"`
	MatchingMetadata  *MatchingMetadataConfig       `yaml:"matchingMetadata"`
	CommSold          []string                      `yaml:"commoditiesSold"`
	CommoditiesBought []*MergedMetadataBoughtConfig `yaml:"commoditiesBought"`
}

type MergedMetadataBoughtConfig struct {
	Provider string   `yaml:"providerType"`
	Comm     []string `yaml:"commodityMetadata"`
}

type MatchingMetadataConfig struct {
	ReturnType               string `yaml:"returnType"`
	ExternalEntityReturnType string `yaml:"externalEntityReturnType"`

	MatchingDataList                   []*MatchingDataConfig `yaml:"matchingData"`
	ExternalEntityMatchingPropertyList []*MatchingDataConfig `yaml:"externalEntityMatchingProperty"`
}
type MatchingDataConfig struct {
	MatchingProperty *MatchingPropertyConfig `yaml:"matchingProperty"`
	MatchingField    *MatchingFieldConfig    `yaml:"matchingField"`
	Delimiter        string                  `yaml:"delimiter"`
}

type MatchingPropertyConfig struct {
	PropertyName string `yaml:"propertyName"`
}

type MatchingFieldConfig struct {
	MessagePath []string `yaml:"messagePath"`
	FieldName   string   `yaml:"fieldName"`
}

// =======================================================================

func LoadSupplyChain(filename string) (*SupplyChainConfig, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		glog.Errorf("%++v\n", err)
		return nil, err
	}
	var supplyChain SupplyChainConfig
	err = yaml.Unmarshal(contents, &supplyChain)

	if err != nil {
		glog.Errorf("%++v\n", err)
		return nil, err
	}

	//if glog.V(3) {
	PrintSupplyChain(supplyChain)
	//}

	return &supplyChain, nil
}

func PrintSupplyChain(supplyChain SupplyChainConfig) {
	for _, node := range supplyChain.Nodes {
		glog.Infof("********* %s::%s::%d", node.TemplateClass, *node.TemplateType, node.TemplatePriority)
		if node.CommodityBoughtList != nil {
			for _, bought := range node.CommodityBoughtList {
				if bought.Provider != nil {
					if bought.Provider.TemplateClass != nil {
						glog.Infof("provider class: %s", *bought.Provider.TemplateClass)
					}
					if bought.Provider.ProviderType != nil {
						glog.Infof("provider type: %s", *bought.Provider.ProviderType)
					}
				}
				if bought.Comms != nil {
					for _, cb := range bought.Comms {
						if cb.CommodityType != nil && cb.Key != nil {
							glog.Infof("bought comm: %s::%s", *cb.CommodityType, *cb.Key)
						} else if cb.CommodityType != nil {
							glog.Infof("bought comm: %s", *cb.CommodityType)
						}
					}
				}
			}
		}

		if node.CommoditySoldList != nil {
			for _, sold := range node.CommoditySoldList {
				if sold.CommodityType != nil && sold.Key != nil {
					glog.Infof("sold comm: %s::%s", *sold.CommodityType, *sold.Key)
				} else if sold.CommodityType != nil {
					glog.Infof("sold comm: %s", *sold.CommodityType)
				}
			}
		}
		for _, externalLink := range node.ExternalLinkList {
			if externalLink.Value == nil {
				glog.Infof("Null external link")
				continue
			}
			link := externalLink.Value
			if link.SellerRef == nil {
				glog.Infof("external link: %s has null seller", *externalLink.Key)
			}
			if link.BuyerRef == nil {
				glog.Infof("external link: %s has null buyer", *externalLink.Key)
			}
			if link.SellerRef != nil && link.BuyerRef != nil {
				glog.Infof("external link: %s : %s --> %s", *externalLink.Key, *link.BuyerRef, *link.SellerRef)
			}
			glog.Infof("external link properties for probe entity: ")
			for _, propDef := range link.ProbeEntityPropertyList {
				glog.Infof("%s ", propDef.Name)
			}
			glog.Infof("external link properties for external entity: ")
			for _, propDef := range link.ExternalEntityPropertyList {
				glog.Infof("Attribute: %v", propDef.Attribute)
				glog.Infof("Entity: %v", propDef.Entity)
				glog.Infof("PropHandler: %v", propDef.PropHandler)
			}
		}

		if node.MergedEntityMetaData != nil {
			metadata := node.MergedEntityMetaData

			glog.Infof("KeepInTopology: %v", metadata.KeepInTopology)
			glog.Infof("matching data returnType: %s", metadata.MatchingMetadata.ReturnType)
			glog.Infof("matching data external returnType: %s", metadata.MatchingMetadata.ExternalEntityReturnType)

			if metadata.MatchingMetadata != nil {
				matchingDataList := metadata.MatchingMetadata.MatchingDataList
				glog.Infof("Internal stitching property: ")
				for _, md := range matchingDataList {
					glog.Infof("	MatchingProperty: %v ", md.MatchingProperty)
					glog.Infof("	MatchingField: %v ", md.MatchingField)
				}
				glog.Infof("External stitching property: ")
				for _, md := range metadata.MatchingMetadata.ExternalEntityMatchingPropertyList {
					glog.Infof("	MatchingProperty: %v ", md.MatchingProperty)
					glog.Infof("	MatchingField: %v ", md.MatchingField)
				}
			}
		}
	}
}

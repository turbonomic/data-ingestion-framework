package supplychain

import (
	"fmt"
	set "github.com/deckarep/golang-set"

	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type SupplyChainBuilder struct {
	supplyChainNodes []*proto.TemplateDTO

	err error
}

func NewSupplyChainBuilder() *SupplyChainBuilder {
	return &SupplyChainBuilder{}
}

func (scb *SupplyChainBuilder) Create() ([]*proto.TemplateDTO, error) {
	if scb.err != nil {
		return nil, scb.err
	}
	if err := scb.validateChargedByBoughtRelationships(); err != nil {
		return nil, err
	}
	if err := scb.validateChargedBySoldRelationships(); err != nil {
		return nil, err
	}
	return scb.supplyChainNodes, nil
}

func (scb *SupplyChainBuilder) validateChargedBySoldRelationships() error {
	for _, node := range scb.supplyChainNodes {
		commoditiesSold := getAllCommoditiesSold(node)
		for _, commodity := range node.GetCommoditySold() {
			for _, chargedBySold := range commodity.GetChargedBySold() {
				if !commoditiesSold.Contains(chargedBySold) {
					return fmt.Errorf("commodity %v of entity template %s is charged by commodity"+
						" %s which is not declared to be sold by the entity",
						commodity.GetCommodityType(), node.GetTemplateClass(), chargedBySold)
				}
			}
		}
	}
	return nil
}

func (scb *SupplyChainBuilder) validateChargedByBoughtRelationships() error {
	for _, node := range scb.supplyChainNodes {
		commoditiesBought := getAllCommoditiesBought(node)
		for _, commodity := range node.GetCommoditySold() {
			for _, chargedBy := range commodity.GetChargedBy() {
				if !commoditiesBought.Contains(chargedBy) {
					return fmt.Errorf("commodity %v of entity template %s is charged by commodity"+
						" %s which is not declared to be bought by the entity",
						commodity.GetCommodityType(), node.GetTemplateClass(), chargedBy)
				}
			}
		}
	}
	return nil
}

func getAllCommoditiesBought(node *proto.TemplateDTO) set.Set {
	commoditiesBought := set.NewSet()
	for _, commBoughtProviderMap := range node.GetCommodityBought() {
		for _, commBought := range commBoughtProviderMap.GetValue() {
			commoditiesBought.Add(commBought.GetCommodityType())
		}
	}
	// Note: there is no need to check buyer reference from external link which is deprecated.
	return commoditiesBought
}

func getAllCommoditiesSold(node *proto.TemplateDTO) set.Set {
	commoditiesSold := set.NewSet()
	for _, commSold := range node.GetCommoditySold() {
		commoditiesSold.Add(commSold.GetCommodityType())
	}
	return commoditiesSold
}

// To build a supply chain, must specify the top supply chain node first.
// Here it initializes supplyChainNodes.
func (scb *SupplyChainBuilder) Top(topNode *proto.TemplateDTO) *SupplyChainBuilder {
	if topNode == nil {
		scb.err = fmt.Errorf("top node cannot be nil")
		return scb
	}
	scb.supplyChainNodes = []*proto.TemplateDTO{}
	scb.supplyChainNodes = append(scb.supplyChainNodes, topNode)

	return scb
}

// Add an entity node to supply chain
func (scb *SupplyChainBuilder) Entity(node *proto.TemplateDTO) *SupplyChainBuilder {
	if scb.err != nil {
		return scb
	}
	if scb.supplyChainNodes == nil || len(scb.supplyChainNodes) == 0 {
		scb.err = fmt.Errorf("must set top supply chain node first")
		return scb
	}

	scb.supplyChainNodes = append(scb.supplyChainNodes, node)
	return scb
}

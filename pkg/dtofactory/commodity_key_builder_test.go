package dtofactory

import (
	"github.com/stretchr/testify/assert"
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"testing"
)

func TestCommodityKeyForApplications(t *testing.T) {

	cType := data.SERVICE
	svcId := "service1"

	eType := data.APPLICATION
	appId1 := "app1"
	appEntity1 := data.NewBasicDIFEntity(eType, appId1)
	appEntity1.SetConsumer(cType, svcId)
	appEntity1.SetProvider(data.VM, "vm1")

	appId2 := "app2"
	appEntity2 := data.NewBasicDIFEntity(eType, appId2)
	appEntity2.SetConsumer(cType, svcId)
	appEntity2.SetProvider(data.VM, "vm1")

	kb1 := NewCommodityKeyBuilder(proto.EntityDTO_APPLICATION_COMPONENT, appEntity1)
	key1 := kb1.GetKey()

	kb2 := NewCommodityKeyBuilder(proto.EntityDTO_APPLICATION_COMPONENT, appEntity2)
	key2 := kb2.GetKey()

	assert.NotNil(t, key1)
	assert.NotNil(t, key2)
	assert.EqualValues(t, svcId, *key1)
	assert.EqualValues(t, *key1, *key2)
}

func TestCommodityKeyForService(t *testing.T) {

	eType := data.SERVICE
	svcId := "service1"

	cType := data.APPLICATION
	appId1 := "app1"
	appId2 := "app2"
	svcEntity := data.NewBasicDIFEntity(eType, svcId)
	svcEntity.SetProvider(cType, appId1)
	svcEntity.SetProvider(cType, appId2)

	kb := NewCommodityKeyBuilder(proto.EntityDTO_SERVICE, svcEntity)
	key := kb.GetKey()

	assert.NotNil(t, key)
	assert.EqualValues(t, svcId, *key)
}

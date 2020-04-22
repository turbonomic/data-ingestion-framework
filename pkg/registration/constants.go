package registration

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	// The default namespace of entity property
	DefaultPropertyNamespace = "DEFAULT"
)

// In most cases, the capacity for a commodity should be provided from the input JSON.
// For some commodities, the capacity may have a fixed default value, for example, the
// capacity for garbage collection time and db cache hit rate are all 100% by default.
type DefaultValue struct {
	Key      *string
	Capacity float64
	HasKey   *bool
}

// ===============================================================================

var (
	templateTypeMapping = map[string]proto.TemplateDTO_TemplateType{
		"BASE":      proto.TemplateDTO_BASE,
		"EXTENSION": proto.TemplateDTO_EXTENSION,
	}
	returnTypeMapping = map[string]builder.ReturnType{
		"STRING":      builder.MergedEntityMetadata_STRING,
		"LIST_STRING": builder.MergedEntityMetadata_LIST_STRING,
	}

	relationshipMapping = map[string]proto.Provider_ProviderType{
		"HOSTING":      proto.Provider_HOSTING,
		"LAYERED_OVER": proto.Provider_LAYERED_OVER,
	}
)

// ===============================================================================
// Mapping of entity type strings as specified in supply chain configuration yaml to Turbo SDK Entity Type
var TemplateEntityTypeMap = map[string]proto.EntityDTO_EntityType{
	"SWITCH":                          proto.EntityDTO_SWITCH,
	"VIRTUAL_DATACENTER":              proto.EntityDTO_VIRTUAL_DATACENTER,
	"STORAGE":                         proto.EntityDTO_STORAGE,
	"SERVICE":                         proto.EntityDTO_SERVICE,
	"DATABASE_SERVER":                 proto.EntityDTO_DATABASE_SERVER,
	"SAVINGS":                         proto.EntityDTO_SAVINGS,
	"OPERATOR":                        proto.EntityDTO_OPERATOR,
	"WEB_SERVER":                      proto.EntityDTO_WEB_SERVER,
	"RIGHT_SIZER":                     proto.EntityDTO_RIGHT_SIZER,
	"THREE_TIER_APPLICATION":          proto.EntityDTO_THREE_TIER_APPLICATION,
	"VIRTUAL_MACHINE":                 proto.EntityDTO_VIRTUAL_MACHINE,
	"DISK_ARRAY":                      proto.EntityDTO_DISK_ARRAY,
	"DATACENTER":                      proto.EntityDTO_DATACENTER,
	"INFRASTRUCTURE":                  proto.EntityDTO_INFRASTRUCTURE,
	"PHYSICAL_MACHINE":                proto.EntityDTO_PHYSICAL_MACHINE,
	"CHASSIS":                         proto.EntityDTO_CHASSIS,
	"LICENSING_SERVICE":               proto.EntityDTO_LICENSING_SERVICE,
	"BUSINESS_USER":                   proto.EntityDTO_BUSINESS_USER,
	"STORAGE_CONTROLLER":              proto.EntityDTO_STORAGE_CONTROLLER,
	"HYPERVISOR_SERVER":               proto.EntityDTO_HYPERVISOR_SERVER,
	"BUSINESS_ENTITY":                 proto.EntityDTO_BUSINESS_ENTITY,
	"IO_MODULE":                       proto.EntityDTO_IO_MODULE,
	"ACTION_MANAGER":                  proto.EntityDTO_ACTION_MANAGER,
	"VLAN":                            proto.EntityDTO_VLAN,
	"APPLICATION_SERVER":              proto.EntityDTO_APPLICATION_SERVER,
	"BUSINESS":                        proto.EntityDTO_BUSINESS,
	"VIRTUAL_APPLICATION":             proto.EntityDTO_VIRTUAL_APPLICATION,
	"NETWORKING_ENDPOINT":             proto.EntityDTO_NETWORKING_ENDPOINT,
	"BUSINESS_ACCOUNT":                proto.EntityDTO_BUSINESS_ACCOUNT,
	"IP":                              proto.EntityDTO_IP,
	"SERVICE_ENTITY_TEMPLATE":         proto.EntityDTO_SERVICE_ENTITY_TEMPLATE,
	"PORT":                            proto.EntityDTO_PORT,
	"NETWORK":                         proto.EntityDTO_NETWORK,
	"APPLICATION":                     proto.EntityDTO_APPLICATION,
	"THIS_ENTITY":                     proto.EntityDTO_THIS_ENTITY,
	"COMPUTE_RESOURCE":                proto.EntityDTO_COMPUTE_RESOURCE,
	"MAC":                             proto.EntityDTO_MAC,
	"INTERNET":                        proto.EntityDTO_INTERNET,
	"MOVER ":                          proto.EntityDTO_MOVER,
	"DISTRIBUTED_VIRTUAL_PORTGROUP":   proto.EntityDTO_DISTRIBUTED_VIRTUAL_PORTGROUP,
	"CONTAINER":                       proto.EntityDTO_CONTAINER,
	"CONTAINER_POD":                   proto.EntityDTO_CONTAINER_POD,
	"LOGICAL_POOL":                    proto.EntityDTO_LOGICAL_POOL,
	"CLOUD_SERVICE":                   proto.EntityDTO_CLOUD_SERVICE,
	"DPOD":                            proto.EntityDTO_DPOD,
	"VPOD":                            proto.EntityDTO_VPOD,
	"DATABASE":                        proto.EntityDTO_DATABASE,
	"LOAD_BALANCER":                   proto.EntityDTO_LOAD_BALANCER,
	"BUSINESS_APPLICATION":            proto.EntityDTO_BUSINESS_APPLICATION,
	"PROCESSOR_POOL":                  proto.EntityDTO_PROCESSOR_POOL,
	"STORAGE_VOLUME":                  proto.EntityDTO_STORAGE_VOLUME,
	"RESERVED_INSTANCE":               proto.EntityDTO_RESERVED_INSTANCE,
	"RESERVED_INSTANCE_SPECIFICATION": proto.EntityDTO_RESERVED_INSTANCE_SPECIFICATION,
	"DESIRED_RESERVED_INSTANCE":       proto.EntityDTO_DESIRED_RESERVED_INSTANCE,
	"REGION":                          proto.EntityDTO_REGION,
	"AVAILABILITY_ZONE":               proto.EntityDTO_AVAILABILITY_ZONE,
	"COMPUTE_TIER":                    proto.EntityDTO_COMPUTE_TIER,
	"STORAGE_TIER":                    proto.EntityDTO_STORAGE_TIER,
	"DATABASE_TIER":                   proto.EntityDTO_DATABASE_TIER,
	"DATABASE_SERVER_TIER":            proto.EntityDTO_DATABASE_SERVER_TIER,
	"VIRTUAL_VOLUME":                  proto.EntityDTO_VIRTUAL_VOLUME,
	"VIEW_POD":                        proto.EntityDTO_VIEW_POD,
	"DESKTOP_POOL":                    proto.EntityDTO_DESKTOP_POOL,
	"SERVICE_PROVIDER":                proto.EntityDTO_SERVICE_PROVIDER,
	"BUSINESS_TRANSACTION":            proto.EntityDTO_BUSINESS_TRANSACTION,
	"APPLICATION_COMPONENT":           proto.EntityDTO_APPLICATION_COMPONENT,
}

// Mapping of commodity type strings as specified in supply chain configuration yaml to Turbo SDK Commodity Type
var TemplateCommodityTypeMap = map[string]proto.CommodityDTO_CommodityType{
	"CLUSTER":                    proto.CommodityDTO_CLUSTER,
	"THREADS":                    proto.CommodityDTO_THREADS,
	"CPU_ALLOCATION":             proto.CommodityDTO_CPU_ALLOCATION,
	"NUMBER_CONSUMERS":           proto.CommodityDTO_NUMBER_CONSUMERS,
	"FLOW_ALLOCATION":            proto.CommodityDTO_FLOW_ALLOCATION,
	"Q1_VCPU":                    proto.CommodityDTO_Q1_VCPU,
	"STORAGE_PROVISIONED":        proto.CommodityDTO_STORAGE_PROVISIONED,
	"LICENSE_COMMODITY":          proto.CommodityDTO_LICENSE_COMMODITY,
	"STORAGE_AMOUNT":             proto.CommodityDTO_STORAGE_AMOUNT,
	"Q16_VCPU":                   proto.CommodityDTO_Q16_VCPU,
	"Q32_VCPU":                   proto.CommodityDTO_Q32_VCPU,
	"SAME_CLUSTER_MOVE_SVC":      proto.CommodityDTO_SAME_CLUSTER_MOVE_SVC,
	"Q3_VCPU":                    proto.CommodityDTO_Q3_VCPU,
	"SLA_COMMODITY":              proto.CommodityDTO_SLA_COMMODITY,
	"KPI":                        proto.CommodityDTO_KPI,
	"CROSS_CLUSTER_MOVE_SVC":     proto.CommodityDTO_CROSS_CLUSTER_MOVE_SVC,
	"NUMBER_CONSUMERS_PM":        proto.CommodityDTO_NUMBER_CONSUMERS_PM,
	"STORAGE_ALLOCATION":         proto.CommodityDTO_STORAGE_ALLOCATION,
	"Q8_VCPU":                    proto.CommodityDTO_Q8_VCPU,
	"SPACE":                      proto.CommodityDTO_SPACE,
	"Q6_VCPU":                    proto.CommodityDTO_Q6_VCPU,
	"POWER":                      proto.CommodityDTO_POWER,
	"MEM":                        proto.CommodityDTO_MEM,
	"STORAGE_LATENCY":            proto.CommodityDTO_STORAGE_LATENCY,
	"Q7_VCPU":                    proto.CommodityDTO_Q7_VCPU,
	"COOLING":                    proto.CommodityDTO_COOLING,
	"PORT_CHANEL":                proto.CommodityDTO_PORT_CHANEL,
	"VCPU":                       proto.CommodityDTO_VCPU,
	"QN_VCPU":                    proto.CommodityDTO_QN_VCPU,
	"CPU_PROVISIONED":            proto.CommodityDTO_CPU_PROVISIONED,
	"RIGHT_SIZE_SVC":             proto.CommodityDTO_RIGHT_SIZE_SVC,
	"MOVE":                       proto.CommodityDTO_MOVE,
	"Q2_VCPU":                    proto.CommodityDTO_Q2_VCPU,
	"Q5_VCPU":                    proto.CommodityDTO_Q5_VCPU,
	"SWAPPING":                   proto.CommodityDTO_SWAPPING,
	"SEGMENTATION":               proto.CommodityDTO_SEGMENTATION,
	"FLOW":                       proto.CommodityDTO_FLOW,
	"DATASTORE":                  proto.CommodityDTO_DATASTORE,
	"CROSS_CLOUD_MOVE_SVC":       proto.CommodityDTO_CROSS_CLOUD_MOVE_SVC,
	"RIGHT_SIZE_DOWN":            proto.CommodityDTO_RIGHT_SIZE_DOWN,
	"IO_THROUGHPUT":              proto.CommodityDTO_IO_THROUGHPUT,
	"CPU":                        proto.CommodityDTO_CPU,
	"BALLOONING":                 proto.CommodityDTO_BALLOONING,
	"VDC":                        proto.CommodityDTO_VDC,
	"Q64_VCPU":                   proto.CommodityDTO_Q64_VCPU,
	"CONNECTION":                 proto.CommodityDTO_CONNECTION,
	"MEM_PROVISIONED":            proto.CommodityDTO_MEM_PROVISIONED,
	"STORAGE":                    proto.CommodityDTO_STORAGE,
	"NET_THROUGHPUT":             proto.CommodityDTO_NET_THROUGHPUT,
	"NUMBER_CONSUMERS_STORAGE":   proto.CommodityDTO_NUMBER_CONSUMERS_STORAGE,
	"TRANSACTION":                proto.CommodityDTO_TRANSACTION,
	"MEM_ALLOCATION":             proto.CommodityDTO_MEM_ALLOCATION,
	"DSPM_ACCESS":                proto.CommodityDTO_DSPM_ACCESS,
	"RESPONSE_TIME":              proto.CommodityDTO_RESPONSE_TIME,
	"VMEM":                       proto.CommodityDTO_VMEM,
	"ACTION_PERMIT":              proto.CommodityDTO_ACTION_PERMIT,
	"DATACENTER":                 proto.CommodityDTO_DATACENTER,
	"APPLICATION":                proto.CommodityDTO_APPLICATION,
	"NETWORK":                    proto.CommodityDTO_NETWORK,
	"Q4_VCPU":                    proto.CommodityDTO_Q4_VCPU,
	"STORAGE_CLUSTER":            proto.CommodityDTO_STORAGE_CLUSTER,
	"EXTENT":                     proto.CommodityDTO_EXTENT,
	"ACCESS":                     proto.CommodityDTO_ACCESS,
	"RIGHT_SIZE_UP":              proto.CommodityDTO_RIGHT_SIZE_UP,
	"VAPP_ACCESS":                proto.CommodityDTO_VAPP_ACCESS,
	"STORAGE_ACCESS":             proto.CommodityDTO_STORAGE_ACCESS,
	"VSTORAGE":                   proto.CommodityDTO_VSTORAGE,
	"DRS_SEGMENTATION":           proto.CommodityDTO_DRS_SEGMENTATION,
	"DB_MEM":                     proto.CommodityDTO_DB_MEM,
	"TRANSACTION_LOG":            proto.CommodityDTO_TRANSACTION_LOG,
	"DB_CACHE_HIT_RATE":          proto.CommodityDTO_DB_CACHE_HIT_RATE,
	"HOT_STORAGE":                proto.CommodityDTO_HOT_STORAGE,
	"COLLECTION_TIME":            proto.CommodityDTO_COLLECTION_TIME,
	"BUFFER_COMMODITY":           proto.CommodityDTO_BUFFER_COMMODITY,
	"SOFTWARE_LICENSE_COMMODITY": proto.CommodityDTO_SOFTWARE_LICENSE_COMMODITY,
	"VMPM_ACCESS":                proto.CommodityDTO_VMPM_ACCESS,
	"HA_COMMODITY":               proto.CommodityDTO_HA_COMMODITY,
	"NETWORK_POLICY":             proto.CommodityDTO_NETWORK_POLICY,
	"HEAP":                       proto.CommodityDTO_HEAP,
	"DISK_ARRAY_ACCESS":          proto.CommodityDTO_DISK_ARRAY_ACCESS,
	"SERVICE_LEVEL_CLUSTER":      proto.CommodityDTO_SERVICE_LEVEL_CLUSTER,
	"PROCESSING_UNITS":           proto.CommodityDTO_PROCESSING_UNITS,
	"HOST_LUN_ACCESS":            proto.CommodityDTO_HOST_LUN_ACCESS,
	"COUPON":                     proto.CommodityDTO_COUPON,
	"TENANCY_ACCESS":             proto.CommodityDTO_TENANCY_ACCESS,
	"LICENSE_ACCESS":             proto.CommodityDTO_LICENSE_ACCESS,
	"TEMPLATE_ACCESS":            proto.CommodityDTO_TEMPLATE_ACCESS,
	"NUM_DISK":                   proto.CommodityDTO_NUM_DISK,
	"ZONE":                       proto.CommodityDTO_ZONE,
	"ACTIVE_SESSIONS":            proto.CommodityDTO_ACTIVE_SESSIONS,
	"POOL_CPU":                   proto.CommodityDTO_POOL_CPU,
	"POOL_MEM":                   proto.CommodityDTO_POOL_MEM,
	"POOL_STORAGE":               proto.CommodityDTO_POOL_STORAGE,
	"IMAGE_CPU":                  proto.CommodityDTO_IMAGE_CPU,
	"IMAGE_MEM":                  proto.CommodityDTO_IMAGE_MEM,
	"IMAGE_STORAGE":              proto.CommodityDTO_IMAGE_STORAGE,
	"INSTANCE_DISK_SIZE":         proto.CommodityDTO_INSTANCE_DISK_SIZE,
	"INSTANCE_DISK_TYPE":         proto.CommodityDTO_INSTANCE_DISK_TYPE,
	"BURST_BALANCE":              proto.CommodityDTO_BURST_BALANCE,
	"TEMPLATE_FAMILY":            proto.CommodityDTO_TEMPLATE_FAMILY,
	"DESIRED_COUPON":             proto.CommodityDTO_DESIRED_COUPON,
	"VCPU_REQUEST":               proto.CommodityDTO_VCPU_REQUEST,
	"VMEM_REQUEST":               proto.CommodityDTO_VMEM_REQUEST,
	"CPU_REQUEST_ALLOCATION":     proto.CommodityDTO_VCPU_REQUEST_QUOTA,
	"MEM_REQUEST_ALLOCATION":     proto.CommodityDTO_VMEM_REQUEST_QUOTA,
	"NETWORK_INTERFACE_COUNT":    proto.CommodityDTO_NETWORK_INTERFACE_COUNT,
	"BICLIQUE":                   proto.CommodityDTO_BICLIQUE,
}

var AccessTemplateCommodityTypeMap = map[string]proto.CommodityDTO_CommodityType{
	"CLUSTER":        proto.CommodityDTO_CLUSTER,
	"LICENSE_ACCESS": proto.CommodityDTO_LICENSE_ACCESS,
	"VMPM_ACCESS":    proto.CommodityDTO_VMPM_ACCESS,
	"APPLICATION":    proto.CommodityDTO_APPLICATION,
	"DATACENTER":     proto.CommodityDTO_DATACENTER,
}

// ============================================================================================

// Mapping to the entity that will provide the key value for the commodities
var KeySupplierMapping = map[proto.EntityDTO_EntityType]proto.EntityDTO_EntityType{
	proto.EntityDTO_APPLICATION_COMPONENT: proto.EntityDTO_SERVICE,
	proto.EntityDTO_SERVICE:               proto.EntityDTO_SERVICE,
	proto.EntityDTO_DATABASE_SERVER:       proto.EntityDTO_SERVICE,
}

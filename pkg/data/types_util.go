package data

type DIFEntityType string

const (
	VM                         DIFEntityType = "virtualMachine"
	CONTAINER                  DIFEntityType = "container"
	CONTAINER_POD              DIFEntityType = "containerPod"
	CONTAINER_SPEC             DIFEntityType = "containerSpec"
	WORKLOAD_CONTROLLER        DIFEntityType = "workloadController"
	NAMESPACE                  DIFEntityType = "namespace"
	CONTAINER_PLATFORM_CLUSTER DIFEntityType = "containerPlatformCluster"
	APPLICATION                DIFEntityType = "application"
	SERVICE                    DIFEntityType = "service"
	DATABASE_SERVER            DIFEntityType = "databaseServer"
	BUSINESS_APP               DIFEntityType = "businessApplication"
	BUSINESS_TRANS             DIFEntityType = "businessTransaction"
)

func ParseEntityType(entityType string) DIFEntityType {
	switch entityType {
	case "virtualMachine":
		return VM
	case "container":
		return CONTAINER
	case "containerPod":
		return CONTAINER_POD
	case "containerSpec":
		return CONTAINER_SPEC
	case "workloadController":
		return WORKLOAD_CONTROLLER
	case "namespace":
		return NAMESPACE
	case "containerPlatformCluster":
		return CONTAINER_PLATFORM_CLUSTER
	case "application":
		return APPLICATION
	case "service":
		return SERVICE
	case "databaseServer":
		return DATABASE_SERVER
	case "businessTransaction":
		return BUSINESS_TRANS
	case "businessApplication":
		return BUSINESS_APP
	default:
		return ""
	}
}

// Mapping of the dif entity type string to supply chain template entity type string
// DIF entity type string is obtained from the JSON DIF data.
// Supply chain template entity type strings are defined in registration->constants.go
var DIFEntityTypeToTemplateEntityStringMap = map[DIFEntityType]string{
	VM:                         "VIRTUAL_MACHINE",
	CONTAINER:                  "CONTAINER",
	CONTAINER_POD:              "CONTAINER_POD",
	CONTAINER_SPEC:             "CONTAINER_SPEC",
	WORKLOAD_CONTROLLER:        "WORKLOAD_CONTROLLER",
	NAMESPACE:                  "NAMESPACE",
	CONTAINER_PLATFORM_CLUSTER: "CONTAINER_PLATFORM_CLUSTER",
	APPLICATION:                "APPLICATION_COMPONENT",
	SERVICE:                    "SERVICE",
	DATABASE_SERVER:            "DATABASE_SERVER",
	BUSINESS_APP:               "BUSINESS_APPLICATION",
	BUSINESS_TRANS:             "BUSINESS_TRANSACTION",
}

type DIFMetricType string

const (
	RESPONSE_TIME      DIFMetricType = "responseTime"
	SERVICE_TIME       DIFMetricType = "serviceTime"
	QUEUING_TIME       DIFMetricType = "queuingTime"
	TRANSACTION        DIFMetricType = "transaction"
	CONNECTION         DIFMetricType = "connection"
	CUSTOM             DIFMetricType = "sla"
	CPU                DIFMetricType = "cpu"
	MEMORY             DIFMetricType = "memory"
	THREADS            DIFMetricType = "threads"
	HEAP               DIFMetricType = "heap"
	COLLECTION_TIME    DIFMetricType = "collectionTime"
	DBMEM              DIFMetricType = "dbMem"
	DBCACHEHITRATE     DIFMetricType = "dbCacheHitRate"
	KPI                DIFMetricType = "kpi"
	GPU                DIFMetricType = "gpu"
	GPU_MEM            DIFMetricType = "gpuMem"
	GPU_REQUEST        DIFMetricType = "gpuRequest"
	GPU_REQUEST_QUOTA  DIFMetricType = "gpuRequestQuota"
	LLM_CACHE          DIFMetricType = "llmCache"
	CONCURRENT_QUERIES DIFMetricType = "concurrentQueries"
	ENERGY             DIFMetricType = "energy"
)

// Mapping of the dif metric string to supply chain template commodity string
// DIF metric string is obtained from the JSON DIF data.
// Supply chain template commodity strings are defined in registration->constants.go
var DIFMetricToTemplateCommodityStringMap = map[string]string{
	"cluster":           "CLUSTER",
	"threads":           "THREADS",
	"cpu":               "VCPU",
	"io":                "IO_THROUGHPUT",
	"connection":        "CONNECTION",
	"netThroughput":     "NET_THROUGHPUT",
	"transaction":       "TRANSACTION",
	"responseTime":      "RESPONSE_TIME",
	"serviceTime":       "SERVICE_TIME",
	"queuingTime":       "QUEUING_TIME",
	"memory":            "VMEM",
	"application":       "APPLICATION",
	"dbMem":             "DB_MEM",
	"transactionLog":    "TRANSACTION_LOG",
	"dbCacheHitRate":    "DB_CACHE_HIT_RATE",
	"collectionTime":    "REMAINING_GC_CAPACITY",
	"heap":              "HEAP",
	"kpi":               "KPI",
	"gpu":               "GPU",
	"gpuMem":            "GPU_MEM",
	"gpuRequest":        "GPU_REQUEST",
	"gpuRequestQuota":   "GPU_REQUEST_QUOTA",
	"llmCache":          "LLM_CACHE",
	"concurrentQueries": "CONCURRENT_QUERIES",
	"energy":            "ENERGY",
}

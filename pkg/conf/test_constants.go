package conf

const (
	SERVICE               string = "SERVICE"
	BUSINESS_APPLICATION  string = "BUSINESS_APPLICATION"
	BUSINESS_TRANSACTION  string = "BUSINESS_TRANSACTION"
	APPLICATION_COMPONENT string = "APPLICATION_COMPONENT"
	VIRTUAL_MACHINE       string = "VIRTUAL_MACHINE"
	DATABASE_SERVER       string = "DATABASE_SERVER"

	TRANSACTION     string = "TRANSACTION"
	RESPONSE_TIME   string = "RESPONSE_TIME"
	APPLICATION     string = "APPLICATION"
	COLLECTION_TIME string = "COLLECTION_TIME"
	THREADS         string = "THREADS"
	HEAP            string = "HEAP"
	VCPU            string = "VCPU"
	VMEM            string = "VMEM"

	LIST_STRING string = "LIST_STRING"
	STRING      string = "STRING"

	LAYERED_OVER string = "LAYERED_OVER"
	HOSTING      string = "HOSTING"
)

var (
	SERVICE_NODE = "supplyChainNode:\n" +
		" - templateClass: SERVICE\n" +
		"   templateType: BASE\n" +
		"   templatePriority: -1\n" +
		"   commoditySold:\n" +
		"     - commodityType: TRANSACTION\n" +
		"       key: key-placeholder\n" +
		"     - commodityType: RESPONSE_TIME\n" +
		"       key: key-placeholder\n" +
		"     - commodityType: APPLICATION\n" +
		"   commodityBought:\n" +
		"     - key:\n" +
		"         templateClass: APPLICATION_COMPONENT\n" +
		"         providerType: LAYERED_OVER\n" +
		"         cardinalityMax: 2147483647\n" +
		"         cardinalityMin: 0\n" +
		"       value:\n" +
		"         - commodityType: TRANSACTION\n" +
		"           key: key-placeholder\n" +
		"         - commodityType: RESPONSE_TIME\n" +
		"           key: key-placeholder\n" +
		"         - commodityType: APPLICATION\n" +
		"     - key:\n" +
		"         templateClass: DATABASE_SERVER\n" +
		"         providerType: LAYERED_OVER\n" +
		"         cardinalityMax: 2147483647\n" +
		"         cardinalityMin: 0\n" +
		"       value:\n" +
		"         - commodityType: TRANSACTION\n" +
		"           key: key-placeholder\n" +
		"         - commodityType: RESPONSE_TIME\n" +
		"           key: key-placeholder\n" +
		"         - commodityType: APPLICATION\n" +
		"   mergedEntityMetaData:\n" +
		"     keepStandalone: false\n" +
		"     matchingMetadata:\n" +
		"       returnType: STRING\n" +
		"       matchingData:\n" +
		"         - matchingProperty:\n" +
		"             propertyName: IP\n" +
		"       externalEntityReturnType: LIST_STRING\n" +
		"       externalEntityMatchingProperty:\n" +
		"         - matchingProperty:\n" +
		"             propertyName: IP\n" +
		"           delimiter: \",\"\n" +
		"     commoditiesSold:\n" +
		"       - RESPONSE_TIME\n" +
		"       - TRANSACTION\n" +
		"     commoditiesBought:\n" +
		"       - providerType: APPLICATION_COMPONENT\n" +
		"         commodityMetadata:\n" +
		"           - RESPONSE_TIME\n" +
		"           - TRANSACTION\n"
)

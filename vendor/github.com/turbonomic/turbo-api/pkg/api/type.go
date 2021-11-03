package api

// CommunicationBindingChannel is an InputField name indicating the field value is the communication binding channel
const CommunicationBindingChannel = "communicationBindingChannel"

type Target struct {
	// Category of a probe, i.e. Hypervisor, Storage and so on.
	Category    string `json:"category,omitempty"`
	ClassName   string `json:"className,omitempty"`
	DisplayName string `json:"displayName,omitempty"`

	// List of field names, identifying the target of this type.
	IdentifyingFields []string `json:"identifyingFields,omitempty"`

	// List of all the account values of the target or probe.
	InputFields []*InputField `json:"inputFields,omitempty"`

	// Date of the last validation.
	LastValidated string  `json:"lastValidated,omitempty"`
	Links         []*Link `json:"links,omitempty"`

	// Description of the status.
	Status string `json:"status,omitempty"`

	// Probe type, i.ee vCenter, Hyper-V and so on.
	Type string `json:"type"`
	UUID string `json:"uuid,omitempty"`
}

// TargetInfo defines the protocols of the GET /target topology processor service
type TargetInfo struct {
	TargetID    int64       `json:"targetId,string"`
	DisplayName string      `json:"displayName"`
	TargetSpec  *TargetSpec `json:"spec"`
}

// TargetSpec defines the protocols of the POST /target topology processor service
type TargetSpec struct {
	// Probe to which the target belongs
	ProbeID int64 `json:"probeId,string"`
	// Is the target hidden from users
	IsHidden bool `json:"isHidden"`
	// Whether the target can be changed through APIs
	ReadOnly bool `json:"readOnly"`
	// The derived target ID associated with this target
	DerivedTargetIDs []string `json:"derivedTargetIds"`
	// Account values to use to add the target
	InputFields []*InputField `json:"inputFields,omitempty"`
	// The communication channel of a target
	CommunicationBindingChannel string `json:"communicationBindingChannel,omitempty"`
}

// ProbeDescription defines the protocols of the GET /probe topology-processor service
type ProbeDescription struct {
	ID       int64  `json:"id,string"`
	Category string `json:"category"`
	Type     string `json:"type"`
}

type InputField struct {
	ClassName string `json:"className,omitempty"`

	// Default value of the field
	DefaultValue string `json:"defaultName,omitempty"`

	// Additional information about what the input to the field should be
	Description string `json:"description,omitempty"`
	DisplayName string `json:"displayName,omitempty"`

	// Group scope structure, filled if this field represents group scope value
	GroupProperties []*List `json:"groupProperties"`

	// Whether the field is mandatory. Valid targets must have all the mandatory fields set.
	IsMandatory bool `json:"isMandatory,omitempty"`

	// Whether the field is secret. This means, that field value is stored in an encrypted value and not shown in any logs.
	IsSecret bool    `json:"isSecret,omitempty"`
	Links    []*Link `json:"links,omitempty"`

	// Name of the field, used for field identification.
	Name string `json:"name"`
	UUID string `json:"uuid,omitempty"`

	// Field value. Used if field holds primitive value (String, number or boolean.
	Value string `json:"value,omitempty"`

	// Type of the value this field holds = ['STRING', 'BOOLEAN', 'NUMERIC', 'GROUP_SCOPE']
	ValueType string `json:"valueType,omitempty"`

	// The regex pattern that needs to be satisfied for the input field text
	VerificationRegex string `json:"verificationRegex,omitempty"`
}

type Link struct {
	HRef      string `json:"href,omitempty"`
	Rel       string `json:"rel,omitempty"`
	Templated bool   `json:"templated,omitempty"`
}

type List struct{}

type APIErrorDTO struct {
	ResponseType int    `json:"type"`
	Exception    string `json:"exception,omitempty"`
	Message      string `json:"message"`
}

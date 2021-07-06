package dbapi

import "github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/api"

type ConnectorStatusPhase = string

const (
	ConnectorStatusPhaseAssigning    ConnectorStatusPhase = "assigning"    // set by kas-fleet-manager - user request
	ConnectorStatusPhaseAssigned     ConnectorStatusPhase = "assigned"     // set by kas-fleet-manager - worker
	ConnectorStatusPhaseUpdating     ConnectorStatusPhase = "updating"     // set by kas-fleet-manager - user request
	ConnectorStatusPhaseProvisioning ConnectorStatusPhase = "provisioning" // set by kas-agent
	ConnectorStatusPhaseReady        ConnectorStatusPhase = "ready"        // set by the agent
	ConnectorStatusPhaseFailed       ConnectorStatusPhase = "failed"       // set by the agent
	ConnectorStatusPhaseDeleting     ConnectorStatusPhase = "deleting"     // set by the kas-fleet-manager - user request
	ConnectorStatusPhaseDeleted      ConnectorStatusPhase = "deleted"      // set by the agent
)

var AgentSetConnectorStatusPhase = []ConnectorStatusPhase{
	ConnectorStatusPhaseProvisioning,
	ConnectorStatusPhaseReady,
	ConnectorStatusPhaseFailed,
	ConnectorStatusPhaseDeleted,
}

type TargetKind = string

const (
	AddonTargetKind         TargetKind = "addon"
	CloudProviderTargetKind TargetKind = "cloud_provider"
)

var AllTargetKind = []TargetKind{
	AddonTargetKind,
	CloudProviderTargetKind,
}

type Connector struct {
	api.Meta

	TargetKind     TargetKind
	AddonClusterId string
	CloudProvider  string
	Region         string
	MultiAZ        bool

	Name           string
	Owner          string
	OrganisationId string
	KafkaID        string
	Version        int64 `gorm:"type:bigserial;index:"`

	ConnectorTypeId string
	ConnectorSpec   api.JSON `gorm:"type:jsonb"`
	DesiredState    string
	Channel         string
	Kafka           KafkaConnectionSettings `gorm:"embedded;embeddedPrefix:kafka_"`

	Status ConnectorStatus `gorm:"foreignKey:ID"`
}

type ConnectorStatus struct {
	api.Meta
	ClusterID string
	Phase     string
}

type ConnectorList []*Connector

// ConnectorDeployment Holds the deployment configuration of a connector
type ConnectorDeployment struct {
	api.Meta
	Version                int64
	ConnectorID            string
	ConnectorVersion       int64
	ConnectorTypeChannelId int64
	ClusterID              string
	AllowUpgrade           bool
	Status                 ConnectorDeploymentStatus `gorm:"foreignKey:ID"`
}

type ConnectorDeploymentList []ConnectorDeployment

type ConnectorDeploymentStatus struct {
	api.Meta
	Phase            string
	Version          int64
	Conditions       api.JSON `gorm:"type:jsonb"`
	Operators        api.JSON `gorm:"type:jsonb"`
	UpgradeAvailable bool
}

type ConnectorDeploymentSpecStatusExtractors struct {
	ApiVersion    string
	Kind          string
	Name          string
	JsonPath      string
	ConditionType string
}

type KafkaConnectionSettings struct {
	BootstrapServer string
	ClientId        string
	ClientSecret    string `gorm:"-"`
	ClientSecretRef string `gorm:"column:client_secret"`
}

type ConnectorDeploymentAvailableUpgrades struct {
	DeploymentID    string                    `json:"deployment_id,omitempty"`
	ConnectorTypeId string                    `json:"connector_type_id,omitempty"`
	Channel         string                    `json:"channel,omitempty"`
	ShardMetadata   *ConnectorTypeUpgrade     `json:"shard_metadata,omitempty"`
	Operator        *ConnectorOperatorUpgrade `json:"operator,omitempty"`
}
type ConnectorTypeUpgrade struct {
	AssignedId  int64 `json:"assigned_id,omitempty"`
	AvailableId int64 `json:"available_id,omitempty"`
}
type ConnectorOperatorUpgrade struct {
	Assigned  ConnectorOperator `json:"assigned"`
	Available ConnectorOperator `json:"available"`
}
type ConnectorOperator struct {
	// the id of the operator
	Id string `json:"id,omitempty"`
	// the type of the operator
	Type string `json:"type,omitempty"`
	// the version of the operator
	Version string `json:"version,omitempty"`
}
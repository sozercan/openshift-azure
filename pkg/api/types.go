package api

const (
	// APIVersion is the version of this API
	APIVersion = "internal"
)

// OpenShiftManagedCluster complies with the ARM model of resource definition in
// a JSON template.
type OpenShiftManagedCluster struct {
	Plan       *ResourcePurchasePlan `json:"plan,omitempty"`
	Properties *Properties           `json:"properties,omitempty"`
	ID         string                `json:"id,omitempty"`
	Name       string                `json:"name,omitempty"`
	Type       string                `json:"type,omitempty"`
	Location   string                `json:"location,omitempty"`
	Tags       map[string]string     `json:"tags"`

	Config *Config `json:"config,omitempty"`
}

// ResourcePurchasePlan defines the resource plan as required by ARM for billing
// purposes.
type ResourcePurchasePlan struct {
	Name          string `json:"name,omitempty"`
	Product       string `json:"product,omitempty"`
	PromotionCode string `json:"promotionCode,omitempty"`
	Publisher     string `json:"publisher,omitempty"`
}

// Properties represents the cluster definition.
type Properties struct {
	// ProvisioningState (out): current state of the OSA resource.
	ProvisioningState ProvisioningState `json:"provisioningState,omitempty"`

	// OpenShiftVersion (in): OpenShift version to be created/updated, e.g.
	// `v3.11`.
	OpenShiftVersion string `json:"openShiftVersion,omitempty"`

	// PublicHostname (in,optional): Optional user-specified FQDN for OpenShift
	// API server.  If specified, after OSA cluster creation, user must create a
	// PublicHostname CNAME record forwarding to the returned FQDN value.
	PublicHostname string `json:"publicHostname,omitempty"`

	// FQDN (in): FQDN for OpenShift API server.  User-specified FQDN for
	// OpenShift API server loadbalancer internal hostname.
	FQDN string `json:"fqdn,omitempty"`

	// NetworkProfile (in): Configuration for OpenShift networking.
	NetworkProfile *NetworkProfile `json:"networkProfile,omitempty"`

	// RouterProfiles (in,optional/out): Configuration for OpenShift router(s).
	RouterProfiles []RouterProfile `json:"routerProfiles,omitempty"`

	// AgentPoolProfiles (in): configuration of OpenShift cluster VMs.
	AgentPoolProfiles []AgentPoolProfile `json:"agentPoolProfiles,omitempty"`

	// AuthProfile (in): configures OpenShift authentication
	AuthProfile *AuthProfile `json:"authProfile,omitempty"`

	ServicePrincipalProfile *ServicePrincipalProfile `json:"servicePrincipalProfile,omitempty"`

	AzProfile *AzProfile `json:"azProfile,omitempty"`
}

// ProvisioningState represents the current state of the OSA resource.
type ProvisioningState string

const (
	// Creating means the OSA resource is being created.
	Creating ProvisioningState = "Creating"
	// Updating means the existing OSA resource is being updated.
	Updating ProvisioningState = "Updating"
	// Failed means the OSA resource is in failed state.
	Failed ProvisioningState = "Failed"
	// Succeeded means the last create/update succeeded.
	Succeeded ProvisioningState = "Succeeded"
	// Deleting means the OSA resource is being deleted.
	Deleting ProvisioningState = "Deleting"
	// Migrating means the OSA resource is being migrated from one subscription
	// or resource group to another.
	Migrating ProvisioningState = "Migrating"
	// Upgrading means the existing OAS resource is being upgraded.
	Upgrading ProvisioningState = "Upgrading"
)

// NetworkProfile contains configuration for OpenShift networking.
type NetworkProfile struct {
	VnetCIDR string `json:"vnetCidr,omitempty"`

	// PeerVnetID is expected to be empty or match
	// `^/subscriptions/[^/]+
	//   /resourceGroups/[^/]+
	//   /providers/Microsoft.Network
	//   /virtualNetworks/[^/]+$`
	PeerVnetID string `json:"peerVnetId,omitempty"`
}

// RouterProfile represents an OpenShift router.
type RouterProfile struct {
	Name string `json:"name,omitempty"`

	// PublicSubdomain (in,optional/out): DNS subdomain for OpenShift router. If
	// specified, after OSA cluster creation, user must create a (wildcard)
	// *.PublicSubdomain CNAME record forwarding to the returned FQDN value.  If
	// not specified, OSA will auto-allocate and setup a PublicSubdomain and
	// return it.  The OpenShift master is configured with the PublicSubdomain
	// of the "default" RouterProfile.
	PublicSubdomain string `json:"publicSubdomain,omitempty"`

	// FQDN (out): Auto-allocated FQDN for the OpenShift router.
	FQDN string `json:"fqdn,omitempty"`
}

// AgentPoolProfile represents configuration of OpenShift cluster VMs.
type AgentPoolProfile struct {
	Name       string `json:"name,omitempty"`
	Count      int    `json:"count,omitempty"`
	VMSize     VMSize `json:"vmSize,omitempty"`
	SubnetCIDR string `json:"subnetCidr,omitempty"`
	OSType     OSType `json:"osType,omitempty"`

	Role AgentPoolProfileRole `json:"role,omitempty"`
}

// OSType represents the OS type of VMs in an AgentPool.
type OSType string

const (
	// OSTypeLinux is Linux.
	OSTypeLinux OSType = "Linux"
	// OSTypeWindows is Windows.
	OSTypeWindows OSType = "Windows"
)

// AgentPoolProfileRole represents the role of the AgentPoolProfile.
type AgentPoolProfileRole string

const (
	// AgentPoolProfileRoleCompute is the compute role.
	AgentPoolProfileRoleCompute AgentPoolProfileRole = "compute"
	// AgentPoolProfileRoleInfra is the infra role.
	AgentPoolProfileRoleInfra AgentPoolProfileRole = "infra"
	// AgentPoolProfileRoleMaster is the master role.
	AgentPoolProfileRoleMaster AgentPoolProfileRole = "master"
)

// VMSize represents supported VMSizes
type VMSize string

// VMSizes
const (
	// General purpose VMs
	StandardD2sV3  VMSize = "Standard_D2s_v3"
	StandardD4sV3  VMSize = "Standard_D4s_v3"
	StandardD8sV3  VMSize = "Standard_D8s_v3"
	StandardD16sV3 VMSize = "Standard_D16s_v3"
	StandardD32sV3 VMSize = "Standard_D32s_v3"
	StandardD64sV3 VMSize = "Standard_D64s_v3"

	StandardDS4V2 VMSize = "Standard_DS4_v2"
	StandardDS5V2 VMSize = "Standard_DS5_v2"

	// Compute optimized VMs
	StandardF8sV2  VMSize = "Standard_F8s_v2"
	StandardF16sV2 VMSize = "Standard_F16s_v2"
	StandardF32sV2 VMSize = "Standard_F32s_v2"
	StandardF64sV2 VMSize = "Standard_F64s_v2"
	StandardF72sV2 VMSize = "Standard_F72s_v2"

	StandardF8s  VMSize = "Standard_F8s"
	StandardF16s VMSize = "Standard_F16s"

	// Memory optimized VMs
	StandardE4sV3  VMSize = "Standard_E4s_v3"
	StandardE8sV3  VMSize = "Standard_E8s_v3"
	StandardE16sV3 VMSize = "Standard_E16s_v3"
	StandardE20sV3 VMSize = "Standard_E20s_v3"
	StandardE32sV3 VMSize = "Standard_E32s_v3"
	StandardE64sV3 VMSize = "Standard_E64s_v3"

	StandardGS2 VMSize = "Standard_GS2"
	StandardGS3 VMSize = "Standard_GS3"
	StandardGS4 VMSize = "Standard_GS4"
	StandardGS5 VMSize = "Standard_GS5"

	StandardDS12V2 VMSize = "Standard_DS12_v2"
	StandardDS13V2 VMSize = "Standard_DS13_v2"
	StandardDS14V2 VMSize = "Standard_DS14_v2"
	StandardDS15V2 VMSize = "Standard_DS15_v2"

	// Storage optimized VMs
	StandardL4s  VMSize = "Standard_L4s"
	StandardL8s  VMSize = "Standard_L8s"
	StandardL16s VMSize = "Standard_L16s"
	StandardL32s VMSize = "Standard_L32s"
)

// AuthProfile defines all possible authentication profiles for the OpenShift
// cluster.
type AuthProfile struct {
	IdentityProviders []IdentityProvider `json:"identityProviders,omitempty"`
}

// IdentityProvider is heavily cut down equivalent to IdentityProvider in the
// upstream.
type IdentityProvider struct {
	Name     string      `json:"name,omitempty"`
	Provider interface{} `json:"provider,omitempty"`
}

// AADIdentityProvider defines Identity provider for MS AAD.  It is based on
// OpenID IdentityProvider.
type AADIdentityProvider struct {
	Kind     string `json:"kind,omitempty"`
	ClientID string `json:"clientId,omitempty"`
	Secret   string `json:"secret,omitempty" conform:"redact"`
	TenantID string `json:"tenantId,omitempty"`
}

// ServicePrincipalProfile contains the client and secret used by the cluster
// for Azure Resource CRUD.
type ServicePrincipalProfile struct {
	ClientID string `json:"clientId,omitempty"`
	Secret   string `json:"secret,omitempty" conform:"redact"`
}

// AzProfile holds the azure context for where the cluster resides
type AzProfile struct {
	TenantID       string `json:"tenantId,omitempty"`
	SubscriptionID string `json:"subscriptionId,omitempty"`
	ResourceGroup  string `json:"resourceGroup,omitempty"`
}

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MulticlusterConfigSpec defines the desired state of MulticlusterConfig
// +k8s:openapi-gen=true
type MulticlusterConfigSpec struct {

	// If this field is omitted, "Standalone" is assumed. For a scenario with multiple clusters, one "Management"
	// cluster can be configured to establish a secure connection with one or more "Managed" clusters.
	// Valid values for this field are: "Standalone", "Management", "Managed".
	// +optional
	// +kubebuilder:validation:Enum=Standalone,Management,Managed
	ClusterManagementType ClusterManagementType `json:"clusterManagementType,omitempty"`

	// Specify where the managed cluster can reach the management cluster. Ex.: "10.128.0.10:30449". A managed cluster
	// should be able to access this address. This field is not used in standalone clusters.
	// +optional
	ManagementClusterAddr string `json:"managementClusterAddr,omitempty"`
}

// ClusterType represents the type of multicluster management to use. Valid options for this field are: "Standalone",
// "Management", "Managed".
type ClusterManagementType string

const (
	ClusterManagementTypeStandalone ClusterManagementType = "Standalone"
	ClusterManagementTypeManagement ClusterManagementType = "Management"
	ClusterManagementTypeManaged    ClusterManagementType = "Managed"
)

// MulticlusterConfigStatus defines the observed state of MulticlusterConfig.
// +k8s:openapi-gen=true
type MulticlusterConfigStatus struct {

	// State provides user-readable status.
	// +optional
	State string `json:"state,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced

// MulticlusterConfig installs the components required for multicluster management. At most one instance
// of this resource is supported. It must be named "tigera-secure".
// +k8s:openapi-gen=true
type MulticlusterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MulticlusterConfigSpec   `json:"spec,omitempty"`
	Status MulticlusterConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MulticlusterConfigList contains a list of MulticlusterConfig.
type MulticlusterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MulticlusterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MulticlusterConfig{}, &MulticlusterConfigList{})
}

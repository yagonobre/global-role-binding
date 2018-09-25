package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GLobalRoleBindingSpec defines the desired state of GLobalRoleBinding
type GLobalRoleBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// GLobalRoleBindingStatus defines the observed state of GLobalRoleBinding
type GLobalRoleBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// GLobalRoleBinding is the Schema for the globalrolebindings API
// +k8s:openapi-gen=true
type GLobalRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GLobalRoleBindingSpec   `json:"spec,omitempty"`
	Status GLobalRoleBindingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// GLobalRoleBindingList contains a list of GLobalRoleBinding
type GLobalRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GLobalRoleBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GLobalRoleBinding{}, &GLobalRoleBindingList{})
}

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Konfig is a top-level type
type Konfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Status KonfigStatus `json:"status,omitempty"`
	Spec   KonfigSpec   `json:"spec,omitempty"`
}

type KonfigSpec struct {
	Data map[string]string `json:"data,omitempty"`
}

type KonfigStatus struct {
	Synced bool `json:"synced"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type KonfigList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Konfig `json:"items"`
}

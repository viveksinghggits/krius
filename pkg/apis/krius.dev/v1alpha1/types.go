package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Konfig `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Sekret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SekretSpec   `json:"spec,omitempty"`
	Status            SekretStatus `json:"status,omitempty"`
}

type SekretSpec struct {
	Type corev1.SecretType `json:"type,omitempty"`
	Data map[string]string `json:"data,omitempty"`
}

type SekretStatus struct {
	Synced bool `json:"synced,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SekretList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Sekret `json:"items"`
}

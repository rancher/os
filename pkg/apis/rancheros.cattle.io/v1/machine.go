package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MachineInventory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineInventorySpec   `json:"spec"`
	Status MachineInventoryStatus `json:"status"`
}

type MachineInventorySpec struct {
	ClusterName            string               `json:"clusterName,omitempty"`
	MachineTokenSecretName string               `json:"machineTokenSecretName,omitempty"`
	Config                 MachineRuntimeConfig `json:"config,omitempty"`
}

type MachineRuntimeConfig struct {
	Role            string            `json:"role,omitempty"`
	NodeName        string            `json:"nodeName,omitempty"`
	Address         string            `json:"address,omitempty"`
	InternalAddress string            `json:"internalAddress,omitempty"`
	Taints          []corev1.Taint    `json:"taints,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	ConfigValues    map[string]string `json:"extraConfig,omitempty"`
}

type MachineInventoryStatus struct {
	ClusterRegistrationTokenNamespace string `json:"clusterRegistrationTokenNamespace,omitempty"`
	ClusterRegistrationTokenName      string `json:"clusterRegistrationTokenName,omitempty"`
}

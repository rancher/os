package v1

import (
	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/wrangler/pkg/genericcondition"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MachineRegistration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineRegistrationSpec   `json:"spec"`
	Status MachineRegistrationStatus `json:"status"`
}

type MachineRegistrationSpec struct {
	MachineName                 string               `json:"machineName,omitempty"`
	MachineInventoryLabels      map[string]string    `json:"machineInventoryLabels,omitempty"`
	MachineInventoryAnnotations map[string]string    `json:"machineInventoryAnnotations,omitempty"`
	CloudConfig                 *v1alpha1.GenericMap `json:"cloudConfig,omitempty"`
}

type MachineRegistrationStatus struct {
	Conditions        []genericcondition.GenericCondition `json:"conditions,omitempty"`
	RegistrationURL   string                              `json:"registrationURL,omitempty"`
	RegistrationToken string                              `json:"registrationToken,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MachineInventory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineInventorySpec   `json:"spec"`
	Status MachineInventoryStatus `json:"status"`
}

type MachineInventorySpec struct {
	TPMHash                string               `json:"tpmHash,omitempty"`
	SMBIOS                 *v1alpha1.GenericMap `json:"smbios,omitempty"`
	ClusterName            string               `json:"clusterName"`
	MachineTokenSecretName string               `json:"machineTokenSecretName,omitempty"`
	Config                 MachineRuntimeConfig `json:"config,omitempty"`
}

type MachineRuntimeConfig struct {
	Role            string            `json:"role"`
	NodeName        string            `json:"nodeName,omitempty"`
	Address         string            `json:"address,omitempty"`
	InternalAddress string            `json:"internalAddress,omitempty"`
	Taints          []corev1.Taint    `json:"taints,omitempty"`
	Labels          map[string]string `json:"labels"`
}

type MachineInventoryStatus struct {
	ClusterRegistrationTokenNamespace string `json:"clusterRegistrationTokenNamespace,omitempty"`
	ClusterRegistrationTokenName      string `json:"clusterRegistrationTokenName,omitempty"`
}

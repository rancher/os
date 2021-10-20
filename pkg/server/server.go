package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/rancher/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/os2/pkg/clients"
	ranchercontrollers "github.com/rancher/os2/pkg/generated/controllers/management.cattle.io/v3"
	roscontrollers "github.com/rancher/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	corecontrollers "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	tokenType                = "rancheros.cattle.io/token"
	tokenKey                 = "token"
	tokenIndex               = "tokenIndex"
	machineBySecretNameIndex = "machineBySecretNameIndex"
)

type InventoryServer struct {
	secretCache              corecontrollers.SecretCache
	settingCache             ranchercontrollers.SettingCache
	machineCache             roscontrollers.MachineInventoryCache
	clusterRegistrationToken ranchercontrollers.ClusterRegistrationTokenCache
}

func New(clients *clients.Clients) *InventoryServer {
	server := &InventoryServer{
		secretCache:              clients.Core.Secret().Cache(),
		settingCache:             clients.Rancher.Setting().Cache(),
		machineCache:             clients.OS.MachineInventory().Cache(),
		clusterRegistrationToken: clients.Rancher.ClusterRegistrationToken().Cache(),
	}

	server.secretCache.AddIndexer(tokenIndex, func(obj *corev1.Secret) ([]string, error) {
		if string(obj.Type) != tokenType {
			return nil, nil
		}
		t := obj.Data[tokenKey]
		if len(t) == 0 {
			return nil, nil
		}
		return []string{base64.StdEncoding.EncodeToString(t)}, nil
	})

	server.machineCache.AddIndexer(machineBySecretNameIndex, func(obj *v1.MachineInventory) ([]string, error) {
		if obj.Spec.MachineTokenSecretName == "" {
			return nil, nil
		}
		return []string{obj.Namespace + "/" + obj.Spec.MachineTokenSecretName}, nil
	})

	server.secretCache.AddIndexer(tokenHash, func(obj *corev1.Secret) ([]string, error) {
		if string(obj.Type) == tokenType {
			return nil, nil
		}
		if token := obj.Data[tokenKey]; len(token) > 0 {
			hash := sha256.Sum256(token)
			return []string{base64.StdEncoding.EncodeToString(hash[:])}, nil
		}
		return nil, nil
	})

	return server
}

func (i *InventoryServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, "/cacerts") {
		i.cacerts(resp, req)
	} else {
		err := i.handle(resp, req)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (i *InventoryServer) handle(resp http.ResponseWriter, req *http.Request) error {
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	secrets, err := i.secretCache.GetByIndex(tokenIndex, token)
	if apierrors.IsNotFound(err) {
		http.Error(resp, "Token not found", http.StatusNotFound)
		return nil
	} else if err != nil {
		return err
	} else if len(secrets) > 1 {
		logrus.Errorf("Multiple machine secrets with the same value [%s/%s, %s/%s, ...]",
			secrets[0].Namespace, secrets[0].Name, secrets[1].Namespace, secrets[1].Name)
		http.Error(resp, "Token not found", http.StatusNotFound)
		return nil
	}

	machines, err := i.machineCache.GetByIndex(machineBySecretNameIndex, secrets[0].Namespace+"/"+secrets[0].Name)
	if len(machines) > 1 {
		logrus.Errorf("Multiple machine inventories with the token: %v", machines)
	}
	if apierrors.IsNotFound(err) || len(machines) != 1 {
		http.Error(resp, "Machine not found", http.StatusNotFound)
		return nil
	} else if err != nil {
		return err
	}

	crt, err := i.clusterRegistrationToken.Get(machines[0].Status.ClusterRegistrationTokenNamespace,
		machines[0].Status.ClusterRegistrationTokenName)
	if apierrors.IsNotFound(err) || crt.Status.Token == "" {
		http.Error(resp, "Cluster token not found", http.StatusNotFound)
		return nil
	}

	return writeResponse(resp, machines[0], crt)
}

type config struct {
	Role            string            `json:"role,omitempty"`
	NodeName        string            `json:"nodeName,omitempty"`
	Address         string            `json:"address,omitempty"`
	InternalAddress string            `json:"internalAddress,omitempty"`
	Taints          []string          `json:"taints,omitempty"`
	Labels          []string          `json:"labels,omitempty"`
	ConfigValues    map[string]string `json:"extraConfig,omitempty"`
	Token           string            `json:"token,omitempty"`
}

func writeResponse(resp http.ResponseWriter, inventory *v1.MachineInventory, crt *v3.ClusterRegistrationToken) error {
	config := config{
		Role:            inventory.Spec.Config.Role,
		NodeName:        inventory.Spec.Config.NodeName,
		Address:         inventory.Spec.Config.Address,
		InternalAddress: inventory.Spec.Config.InternalAddress,
		Taints:          nil,
		Labels:          nil,
		ConfigValues:    inventory.Spec.Config.ConfigValues,
		Token:           crt.Status.Token,
	}
	for k, v := range inventory.Spec.Config.Labels {
		config.Labels = append(config.Labels, fmt.Sprintf("%s=%s", k, v))
	}
	for _, taint := range inventory.Spec.Config.Taints {
		config.Labels = append(config.Labels, taint.ToString())
	}
	resp.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(resp).Encode(config)
}

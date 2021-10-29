package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	v1 "github.com/rancher/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/os2/pkg/clients"
	ranchercontrollers "github.com/rancher/os2/pkg/generated/controllers/management.cattle.io/v3"
	roscontrollers "github.com/rancher/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	"github.com/rancher/os2/pkg/tpm"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	corecontrollers "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	tokenType                = "rancheros.cattle.io/token"
	tokenKey                 = "token"
	tokenIndex               = "tokenIndex"
	machineBySecretNameIndex = "machineBySecretNameIndex"
	registrationTokenIndex   = "registrationTokenIndex"
	tpmHashIndex             = "tpmHashIndex"
)

type authenticator interface {
	Authenticate(resp http.ResponseWriter, req *http.Request, registerNamespace string) (*v1.MachineInventory, bool, io.WriteCloser, error)
}

type InventoryServer struct {
	settingCache             ranchercontrollers.SettingCache
	secretCache              corecontrollers.SecretCache
	machineCache             roscontrollers.MachineInventoryCache
	machineClient            roscontrollers.MachineInventoryClient
	machineRegistrationCache roscontrollers.MachineRegistrationCache
	clusterRegistrationToken ranchercontrollers.ClusterRegistrationTokenCache
	authenticators           []authenticator
}

func New(clients *clients.Clients) *InventoryServer {
	server := &InventoryServer{
		authenticators: []authenticator{
			tpm.New(clients),
			newSharedSecretAuth(clients),
		},
		secretCache:              clients.Core.Secret().Cache(),
		machineCache:             clients.OS.MachineInventory().Cache(),
		machineClient:            clients.OS.MachineInventory(),
		machineRegistrationCache: clients.OS.MachineRegistration().Cache(),
		settingCache:             clients.Rancher.Setting().Cache(),
		clusterRegistrationToken: clients.Rancher.ClusterRegistrationToken().Cache(),
	}

	server.secretCache.AddIndexer(tokenHash, func(obj *corev1.Secret) ([]string, error) {
		if string(obj.Type) != tokenType {
			return nil, nil
		}
		if token := obj.Data[tokenKey]; len(token) > 0 {
			hash := sha256.Sum256(token)
			return []string{base64.StdEncoding.EncodeToString(hash[:])}, nil
		}
		return nil, nil
	})

	server.machineCache.AddIndexer(tokenHash, func(obj *v1.MachineInventory) ([]string, error) {
		if obj.Spec.TPMHash == "" {
			return nil, nil
		}
		hash := sha256.Sum256([]byte(obj.Spec.TPMHash))
		return []string{base64.StdEncoding.EncodeToString(hash[:])}, nil
	})

	server.machineRegistrationCache.AddIndexer(registrationTokenIndex, func(obj *v1.MachineRegistration) ([]string, error) {
		if obj.Status.RegistrationToken == "" {
			return nil, nil
		}
		return []string{
			obj.Status.RegistrationToken,
		}, nil
	})

	server.machineCache.AddIndexer(tpmHashIndex, func(obj *v1.MachineInventory) ([]string, error) {
		if obj.Spec.TPMHash == "" {
			return nil, nil
		}
		return []string{obj.Spec.TPMHash}, nil
	})

	return server
}

func (i *InventoryServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.URL.Path, "/registration/") {
		i.register(resp, req)
	} else if strings.HasSuffix(req.URL.Path, "/cacerts") {
		i.cacerts(resp, req)
	} else {
		i.handle(resp, req)
	}
}

func (i *InventoryServer) authMachine(resp http.ResponseWriter, req *http.Request, registerNamespace string) (*v1.MachineInventory, io.WriteCloser, error) {
	for _, auth := range i.authenticators {
		machine, cont, writer, err := auth.Authenticate(resp, req, registerNamespace)
		if err != nil {
			return nil, nil, err
		}
		if machine != nil || !cont {
			return machine, writer, nil
		}
	}
	return nil, nil, nil
}

func writeErr(writer io.Writer, resp http.ResponseWriter, err error) {
	message := "Unauthorized"
	if err != nil {
		message = err.Error()
	}
	if writer == nil {
		http.Error(resp, message, http.StatusUnauthorized)
	} else {
		writer.Write([]byte(message))
	}
}

func (i *InventoryServer) handle(resp http.ResponseWriter, req *http.Request) {
	machine, writer, err := i.authMachine(resp, req, "")
	if machine == nil || err != nil {
		writeErr(writer, resp, err)
		return
	}
	defer writer.Close()

	if machine.Spec.ClusterName == "" {
		writeErr(writer, resp, errors.New("cluster not assigned"))
		return
	}
	crt, err := i.clusterRegistrationToken.Get(machine.Status.ClusterRegistrationTokenNamespace,
		machine.Status.ClusterRegistrationTokenName)
	if apierrors.IsNotFound(err) || crt.Status.Token == "" {
		writeErr(writer, resp, errors.New("cluster token not assigned"))
	}

	if err := writeResponse(writer, machine, crt); err != nil {
		writeErr(writer, resp, err)
	}
}

type config struct {
	Role            string   `json:"role,omitempty"`
	NodeName        string   `json:"nodeName,omitempty"`
	Address         string   `json:"address,omitempty"`
	InternalAddress string   `json:"internalAddress,omitempty"`
	Taints          []string `json:"taints,omitempty"`
	Labels          []string `json:"labels,omitempty"`
	Token           string   `json:"token,omitempty"`
}

func writeResponse(writer io.Writer, inventory *v1.MachineInventory, crt *v3.ClusterRegistrationToken) error {
	config := config{
		Role:            inventory.Spec.Config.Role,
		NodeName:        inventory.Spec.Config.NodeName,
		Address:         inventory.Spec.Config.Address,
		InternalAddress: inventory.Spec.Config.InternalAddress,
		Taints:          nil,
		Labels:          nil,
		Token:           crt.Status.Token,
	}
	for k, v := range inventory.Spec.Config.Labels {
		config.Labels = append(config.Labels, fmt.Sprintf("%s=%s", k, v))
	}
	for _, taint := range inventory.Spec.Config.Taints {
		config.Labels = append(config.Labels, taint.ToString())
	}
	return json.NewEncoder(writer).Encode(config)
}

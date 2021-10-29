package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	v1 "github.com/rancher/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/os2/pkg/clients"
	roscontrollers "github.com/rancher/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	corecontrollers "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type sharedSecretAuth struct {
	secretCache  corecontrollers.SecretCache
	machineCache roscontrollers.MachineInventoryCache
}

func newSharedSecretAuth(clients *clients.Clients) *sharedSecretAuth {
	server := &sharedSecretAuth{
		secretCache:  clients.Core.Secret().Cache(),
		machineCache: clients.OS.MachineInventory().Cache(),
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

	return server
}

func (s *sharedSecretAuth) Authenticate(resp http.ResponseWriter, req *http.Request, registerNamespace string) (*v1.MachineInventory, bool, io.WriteCloser, error) {
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	if token == "" || registerNamespace != "" {
		return nil, true, nil, nil
	}

	secrets, err := s.secretCache.GetByIndex(tokenIndex, token)
	if apierrors.IsNotFound(err) || len(secrets) == 0 {
		return nil, false, nil, fmt.Errorf("token not found")
	} else if err != nil {
		return nil, false, nil, err
	} else if len(secrets) > 1 {
		logrus.Errorf("Multiple machine secrets with the same value [%s/%s, %s/%s, ...]",
			secrets[0].Namespace, secrets[0].Name, secrets[1].Namespace, secrets[1].Name)
		return nil, false, nil, fmt.Errorf("token not found")
	}

	machines, err := s.machineCache.GetByIndex(machineBySecretNameIndex, secrets[0].Namespace+"/"+secrets[0].Name)
	if len(machines) > 1 {
		logrus.Errorf("Multiple machine inventories with the token: %v", machines)
	}
	if apierrors.IsNotFound(err) || len(machines) != 1 {
		return nil, false, nil, fmt.Errorf("machine not found")
	} else if err != nil {
		return nil, false, nil, err
	}
	return machines[0], false, writeCloser{resp}, nil
}

type writeCloser struct {
	io.Writer
}

func (writeCloser) Close() error {
	return nil
}

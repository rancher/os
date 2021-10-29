package tpm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/certificate-transparency-go/x509"
	"github.com/google/go-attestation/attest"
	"github.com/gorilla/websocket"
	v1 "github.com/rancher/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/merr"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *AuthServer) verifyChain(ek *attest.EK, namespace string) error {
	secret, err := a.secretCache.Get(namespace, tpmCACert)
	if apierrors.IsNotFound(err) {
		return nil
	}

	roots := x509.NewCertPool()
	_ = roots.AppendCertsFromPEM(secret.Data[corev1.TLSCertKey])
	opts := x509.VerifyOptions{
		Roots: roots,
	}
	_, err = ek.Certificate.Verify(opts)
	return err
}

func (a *AuthServer) generateChallenge(ek *attest.EK, attestationData *AttestationData) ([]byte, []byte, error) {
	ap := attest.ActivationParameters{
		TPMVersion: attest.TPMVersion20,
		EK:         ek.Public,
		AK:         *attestationData.AK,
	}

	secret, ec, err := ap.Generate()
	if err != nil {
		return nil, nil, fmt.Errorf("generating challenge: %w", err)
	}

	challengeBytes, err := json.Marshal(Challenge{EC: ec})
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling challenge: %w", err)
	}

	return secret, challengeBytes, nil
}

func (a *AuthServer) validateChallenge(secret, resp []byte) error {
	var response ChallengeResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return fmt.Errorf("unmarshalling challenge response: %w", err)
	}
	if !bytes.Equal(secret, response.Secret) {
		return fmt.Errorf("invalid challenge response")
	}
	return nil
}

func (a *AuthServer) validHash(ek *attest.EK, registerNamespace string) (*v1.MachineInventory, error) {
	hashEncoded, err := GetPubHash(ek)
	if err != nil {
		return nil, fmt.Errorf("tpm: could not get public key hash: %v", err)
	}

	if registerNamespace != "" {
		if err := a.verifyChain(ek, registerNamespace); err != nil {
			return nil, fmt.Errorf("verifying chain: %w", err)
		}
		return &v1.MachineInventory{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: registerNamespace,
			},
			Spec: v1.MachineInventorySpec{
				TPMHash: hashEncoded,
			},
		}, nil
	}

	machines, err := a.machineCache.GetByIndex(machineByHash, hashEncoded)
	if apierrors.IsNotFound(err) || len(machines) != 1 {
		if len(machines) > 1 {
			logrus.Errorf("multiple machines for same hash %s found: %v", hashEncoded, machines)
		}
		return nil, fmt.Errorf("failed to find machine")
	}

	if err := a.verifyChain(ek, machines[0].Namespace); err != nil {
		return nil, fmt.Errorf("verifying chain: %w", err)
	}

	return machines[0], nil
}

func writeRead(conn *websocket.Conn, input []byte) ([]byte, error) {
	writer, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(input); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	_, reader, err := conn.NextReader()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(reader)
}

func upgrade(resp http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		return nil, err
	}
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	return conn, err
}

func (a *AuthServer) getAttestationData(header string) (*attest.EK, *AttestationData, error) {
	tpmBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, "Bearer TPM"))
	if err != nil {
		return nil, nil, err
	}

	var attestationData AttestationData
	if err := json.Unmarshal(tpmBytes, &attestationData); err != nil {
		return nil, nil, err
	}

	ek, err := DecodeEK(attestationData.EK)
	if err != nil {
		return nil, nil, err
	}

	return ek, &attestationData, nil
}

func (a *AuthServer) Authenticate(resp http.ResponseWriter, req *http.Request, registerNamespace string) (*v1.MachineInventory, bool, io.WriteCloser, error) {
	header := req.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer TPM") {
		return nil, true, nil, nil
	}

	ek, attestationData, err := a.getAttestationData(header)
	if err != nil {
		return nil, false, nil, err
	}

	machine, err := a.validHash(ek, registerNamespace)
	if err != nil {
		return nil, false, nil, err
	}

	secret, challenge, err := a.generateChallenge(ek, attestationData)
	if err != nil {
		return nil, false, nil, err
	}

	conn, err := upgrade(resp, req)
	if err != nil {
		return nil, false, nil, err
	}

	challResp, err := writeRead(conn, challenge)
	if err != nil {
		return nil, false, nil, err
	}

	if err := a.validateChallenge(secret, challResp); err != nil {
		return nil, false, nil, err
	}

	writer, err := conn.NextWriter(websocket.BinaryMessage)
	return machine, false, &responseWriter{
		WriteCloser: writer,
		conn:        conn,
	}, err
}

type responseWriter struct {
	io.WriteCloser
	conn *websocket.Conn
}

func (r *responseWriter) Close() error {
	err := r.WriteCloser.Close()
	err2 := r.conn.Close()
	return merr.NewErrors(err, err2)
}

package server

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"net/http"
	"strings"
)

var (
	tokenHash = "tokenByHash"
)

func (i *InventoryServer) cacerts(rw http.ResponseWriter, req *http.Request) {
	ca := i.cacert()

	rw.Header().Set("Content-Type", "text/plain")
	var bytes []byte
	if strings.TrimSpace(ca) != "" {
		if !strings.HasSuffix(ca, "\n") {
			ca += "\n"
		}
		bytes = []byte(ca)
	}

	nonce := req.Header.Get("X-Cattle-Nonce")
	authorization := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")

	if authorization != "" && nonce != "" {
		crt, err := i.secretCache.GetByIndex(tokenHash, authorization)
		if err == nil && len(crt) >= 0 {
			digest := hmac.New(sha512.New, crt[0].Data[tokenKey])
			digest.Write([]byte(nonce))
			digest.Write([]byte{0})
			digest.Write(bytes)
			digest.Write([]byte{0})
			hash := digest.Sum(nil)
			rw.Header().Set("X-Cattle-Hash", base64.StdEncoding.EncodeToString(hash))
		}
	}

	if len(bytes) > 0 {
		_, _ = rw.Write([]byte(ca))
	}
}

func (i *InventoryServer) cacert() string {
	setting, err := i.settingCache.Get("cacerts")
	if err != nil {
		return ""
	}
	return setting.Value
}

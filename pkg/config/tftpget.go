package config

import (
	"bytes"
	"fmt"
	"net"
	"net/url"

	"gopkg.in/pin/tftp.v2"
	"sigs.k8s.io/yaml"
)

func tftpGet(tftpURL string) (map[string]interface{}, error) {
	u, err := url.Parse(tftpURL)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host + ":69"
	}

	fmt.Printf("Downloading config from host %s, file %s\n", host, u.Path)
	client, err := tftp.NewClient(host)
	if err != nil {
		return nil, err
	}
	writerTo, err := client.Receive(u.Path, "octet")
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if _, err := writerTo.WriteTo(buf); err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	return result, yaml.Unmarshal(buf.Bytes(), &result)
}

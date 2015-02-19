package config

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
)

func DecodeBase64Content(content string) ([]byte, error) {
	output, err := base64.StdEncoding.DecodeString(content)

	if err != nil {
		return nil, fmt.Errorf("Unable to decode base64: %q", err)
	}

	return output, nil
}

func DecodeGzipContent(content string) ([]byte, error) {
	gzr, err := gzip.NewReader(bytes.NewReader([]byte(content)))

	if err != nil {
		return nil, fmt.Errorf("Unable to decode gzip: %q", err)
	}
	defer gzr.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(gzr)

	return buf.Bytes(), nil
}

func DecodeContent(content string, encoding string) ([]byte, error) {
	switch encoding {
	case "":
		return []byte(content), nil

	case "b64", "base64":
		return DecodeBase64Content(content)

	case "gz", "gzip":
		return DecodeGzipContent(content)

	case "gz+base64", "gzip+base64", "gz+b64", "gzip+b64":
		gz, err := DecodeBase64Content(content)

		if err != nil {
			return nil, err
		}

		return DecodeGzipContent(string(gz))
	}

	return nil, fmt.Errorf("Unsupported encoding %q", encoding)
}

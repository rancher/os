// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
)

const (
	HTTP_2xx = 2
	HTTP_4xx = 4
)

type Err error

type ErrTimeout struct {
	Err
}

type ErrNotFound struct {
	Err
}

type ErrInvalid struct {
	Err
}

type ErrServer struct {
	Err
}

type ErrNetwork struct {
	Err
}

type HttpClient struct {
	// Maximum exp backoff duration. Defaults to 5 seconds
	MaxBackoff time.Duration

	// Maximum number of connection retries. Defaults to 15
	MaxRetries int

	// HTTP client timeout, this is suggested to be low since exponential
	// backoff will kick off too. Defaults to 2 seconds
	Timeout time.Duration

	// Whether or not to skip TLS verification. Defaults to false
	SkipTLS bool

	client *http.Client
}

type Getter interface {
	Get(string) ([]byte, error)
	GetRetry(string) ([]byte, error)
}

func NewHttpClient() *HttpClient {
	hc := &HttpClient{
		MaxBackoff: time.Second * 5,
		MaxRetries: 15,
		Timeout:    time.Duration(2) * time.Second,
		SkipTLS:    false,
	}

	// We need to create our own client in order to add timeout support.
	// TODO(c4milo) Replace it once Go 1.3 is officially used by CoreOS
	// More info: https://code.google.com/p/go/source/detail?r=ada6f2d5f99f
	hc.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: hc.SkipTLS,
			},
			Dial: func(network, addr string) (net.Conn, error) {
				deadline := time.Now().Add(hc.Timeout)
				c, err := net.DialTimeout(network, addr, hc.Timeout)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	return hc
}

func ExpBackoff(interval, max time.Duration) time.Duration {
	interval = interval * 2
	if interval > max {
		interval = max
	}
	return interval
}

// GetRetry fetches a given URL with support for exponential backoff and maximum retries
func (h *HttpClient) GetRetry(rawurl string) ([]byte, error) {
	if rawurl == "" {
		return nil, ErrInvalid{errors.New("URL is empty. Skipping.")}
	}

	url, err := neturl.Parse(rawurl)
	if err != nil {
		return nil, ErrInvalid{err}
	}

	// Unfortunately, url.Parse is too generic to throw errors if a URL does not
	// have a valid HTTP scheme. So, we have to do this extra validation
	if !strings.HasPrefix(url.Scheme, "http") {
		return nil, ErrInvalid{fmt.Errorf("URL %s does not have a valid HTTP scheme. Skipping.", rawurl)}
	}

	dataURL := url.String()

	duration := 50 * time.Millisecond
	for retry := 1; retry <= h.MaxRetries; retry++ {
		log.Printf("Fetching data from %s. Attempt #%d", dataURL, retry)

		data, err := h.Get(dataURL)
		switch err.(type) {
		case ErrNetwork:
			log.Printf(err.Error())
		case ErrServer:
			log.Printf(err.Error())
		case ErrNotFound:
			return data, err
		default:
			return data, err
		}

		duration = ExpBackoff(duration, h.MaxBackoff)
		log.Printf("Sleeping for %v...", duration)
		time.Sleep(duration)
	}

	return nil, ErrTimeout{fmt.Errorf("Unable to fetch data. Maximum retries reached: %d", h.MaxRetries)}
}

func (h *HttpClient) Get(dataURL string) ([]byte, error) {
	if resp, err := h.client.Get(dataURL); err == nil {
		defer resp.Body.Close()
		switch resp.StatusCode / 100 {
		case HTTP_2xx:
			return ioutil.ReadAll(resp.Body)
		case HTTP_4xx:
			return nil, ErrNotFound{fmt.Errorf("Not found. HTTP status code: %d", resp.StatusCode)}
		default:
			return nil, ErrServer{fmt.Errorf("Server error. HTTP status code: %d", resp.StatusCode)}
		}
	} else {
		return nil, ErrNetwork{fmt.Errorf("Unable to fetch data: %s", err.Error())}
	}
}

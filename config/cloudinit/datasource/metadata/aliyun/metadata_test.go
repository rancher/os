package aliyun

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/burmilla/os/config/cloudinit/datasource"
	"github.com/burmilla/os/config/cloudinit/datasource/metadata"
	"github.com/burmilla/os/config/cloudinit/datasource/metadata/test"
	"github.com/burmilla/os/config/cloudinit/pkg"
)

func TestType(t *testing.T) {
	want := "aliyun-metadata-service"
	if kind := (MetadataService{}).Type(); kind != want {
		t.Fatalf("bad type: want %q, got %q", want, kind)
	}
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root         string
		metadataPath string
		resources    map[string]string
		expect       datasource.Metadata
		clientErr    error
		expectErr    error
	}{
		{
			root:         "/",
			metadataPath: "2016-01-01/meta-data/",
			resources: map[string]string{
				"/2016-01-01/meta-data/": "hostname\n",
			},
			expectErr: fmt.Errorf("The public-keys should be enable in aliyun-metadata-service"),
		},
		{
			root:         "/",
			metadataPath: "2016-01-01/meta-data/",
			resources: map[string]string{
				"/2016-01-01/meta-data/":                           "hostname\npublic-keys/\n",
				"/2016-01-01/meta-data/hostname":                   "host",
				"/2016-01-01/meta-data/public-keys/":               "xx/",
				"/2016-01-01/meta-data/public-keys/xx/":            "openssh-key",
				"/2016-01-01/meta-data/public-keys/xx/openssh-key": "key",
			},
			expect: datasource.Metadata{
				Hostname:      "host",
				SSHPublicKeys: map[string]string{"xx": "key"},
			},
		},
		{
			clientErr: pkg.ErrTimeout{Err: fmt.Errorf("test error")},
			expectErr: pkg.ErrTimeout{Err: fmt.Errorf("test error")},
		},
	} {
		service := &MetadataService{metadata.Service{
			Root:         tt.root,
			Client:       &test.HTTPClient{Resources: tt.resources, Err: tt.clientErr},
			MetadataPath: tt.metadataPath,
		}}
		metadata, err := service.FetchMetadata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): \nwant %q, \ngot %q\n", tt.resources, tt.expectErr, err)
		}
		if !reflect.DeepEqual(tt.expect, metadata) {
			t.Fatalf("bad fetch (%q): \nwant %#v, \ngot %#v\n", tt.resources, tt.expect, metadata)
		}
	}
}

func Error(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

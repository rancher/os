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

package config

type Etcd2 struct {
	AdvertiseClientURLs      string `yaml:"advertise_client_urls"         env:"ETCD_ADVERTISE_CLIENT_URLS"`
	CAFile                   string `yaml:"ca_file"                       env:"ETCD_CA_FILE"`
	CertFile                 string `yaml:"cert_file"                     env:"ETCD_CERT_FILE"`
	CorsOrigins              string `yaml:"cors"                          env:"ETCD_CORS"`
	DataDir                  string `yaml:"data_dir"                      env:"ETCD_DATA_DIR"`
	Discovery                string `yaml:"discovery"                     env:"ETCD_DISCOVERY"`
	DiscoveryFallback        string `yaml:"discovery_fallback"            env:"ETCD_DISCOVERY_FALLBACK"`
	DiscoverySRV             string `yaml:"discovery_srv"                 env:"ETCD_DISCOVERY_SRV"`
	DiscoveryProxy           string `yaml:"discovery_proxy"               env:"ETCD_DISCOVERY_PROXY"`
	ElectionTimeout          int    `yaml:"election_timeout"              env:"ETCD_ELECTION_TIMEOUT"`
	HeartbeatInterval        int    `yaml:"heartbeat_interval"            env:"ETCD_HEARTBEAT_INTERVAL"`
	InitialAdvertisePeerURLs string `yaml:"initial_advertise_peer_urls"   env:"ETCD_INITIAL_ADVERTISE_PEER_URLS"`
	InitialCluster           string `yaml:"initial_cluster"               env:"ETCD_INITIAL_CLUSTER"`
	InitialClusterState      string `yaml:"initial_cluster_state"         env:"ETCD_INITIAL_CLUSTER_STATE"`
	InitialClusterToken      string `yaml:"initial_cluster_token"         env:"ETCD_INITIAL_CLUSTER_TOKEN"`
	KeyFile                  string `yaml:"key_file"                      env:"ETCD_KEY_FILE"`
	ListenClientURLs         string `yaml:"listen_client_urls"            env:"ETCD_LISTEN_CLIENT_URLS"`
	ListenPeerURLs           string `yaml:"listen_peer_urls"              env:"ETCD_LISTEN_PEER_URLS"`
	MaxSnapshots             int    `yaml:"max_snapshots"                 env:"ETCD_MAX_SNAPSHOTS"`
	MaxWALs                  int    `yaml:"max_wals"                      env:"ETCD_MAX_WALS"`
	Name                     string `yaml:"name"                          env:"ETCD_NAME"`
	PeerCAFile               string `yaml:"peer_ca_file"                  env:"ETCD_PEER_CA_FILE"`
	PeerCertFile             string `yaml:"peer_cert_file"                env:"ETCD_PEER_CERT_FILE"`
	PeerKeyFile              string `yaml:"peer_key_file"                 env:"ETCD_PEER_KEY_FILE"`
	Proxy                    string `yaml:"proxy"                         env:"ETCD_PROXY"`
	SnapshotCount            int    `yaml:"snapshot_count"                env:"ETCD_SNAPSHOTCOUNT"`
}

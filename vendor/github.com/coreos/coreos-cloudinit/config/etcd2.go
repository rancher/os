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
	CAFile                   string `yaml:"ca_file"                       env:"ETCD_CA_FILE"                     deprecated:"ca_file obsoleted by trusted_ca_file and client_cert_auth"`
	CertFile                 string `yaml:"cert_file"                     env:"ETCD_CERT_FILE"`
	ClientCertAuth           bool   `yaml:"client_cert_auth"              env:"ETCD_CLIENT_CERT_AUTH"`
	CorsOrigins              string `yaml:"cors"                          env:"ETCD_CORS"`
	DataDir                  string `yaml:"data_dir"                      env:"ETCD_DATA_DIR"`
	Debug                    bool   `yaml:"debug"                         env:"ETCD_DEBUG"`
	Discovery                string `yaml:"discovery"                     env:"ETCD_DISCOVERY"`
	DiscoveryFallback        string `yaml:"discovery_fallback"            env:"ETCD_DISCOVERY_FALLBACK"`
	DiscoverySRV             string `yaml:"discovery_srv"                 env:"ETCD_DISCOVERY_SRV"`
	DiscoveryProxy           string `yaml:"discovery_proxy"               env:"ETCD_DISCOVERY_PROXY"`
	ElectionTimeout          int    `yaml:"election_timeout"              env:"ETCD_ELECTION_TIMEOUT"`
	EnablePprof              bool   `yaml:"enable_pprof"                  env:"ETCD_ENABLE_PPROF"`
	ForceNewCluster          bool   `yaml:"force_new_cluster"             env:"ETCD_FORCE_NEW_CLUSTER"`
	HeartbeatInterval        int    `yaml:"heartbeat_interval"            env:"ETCD_HEARTBEAT_INTERVAL"`
	InitialAdvertisePeerURLs string `yaml:"initial_advertise_peer_urls"   env:"ETCD_INITIAL_ADVERTISE_PEER_URLS"`
	InitialCluster           string `yaml:"initial_cluster"               env:"ETCD_INITIAL_CLUSTER"`
	InitialClusterState      string `yaml:"initial_cluster_state"         env:"ETCD_INITIAL_CLUSTER_STATE"`
	InitialClusterToken      string `yaml:"initial_cluster_token"         env:"ETCD_INITIAL_CLUSTER_TOKEN"`
	KeyFile                  string `yaml:"key_file"                      env:"ETCD_KEY_FILE"`
	ListenClientURLs         string `yaml:"listen_client_urls"            env:"ETCD_LISTEN_CLIENT_URLS"`
	ListenPeerURLs           string `yaml:"listen_peer_urls"              env:"ETCD_LISTEN_PEER_URLS"`
	LogPackageLevels         string `yaml:"log_package_levels"            env:"ETCD_LOG_PACKAGE_LEVELS"`
	MaxSnapshots             int    `yaml:"max_snapshots"                 env:"ETCD_MAX_SNAPSHOTS"`
	MaxWALs                  int    `yaml:"max_wals"                      env:"ETCD_MAX_WALS"`
	Name                     string `yaml:"name"                          env:"ETCD_NAME"`
	PeerCAFile               string `yaml:"peer_ca_file"                  env:"ETCD_PEER_CA_FILE"                deprecated:"peer_ca_file obsoleted peer_trusted_ca_file and peer_client_cert_auth"`
	PeerCertFile             string `yaml:"peer_cert_file"                env:"ETCD_PEER_CERT_FILE"`
	PeerKeyFile              string `yaml:"peer_key_file"                 env:"ETCD_PEER_KEY_FILE"`
	PeerClientCertAuth       bool   `yaml:"peer_client_cert_auth"         env:"ETCD_PEER_CLIENT_CERT_AUTH"`
	PeerTrustedCAFile        string `yaml:"peer_trusted_ca_file"          env:"ETCD_PEER_TRUSTED_CA_FILE"`
	Proxy                    string `yaml:"proxy"                         env:"ETCD_PROXY"                       valid:"^(on|off|readonly)$"`
	ProxyDialTimeout         int    `yaml:"proxy_dial_timeout"            env:"ETCD_PROXY_DIAL_TIMEOUT"`
	ProxyFailureWait         int    `yaml:"proxy_failure_wait"            env:"ETCD_PROXY_FAILURE_WAIT"`
	ProxyReadTimeout         int    `yaml:"proxy_read_timeout"            env:"ETCD_PROXY_READ_TIMEOUT"`
	ProxyRefreshInterval     int    `yaml:"proxy_refresh_interval"        env:"ETCD_PROXY_REFRESH_INTERVAL"`
	ProxyWriteTimeout        int    `yaml:"proxy_write_timeout"           env:"ETCD_PROXY_WRITE_TIMEOUT"`
	SnapshotCount            int    `yaml:"snapshot_count"                env:"ETCD_SNAPSHOT_COUNT"`
	StrictReconfigCheck      bool   `yaml:"strict_reconfig_check"         env:"ETCD_STRICT_RECONFIG_CHECK"`
	TrustedCAFile            string `yaml:"trusted_ca_file"               env:"ETCD_TRUSTED_CA_FILE"`
	WalDir                   string `yaml:"wal_dir"                       env:"ETCD_WAL_DIR"`
}

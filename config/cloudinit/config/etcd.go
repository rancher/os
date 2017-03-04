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

type Etcd struct {
	Addr                     string  `yaml:"addr"                          env:"ETCD_ADDR"`
	AdvertiseClientURLs      string  `yaml:"advertise_client_urls"         env:"ETCD_ADVERTISE_CLIENT_URLS"       deprecated:"etcd2 options no longer work for etcd"`
	BindAddr                 string  `yaml:"bind_addr"                     env:"ETCD_BIND_ADDR"`
	CAFile                   string  `yaml:"ca_file"                       env:"ETCD_CA_FILE"`
	CertFile                 string  `yaml:"cert_file"                     env:"ETCD_CERT_FILE"`
	ClusterActiveSize        int     `yaml:"cluster_active_size"           env:"ETCD_CLUSTER_ACTIVE_SIZE"`
	ClusterRemoveDelay       float64 `yaml:"cluster_remove_delay"          env:"ETCD_CLUSTER_REMOVE_DELAY"`
	ClusterSyncInterval      float64 `yaml:"cluster_sync_interval"         env:"ETCD_CLUSTER_SYNC_INTERVAL"`
	CorsOrigins              string  `yaml:"cors"                          env:"ETCD_CORS"`
	DataDir                  string  `yaml:"data_dir"                      env:"ETCD_DATA_DIR"`
	Discovery                string  `yaml:"discovery"                     env:"ETCD_DISCOVERY"`
	DiscoveryFallback        string  `yaml:"discovery_fallback"            env:"ETCD_DISCOVERY_FALLBACK"          deprecated:"etcd2 options no longer work for etcd"`
	DiscoverySRV             string  `yaml:"discovery_srv"                 env:"ETCD_DISCOVERY_SRV"               deprecated:"etcd2 options no longer work for etcd"`
	DiscoveryProxy           string  `yaml:"discovery_proxy"               env:"ETCD_DISCOVERY_PROXY"             deprecated:"etcd2 options no longer work for etcd"`
	ElectionTimeout          int     `yaml:"election_timeout"              env:"ETCD_ELECTION_TIMEOUT"            deprecated:"etcd2 options no longer work for etcd"`
	ForceNewCluster          bool    `yaml:"force_new_cluster"             env:"ETCD_FORCE_NEW_CLUSTER"           deprecated:"etcd2 options no longer work for etcd"`
	GraphiteHost             string  `yaml:"graphite_host"                 env:"ETCD_GRAPHITE_HOST"`
	HeartbeatInterval        int     `yaml:"heartbeat_interval"            env:"ETCD_HEARTBEAT_INTERVAL"          deprecated:"etcd2 options no longer work for etcd"`
	HTTPReadTimeout          float64 `yaml:"http_read_timeout"             env:"ETCD_HTTP_READ_TIMEOUT"`
	HTTPWriteTimeout         float64 `yaml:"http_write_timeout"            env:"ETCD_HTTP_WRITE_TIMEOUT"`
	InitialAdvertisePeerURLs string  `yaml:"initial_advertise_peer_urls"   env:"ETCD_INITIAL_ADVERTISE_PEER_URLS" deprecated:"etcd2 options no longer work for etcd"`
	InitialCluster           string  `yaml:"initial_cluster"               env:"ETCD_INITIAL_CLUSTER"             deprecated:"etcd2 options no longer work for etcd"`
	InitialClusterState      string  `yaml:"initial_cluster_state"         env:"ETCD_INITIAL_CLUSTER_STATE"       deprecated:"etcd2 options no longer work for etcd"`
	InitialClusterToken      string  `yaml:"initial_cluster_token"         env:"ETCD_INITIAL_CLUSTER_TOKEN"       deprecated:"etcd2 options no longer work for etcd"`
	KeyFile                  string  `yaml:"key_file"                      env:"ETCD_KEY_FILE"`
	ListenClientURLs         string  `yaml:"listen_client_urls"            env:"ETCD_LISTEN_CLIENT_URLS"          deprecated:"etcd2 options no longer work for etcd"`
	ListenPeerURLs           string  `yaml:"listen_peer_urls"              env:"ETCD_LISTEN_PEER_URLS"            deprecated:"etcd2 options no longer work for etcd"`
	MaxResultBuffer          int     `yaml:"max_result_buffer"             env:"ETCD_MAX_RESULT_BUFFER"`
	MaxRetryAttempts         int     `yaml:"max_retry_attempts"            env:"ETCD_MAX_RETRY_ATTEMPTS"`
	MaxSnapshots             int     `yaml:"max_snapshots"                 env:"ETCD_MAX_SNAPSHOTS"               deprecated:"etcd2 options no longer work for etcd"`
	MaxWALs                  int     `yaml:"max_wals"                      env:"ETCD_MAX_WALS"                    deprecated:"etcd2 options no longer work for etcd"`
	Name                     string  `yaml:"name"                          env:"ETCD_NAME"`
	PeerAddr                 string  `yaml:"peer_addr"                     env:"ETCD_PEER_ADDR"`
	PeerBindAddr             string  `yaml:"peer_bind_addr"                env:"ETCD_PEER_BIND_ADDR"`
	PeerCAFile               string  `yaml:"peer_ca_file"                  env:"ETCD_PEER_CA_FILE"`
	PeerCertFile             string  `yaml:"peer_cert_file"                env:"ETCD_PEER_CERT_FILE"`
	PeerElectionTimeout      int     `yaml:"peer_election_timeout"         env:"ETCD_PEER_ELECTION_TIMEOUT"`
	PeerHeartbeatInterval    int     `yaml:"peer_heartbeat_interval"       env:"ETCD_PEER_HEARTBEAT_INTERVAL"`
	PeerKeyFile              string  `yaml:"peer_key_file"                 env:"ETCD_PEER_KEY_FILE"`
	Peers                    string  `yaml:"peers"                         env:"ETCD_PEERS"`
	PeersFile                string  `yaml:"peers_file"                    env:"ETCD_PEERS_FILE"`
	Proxy                    string  `yaml:"proxy"                         env:"ETCD_PROXY"                       deprecated:"etcd2 options no longer work for etcd"`
	RetryInterval            float64 `yaml:"retry_interval"                env:"ETCD_RETRY_INTERVAL"`
	Snapshot                 bool    `yaml:"snapshot"                      env:"ETCD_SNAPSHOT"`
	SnapshotCount            int     `yaml:"snapshot_count"                env:"ETCD_SNAPSHOTCOUNT"`
	StrTrace                 string  `yaml:"trace"                         env:"ETCD_TRACE"`
	Verbose                  bool    `yaml:"verbose"                       env:"ETCD_VERBOSE"`
	VeryVerbose              bool    `yaml:"very_verbose"                  env:"ETCD_VERY_VERBOSE"`
	VeryVeryVerbose          bool    `yaml:"very_very_verbose"             env:"ETCD_VERY_VERY_VERBOSE"`
}

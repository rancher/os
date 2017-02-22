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

type Fleet struct {
	AgentTTL                string  `yaml:"agent_ttl"                 env:"FLEET_AGENT_TTL"`
	AuthorizedKeysFile      string  `yaml:"authorized_keys_file"      env:"FLEET_AUTHORIZED_KEYS_FILE"`
	DisableEngine           bool    `yaml:"disable_engine"            env:"FLEET_DISABLE_ENGINE"`
	EngineReconcileInterval float64 `yaml:"engine_reconcile_interval" env:"FLEET_ENGINE_RECONCILE_INTERVAL"`
	EtcdCAFile              string  `yaml:"etcd_cafile"               env:"FLEET_ETCD_CAFILE"`
	EtcdCertFile            string  `yaml:"etcd_certfile"             env:"FLEET_ETCD_CERTFILE"`
	EtcdKeyFile             string  `yaml:"etcd_keyfile"              env:"FLEET_ETCD_KEYFILE"`
	EtcdKeyPrefix           string  `yaml:"etcd_key_prefix"           env:"FLEET_ETCD_KEY_PREFIX"`
	EtcdRequestTimeout      float64 `yaml:"etcd_request_timeout"      env:"FLEET_ETCD_REQUEST_TIMEOUT"`
	EtcdServers             string  `yaml:"etcd_servers"              env:"FLEET_ETCD_SERVERS"`
	EtcdUsername            string  `yaml:"etcd_username"             env:"FLEET_ETCD_USERNAME"`
	EtcdPassword            string  `yaml:"etcd_password"             env:"FLEET_ETCD_PASSWORD"`
	Metadata                string  `yaml:"metadata"                  env:"FLEET_METADATA"`
	PublicIP                string  `yaml:"public_ip"                 env:"FLEET_PUBLIC_IP"`
	TokenLimit              int     `yaml:"token_limit"               env:"FLEET_TOKEN_LIMIT"`
	Verbosity               int     `yaml:"verbosity"                 env:"FLEET_VERBOSITY"`
	VerifyUnits             bool    `yaml:"verify_units"              env:"FLEET_VERIFY_UNITS"`
}

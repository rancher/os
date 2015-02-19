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

type Update struct {
	RebootStrategy string `yaml:"reboot_strategy" env:"REBOOT_STRATEGY" valid:"^(best-effort|etcd-lock|reboot|off)$"`
	Group          string `yaml:"group"           env:"GROUP"`
	Server         string `yaml:"server"          env:"SERVER"`
}

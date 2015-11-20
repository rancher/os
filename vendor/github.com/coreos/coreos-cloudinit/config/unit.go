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

type Unit struct {
	Name    string       `yaml:"name"`
	Mask    bool         `yaml:"mask"`
	Enable  bool         `yaml:"enable"`
	Runtime bool         `yaml:"runtime"`
	Content string       `yaml:"content"`
	Command string       `yaml:"command" valid:"^(start|stop|restart|reload|try-restart|reload-or-restart|reload-or-try-restart)$"`
	DropIns []UnitDropIn `yaml:"drop_ins"`
}

type UnitDropIn struct {
	Name    string `yaml:"name"`
	Content string `yaml:"content"`
}

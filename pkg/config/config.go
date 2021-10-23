package config

type RancherOS struct {
	Install Install `json:"install,omitempty"`
}

type Install struct {
	Automatic bool   `json:"automatic,omitempty"`
	ForceEFI  bool   `json:"forceEfi,omitempty"`
	Device    string `json:"device,omitempty"`
	ConfigURL string `json:"configUrl,omitempty"`
	Silent    bool   `json:"silent,omitempty"`
	ISOURL    string `json:"isoUrl,omitempty"`
	PowerOff  bool   `json:"powerOff,omitempty"`
	NoFormat  bool   `json:"noFormat,omitempty"`
	Debug     bool   `json:"debug,omitempty"`
	TTY       string `json:"tty,omitempty"`
	ServerURL string `json:"-"`
	Token     string `json:"-"`
	Role      string `json:"-"`
	Password  string `json:"password,omitempty"`
}

type Config struct {
	SSHAuthorizedKeys []string  `json:"sshAuthorizedKeys,omitempty"`
	RancherOS         RancherOS `json:"rancheros,omitempty"`
}

type YipConfig struct {
	Stages   map[string][]Stage `json:"stages,omitempty"`
	Rancherd Rancherd           `json:"rancherd,omitempty"`
}

type Stage struct {
	Users map[string]User `json:"users,omitempty"`
}

type Rancherd struct {
	Server string `json:"server,omitempty"`
	Role   string `json:"role,omitempty"`
	Token  string `json:"token,omitempty"`
}

type User struct {
	Name              string   `json:"name,omitempty"`
	PasswordHash      string   `json:"passwd,omitempty"`
	SSHAuthorizedKeys []string `json:"ssh_authorized_keys,omitempty"`
}

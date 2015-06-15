package config

func NewConfig() *Config {
	return ReadConfig(OsConfigFile)
}

func ReadConfig(file string) *Config {
	if data, err := readConfig(nil, file); err == nil {
		c := &Config{}
		c.merge(data)
		return c
	} else {
		return nil
	}
}

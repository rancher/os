package config

func NewConfig() *CloudConfig {
	return ReadConfig(OsConfigFile)
}

func ReadConfig(file string) *CloudConfig {
	if data, err := readConfig(nil, file); err == nil {
		c := &CloudConfig{}
		c.merge(data)
		c.amendNils()
		return c
	} else {
		return nil
	}
}

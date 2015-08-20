package config

func NewConfig() *CloudConfig {
	return ReadConfig(nil, OsConfigFile)
}

func ReadConfig(bytes []byte, files ...string) *CloudConfig {
	if data, err := readConfig(bytes, files...); err == nil {
		c := &CloudConfig{}
		if err := c.merge(data); err != nil {
			return nil
		}
		c.amendNils()
		return c
	} else {
		return nil
	}
}

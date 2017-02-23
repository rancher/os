package rpcvmx

import (
	"fmt"
	"strconv"

	"github.com/vmware/vmw-guestinfo/rpcout"
)

// Config gives access to the vmx config through the VMware backdoor
type Config struct{}

// NewConfig creates a new Config object
func NewConfig() *Config {
	return &Config{}
}

// String returns the config string in the guestinfo.* namespace
func (c *Config) String(key string, defaultValue string) (string, error) {
	out, ok, err := rpcout.SendOne("info-get guestinfo.%s", key)
	if err != nil {
		return "", err
	} else if !ok {
		return defaultValue, nil
	}
	return string(out), nil
}

// Bool returns the config boolean in the guestinfo.* namespace
func (c *Config) Bool(key string, defaultValue bool) (bool, error) {
	val, err := c.String(key, fmt.Sprintf("%t", defaultValue))
	if err != nil {
		return false, err
	}
	res, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue, nil
	}
	return res, nil
}

// Int returns the config integer in the guestinfo.* namespace
func (c *Config) Int(key string, defaultValue int) (int, error) {
	val, err := c.String(key, "")
	if err != nil {
		return 0, err
	}
	res, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue, nil
	}
	return res, nil
}

// SetString sets the guestinfo.KEY with the string VALUE
func (c *Config) SetString(key string, value string) error {
	_, _, err := rpcout.SendOne("info-set guestinfo.%s %s", key, value)
	if err != nil {
		return err
	}
	return nil
}

// SetString sets the guestinfo.KEY with the bool VALUE
func (c *Config) SetBool(key string, value bool) error {
	return c.SetString(key, strconv.FormatBool(value))
}

// SetString sets the guestinfo.KEY with the int VALUE
func (c *Config) SetInt(key string, value int) error {
	return c.SetString(key, strconv.Itoa(value))
}

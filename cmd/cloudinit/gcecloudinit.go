package cloudinit

import (
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/cloud/compute/metadata"
	"gopkg.in/yaml.v2"
)

type GceCloudConfig struct {
	FileName           string
	UserData           string
	NonUserDataSSHKeys []string
}

const (
	gceCloudConfigFile = "/var/lib/rancher/conf/gce_cloudinit_config.yml"
)

func NewGceCloudConfig() *GceCloudConfig {

	userData, err := metadata.InstanceAttributeValue("user-data")
	if err != nil {
		log.Errorf("Could not retrieve user-data: %s", err)
	}

	projectSSHKeys, err := metadata.ProjectAttributeValue("sshKeys")
	if err != nil {
		log.Errorf("Could not retrieve project SSH Keys: %s", err)
	}

	instanceSSHKeys, err := metadata.InstanceAttributeValue("sshKeys")
	if err != nil {
		log.Errorf("Could not retrieve instance SSH Keys: %s", err)
	}

	nonUserDataSSHKeysRaw := projectSSHKeys + instanceSSHKeys
	nonUserDataSSHKeys := gceSshKeyFormatter(nonUserDataSSHKeysRaw)

	gceCC := &GceCloudConfig{
		FileName:           gceCloudConfigFile,
		UserData:           userData,
		NonUserDataSSHKeys: nonUserDataSSHKeys,
	}

	return gceCC
}

func GetAndCreateGceDataSourceFilename() (string, error) {
	gceCC := NewGceCloudConfig()
	err := gceCC.saveToFile(gceCC.FileName)
	if err != nil {
		log.Errorf("Error: %s", err)
		return "", err
	}
	return gceCC.FileName, nil
}

func (cc *GceCloudConfig) saveToFile(filename string) error {
	//Get Merged UserData sshkeys
	data, err := cc.getMergedUserData()
	if err != nil {
		log.Errorf("Could not process userdata: %s", err)
		return err
	}
	//write file
	writeFile(filename, data)
	return nil
}

func (cc *GceCloudConfig) getMergedUserData() ([]byte, error) {
	var returnUserData []byte
	userdata := make(map[string]interface{})

	if cc.UserData != "" {
		log.Infof("Found UserData Config")
		err := yaml.Unmarshal([]byte(cc.UserData), &userdata)
		if err != nil {
			log.Errorf("Could not unmarshal data: %s", err)
			return nil, err
		}
	}

	var auth_keys []string
	if _, exists := userdata["ssh_authorized_keys"]; exists {
		udSshKeys := userdata["ssh_authorized_keys"].([]interface{})
		log.Infof("userdata %s", udSshKeys)

		for _, value := range udSshKeys {
			auth_keys = append(auth_keys, value.(string))
		}
	}
	if cc.NonUserDataSSHKeys != nil {
		for _, value := range cc.NonUserDataSSHKeys {
			auth_keys = append(auth_keys, value)
		}
	}
	userdata["ssh_authorized_keys"] = auth_keys

	yamlUserData, err := yaml.Marshal(&userdata)
	if err != nil {
		log.Errorf("Could not Marshal userdata: %s", err)
		return nil, err
	} else {
		returnUserData = append([]byte("#cloud-config\n"), yamlUserData...)
	}

	return returnUserData, nil
}

func writeFile(filename string, data []byte) error {
	if err := ioutil.WriteFile(filename, data, 400); err != nil {
		log.Errorf("Could not write file %v", err)
		return err
	}
	return nil
}

func gceSshKeyFormatter(rawKeys string) []string {
	keySlice := strings.Split(rawKeys, "\n")
	var cloudFormatedKeys []string

	if len(keySlice) > 0 {
		for i := range keySlice {
			keyString := keySlice[i]
			sIdx := strings.Index(keyString, ":")
			if sIdx != -1 {
				key := strings.TrimSpace(keyString[sIdx+1:])
				keyA := strings.Split(key, " ")
				key = strings.Join(keyA, " ")
				if key != "" {
					cloudFormatedKeys = append(cloudFormatedKeys, key)
				}
			}
		}
	}
	return cloudFormatedKeys
}

package cloudinitexecute

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/os/util"
)

var (
	sshDir             = ".ssh"
	authorizedKeysFile = path.Join(sshDir, "authorized_keys")
)

func authorizeSSHKeys(username string, authorizedKeys []string, name string) error {
	var uid int
	var gid int
	var homeDir string

	bytes, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(bytes), "\n") {
		if strings.HasPrefix(line, username) {
			split := strings.Split(line, ":")
			if len(split) < 6 {
				break
			}
			uid, err = strconv.Atoi(split[2])
			if err != nil {
				return err
			}
			gid, err = strconv.Atoi(split[3])
			if err != nil {
				return err
			}
			homeDir = split[5]
		}
	}

	sshDir = path.Join(homeDir, sshDir)
	authorizedKeysFile = path.Join(homeDir, authorizedKeysFile)

	for _, authorizedKey := range authorizedKeys {
		if err = authorizeSSHKey(authorizedKey, uid, gid, homeDir); err != nil {
			log.Errorf("Failed to authorize SSH key %s: %v", authorizedKey, err)
		}
	}

	return nil
}

func authorizeSSHKey(authorizedKey string, uid, gid int, homeDir string) error {
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		if err = os.Mkdir(sshDir, 0700); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	authorizedKeysFileInfo, err := os.Stat(authorizedKeysFile)
	if os.IsNotExist(err) {
		keysFile, err := os.Create(authorizedKeysFile)
		if err != nil {
			return err
		}
		if err = keysFile.Chmod(0600); err != nil {
			return err
		}
		if err = keysFile.Close(); err != nil {
			return err
		}
		authorizedKeysFileInfo, err = os.Stat(authorizedKeysFile)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	bytes, err := ioutil.ReadFile(authorizedKeysFile)
	if err != nil {
		return err
	}

	if !strings.Contains(string(bytes), authorizedKey) {
		bytes = append(bytes, []byte(authorizedKey)...)
		bytes = append(bytes, '\n')
	}

	perm := authorizedKeysFileInfo.Mode().Perm()
	if err = util.WriteFileAtomic(authorizedKeysFile, bytes, perm); err != nil {
		return err
	}
	if err = os.Chown(sshDir, uid, gid); err != nil {
		return err
	}
	if err = os.Chown(authorizedKeysFile, uid, gid); err != nil {
		return err
	}

	return nil
}

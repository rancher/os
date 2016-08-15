package cloudinitexecute

import (
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

func authorizeSSHKeys(user string, authorizedKeys []string, name string) {
	for _, authorizedKey := range authorizedKeys {
		cmd := exec.Command("update-ssh-keys", user, authorizedKey)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.WithFields(log.Fields{"err": err, "user": user, "auth_key": authorizedKey}).Error("Error updating SSH authorized_keys")
		}
	}
}

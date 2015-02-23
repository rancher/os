package cloudinit

import(
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

func authorizeSSHKeys(user string, authorizedKeys []string, name string) {
	for _, authorizedKey := range authorizedKeys {
		cmd := exec.Command("update-ssh-keys", user, authorizedKey)
		err := cmd.Run() 
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

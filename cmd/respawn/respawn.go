package respawn

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

func Main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var wait sync.WaitGroup

	for _, line := range strings.Split(string(input), "\n") {
		wait.Add(1)
		go execute(line, wait)
	}

	wait.Wait()
}

func execute(line string, wait sync.WaitGroup) {
	start := time.Now()
	count := 0

	for {
		args := strings.Split(line, " ")

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err := cmd.Start()
		if err != nil {
			log.Error("%s : %v", line, err)
		}

		err = cmd.Wait()
		if err != nil {
			log.Error("%s : %v", line, err)
		}

		count++

		if count > 10 {
			if start.Sub(time.Now()) <= (1 * time.Second) {
				log.Errorf("%s : restarted too fast, not executing", line)
				break
			}

			count = 0
			start = time.Now()
		}
	}

	wait.Done()
}

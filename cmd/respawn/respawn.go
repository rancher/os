package respawn

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func Main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Optional config file to load",
		},
	}
	app.Action = run

	app.Run(os.Args)

}

func run(c *cli.Context) {
	var stream io.Reader = os.Stdin
	var err error

	inputFileName := c.String("file")

	if inputFileName != "" {
		stream, err = os.Open(inputFileName)
		if err != nil {
			log.Fatal(err)
		}
	}

	input, err := ioutil.ReadAll(stream)
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

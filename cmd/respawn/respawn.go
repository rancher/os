package respawn

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	running     bool                = true
	processes   map[int]*os.Process = map[int]*os.Process{}
	processLock                     = sync.Mutex{}
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

func setupSigterm() {
	sigtermChan := make(chan os.Signal)
	signal.Notify(sigtermChan, syscall.SIGTERM)
	go func() {
		for _ = range sigtermChan {
			termPids()
		}
	}()
}

func run(c *cli.Context) {
	setupSigterm()

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

	var wg sync.WaitGroup

	for _, line := range strings.Split(string(input), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		wg.Add(1)
		go execute(line, &wg)
	}

	wg.Wait()
}

func addProcess(process *os.Process) {
	processLock.Lock()
	defer processLock.Unlock()
	processes[process.Pid] = process
}

func removeProcess(process *os.Process) {
	processLock.Lock()
	defer processLock.Unlock()
	delete(processes, process.Pid)
}

func termPids() {
	running = false
	processLock.Lock()
	defer processLock.Unlock()

	for _, process := range processes {
		process.Signal(syscall.SIGTERM)
	}
}

func execute(line string, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	count := 0

	for {
		args := strings.Split(line, " ")

		cmd := exec.Command("setsid", args...)

		err := cmd.Start()
		if err != nil {
			log.Errorf("%s : %v", line, err)
		}

		if err == nil {
			addProcess(cmd.Process)
			err = cmd.Wait()
			removeProcess(cmd.Process)
		}

		if err != nil {
			log.Errorf("%s : %v", line, err)
		}

		if !running {
			log.Infof("%s : not restarting, exiting", line)
			break
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
}

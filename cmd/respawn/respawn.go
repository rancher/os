package respawn

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/log"

	"github.com/codegangsta/cli"
)

var (
	running     = true
	processes   = map[int]*os.Process{}
	processLock = sync.Mutex{}
)

func Main() {
	log.InitLogger()
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = fmt.Sprintf("%s RancherOS\nbuilt: %s", app.Name, config.BuildDate)
	app.Version = config.Version
	app.Author = "Rancher Labs, Inc."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Optional config file to load",
		},
	}
	app.Action = run

	log.Infof("%s, %s", app.Usage, app.Version)
	fmt.Printf("%s, %s", app.Usage, app.Version)

	app.Run(os.Args)
}

func setupSigterm() {
	sigtermChan := make(chan os.Signal)
	signal.Notify(sigtermChan, syscall.SIGTERM)
	go func() {
		for range sigtermChan {
			termPids()
		}
	}()
}

func run(c *cli.Context) error {
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

	lines := strings.Split(string(input), "\n")
	doneChannel := make(chan string, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.Index(strings.TrimSpace(line), "#") == 0 {
			continue
		}
		go execute(line, doneChannel)
	}

	for i := 0; i < len(lines); i++ {
		line := <-doneChannel
		log.Infof("FINISHED: %s", line)
		fmt.Printf("FINISHED: %s", line)
	}
	return nil
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
		log.Infof("sending SIGTERM to %d", process.Pid)
		process.Signal(syscall.SIGTERM)
	}
}

func execute(line string, doneChannel chan string) {
	defer func() { doneChannel <- line }()

	start := time.Now()
	count := 0

	args := strings.Split(line, " ")

	for {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		if err := cmd.Start(); err == nil {
			addProcess(cmd.Process)
			if err = cmd.Wait(); err != nil {
				log.Errorf("Wait cmd to exit: %s, err: %v", line, err)
			}
			removeProcess(cmd.Process)
		} else {
			log.Errorf("Start cmd: %s, err: %v", line, err)
		}

		if !running {
			log.Infof("%s : not restarting, exiting", line)
			break
		}

		count++

		if count > 10 {
			if time.Now().Sub(start) <= (1 * time.Second) {
				log.Errorf("%s : restarted too fast, not executing", line)
				break
			}

			count = 0
			start = time.Now()
		}
	}
}

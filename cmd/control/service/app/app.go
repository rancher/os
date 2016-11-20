package app

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
)

func ProjectPs(p project.APIProject, c *cli.Context) error {
	qFlag := c.Bool("q")
	allInfo, err := p.Ps(context.Background(), qFlag, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	os.Stdout.WriteString(allInfo.String(!qFlag))
	return nil
}

func ProjectStop(p project.APIProject, c *cli.Context) error {
	err := p.Stop(context.Background(), c.Int("timeout"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectDown(p project.APIProject, c *cli.Context) error {
	options := options.Down{
		RemoveVolume:  c.Bool("volumes"),
		RemoveImages:  options.ImageType(c.String("rmi")),
		RemoveOrphans: c.Bool("remove-orphans"),
	}
	err := p.Down(context.Background(), options, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectBuild(p project.APIProject, c *cli.Context) error {
	config := options.Build{
		NoCache:     c.Bool("no-cache"),
		ForceRemove: c.Bool("force-rm"),
		Pull:        c.Bool("pull"),
	}
	err := p.Build(context.Background(), config, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectCreate(p project.APIProject, c *cli.Context) error {
	options := options.Create{
		NoRecreate:    c.Bool("no-recreate"),
		ForceRecreate: c.Bool("force-recreate"),
		NoBuild:       c.Bool("no-build"),
	}
	err := p.Create(context.Background(), options, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectUp(p project.APIProject, c *cli.Context) error {
	options := options.Up{
		Create: options.Create{
			NoRecreate:    c.Bool("no-recreate"),
			ForceRecreate: c.Bool("force-recreate"),
			NoBuild:       c.Bool("no-build"),
		},
	}
	ctx, cancelFun := context.WithCancel(context.Background())
	err := p.Up(ctx, options, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if c.Bool("foreground") {
		signalChan := make(chan os.Signal, 1)
		cleanupDone := make(chan bool)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		errChan := make(chan error)
		go func() {
			errChan <- p.Log(ctx, true, c.Args()...)
		}()
		go func() {
			select {
			case <-signalChan:
				fmt.Printf("\nGracefully stopping...\n")
				cancelFun()
				ProjectStop(p, c)
				cleanupDone <- true
			case err := <-errChan:
				if err != nil {
					logrus.Fatal(err)
				}
				cleanupDone <- true
			}
		}()
		<-cleanupDone
		return nil
	}
	return nil
}

func ProjectStart(p project.APIProject, c *cli.Context) error {
	err := p.Start(context.Background(), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectRestart(p project.APIProject, c *cli.Context) error {
	err := p.Restart(context.Background(), c.Int("timeout"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectLog(p project.APIProject, c *cli.Context) error {
	err := p.Log(context.Background(), c.Bool("follow"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectPull(p project.APIProject, c *cli.Context) error {
	err := p.Pull(context.Background(), c.Args()...)
	if err != nil && !c.Bool("ignore-pull-failures") {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectDelete(p project.APIProject, c *cli.Context) error {
	options := options.Delete{
		RemoveVolume: c.Bool("v"),
	}
	if !c.Bool("force") {
		options.BeforeDeleteCallback = func(stoppedContainers []string) bool {
			fmt.Printf("Going to remove %v\nAre you sure? [yN]\n", strings.Join(stoppedContainers, ", "))
			var answer string
			_, err := fmt.Scanln(&answer)
			if err != nil {
				logrus.Error(err)
				return false
			}
			if answer != "y" && answer != "Y" {
				return false
			}
			return true
		}
	}
	err := p.Delete(context.Background(), options, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectKill(p project.APIProject, c *cli.Context) error {
	err := p.Kill(context.Background(), c.String("signal"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

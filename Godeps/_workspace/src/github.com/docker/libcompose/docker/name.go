package docker

import (
	"fmt"
	"io"
	"time"

	"github.com/samalba/dockerclient"
)

const format = "%s_%s_%d"

type Namer interface {
	io.Closer
	Next() string
}

type inOrderNamer struct {
	names chan string
	done  chan bool
}

func OneName(client dockerclient.Client, project, service string) (string, error) {
	namer := NewNamer(client, project, service)
	defer namer.Close()

	return namer.Next(), nil
}

func NewNamer(client dockerclient.Client, project, service string) Namer {
	namer := &inOrderNamer{
		names: make(chan string),
		done:  make(chan bool),
	}

	go func() {
		for i := 1; true; i++ {
			name := fmt.Sprintf(format, project, service, i)
			c, err := GetContainerByName(client, name)
			if err != nil {
				// Sleep here to avoid crazy tight loop when things go south
				time.Sleep(time.Second * 1)
				continue
			}
			if c != nil {
				continue
			}

			select {
			case namer.names <- name:
			case <-namer.done:
				close(namer.names)
				return
			}
		}
	}()

	return namer
}

func (i *inOrderNamer) Next() string {
	return <-i.names
}

func (i *inOrderNamer) Close() error {
	close(i.done)
	return nil
}

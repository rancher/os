package project

import (
	"bytes"

	"github.com/Sirupsen/logrus"
)

var (
	infoEvents = map[Event]bool{
		PROJECT_DELETE_DONE:   true,
		PROJECT_DELETE_START:  true,
		PROJECT_DOWN_DONE:     true,
		PROJECT_DOWN_START:    true,
		PROJECT_RESTART_DONE:  true,
		PROJECT_RESTART_START: true,
		PROJECT_UP_DONE:       true,
		PROJECT_UP_START:      true,
		SERVICE_DELETE_START:  true,
		SERVICE_DELETE:        true,
		SERVICE_DOWN_START:    true,
		SERVICE_DOWN:          true,
		SERVICE_RESTART_START: true,
		SERVICE_RESTART:       true,
		SERVICE_UP_START:      true,
		SERVICE_UP:            true,
	}
)

type defaultListener struct {
	project    *Project
	listenChan chan ProjectEvent
	upCount    int
}

func NewDefaultListener(p *Project) chan<- ProjectEvent {
	l := defaultListener{
		listenChan: make(chan ProjectEvent),
		project:    p,
	}
	go l.start()
	return l.listenChan
}

func (d *defaultListener) start() {
	for event := range d.listenChan {
		buffer := bytes.NewBuffer(nil)
		if event.Data != nil {
			for k, v := range event.Data {
				if buffer.Len() > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(k)
				buffer.WriteString("=")
				buffer.WriteString(v)
			}
		}

		if event.Event == SERVICE_UP {
			d.upCount++
		}

		logf := logrus.Debugf

		if infoEvents[event.Event] {
			logf = logrus.Infof
		}

		if event.ServiceName == "" {
			logf("Project [%s]: %s %s", d.project.Name, event.Event, buffer.Bytes())
		} else {
			logf("[%d/%d] [%s]: %s %s", d.upCount, len(d.project.Configs), event.ServiceName, event.Event, buffer.Bytes())
		}
	}
}

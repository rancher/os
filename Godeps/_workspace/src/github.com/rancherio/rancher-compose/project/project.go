package project

import (
	"bytes"
	"errors"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ServiceState string

var (
	EXECUTED   ServiceState = ServiceState("executed")
	UNKNOWN    ServiceState = ServiceState("unknown")
	ErrRestart error        = errors.New("Restart execution")
)

type ProjectEvent struct {
	Event   Event
	Service Service
	Data    map[string]string
}

func NewProject(name string, factory ServiceFactory) *Project {
	return &Project{
		Name:     name,
		configs:  make(map[string]*ServiceConfig),
		Services: make(map[string]Service),
		factory:  factory,
	}
}

func (p *Project) AddConfig(name string, config *ServiceConfig) error {
	service, err := p.factory.Create(p, name, config)
	if err != nil {
		log.Errorf("Failed to create service for %s : %v", name, err)
		return err
	}

	p.Notify(SERVICE_ADD, service, nil)

	p.configs[name] = config
	p.Services[name] = service

	return nil
}

func (p *Project) Load(bytes []byte) error {
	configs := make(map[string]*ServiceConfig)
	err := yaml.Unmarshal(bytes, configs)
	if err != nil {
		log.Fatalf("Could not parse config for project %s : %v", p.Name, err)
	}

	for name, config := range configs {
		err := p.AddConfig(name, config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) Up() error {
	wrappers := make(map[string]*ServiceWrapper)

	for name, _ := range p.Services {
		wrappers[name] = NewServiceWrapper(name, p)
	}

	p.Notify(PROJECT_UP_START, nil, nil)

	return p.startAll(wrappers)
}

func (p *Project) startAll(wrappers map[string]*ServiceWrapper) error {
	for name, _ := range p.Services {
		if _, ok := wrappers[name]; !ok {
			wrappers[name] = NewServiceWrapper(name, p)
		}
	}

	restart := false

	for _, wrapper := range wrappers {
		wrapper.Reset()
	}

	for _, wrapper := range wrappers {
		go wrapper.Start(wrappers)
	}

	var firstError error

	for _, wrapper := range wrappers {
		err := wrapper.Wait()
		if err == ErrRestart {
			restart = true
		} else if err != nil {
			log.Errorf("Failed to start: %s : %v", wrapper.name, err)
			if firstError == nil {
				firstError = err
			}
		}
	}

	if restart {
		if p.ReloadCallback != nil {
			if err := p.ReloadCallback(); err != nil {
				log.Errorf("Failed calling callback: %v", err)
			}
		}
		return p.startAll(wrappers)
	} else {
		return firstError
	}
}

type ServiceWrapper struct {
	name     string
	services map[string]Service
	service  Service
	done     sync.WaitGroup
	state    ServiceState
	err      error
	project  *Project
}

func NewServiceWrapper(name string, p *Project) *ServiceWrapper {
	wrapper := &ServiceWrapper{
		name:     name,
		services: make(map[string]Service),
		service:  p.Services[name],
		state:    UNKNOWN,
		project:  p,
	}
	return wrapper
}

func (s *ServiceWrapper) Reset() {
	if s.err == ErrRestart {
		s.err = nil
	}
	s.done.Add(1)
}

func (s *ServiceWrapper) Start(wrappers map[string]*ServiceWrapper) {
	defer s.done.Done()

	if s.state == EXECUTED {
		return
	}

	for _, link := range append(s.service.Config().Links, s.service.Config().VolumesFrom...) {
		name := strings.Split(link, ":")[0]
		if wrapper, ok := wrappers[name]; ok {
			if wrapper.Wait() == ErrRestart {
				s.project.Notify(PROJECT_RELOAD, wrapper.service, nil)
				s.err = ErrRestart
				return
			}
		} else {
			log.Errorf("Failed to find %s", name)
		}
	}

	s.state = EXECUTED

	s.project.Notify(SERVICE_UP_START, s.service, nil)

	s.err = s.service.Up()
	if s.err == ErrRestart {
		s.project.Notify(SERVICE_UP, s.service, nil)
		s.project.Notify(PROJECT_RELOAD_TRIGGER, s.service, nil)
	} else if s.err != nil {
		log.Errorf("Failed to start %s : %v", s.name, s.err)
	} else {
		s.project.Notify(SERVICE_UP, s.service, nil)
	}
}

func (s *ServiceWrapper) Wait() error {
	s.done.Wait()
	return s.err
}

func (p *Project) AddListener(c chan<- ProjectEvent) {
	p.listeners = append(p.listeners, c)
}

func (p *Project) Notify(event Event, service Service, data map[string]string) {
	buffer := bytes.NewBuffer(nil)
	if data != nil {
		for k, v := range data {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(k)
			buffer.WriteString("=")
			buffer.WriteString(v)
		}
	}

	if event == SERVICE_UP {
		p.upCount++
	}

	logf := log.Debugf

	if SERVICE_UP == event {
		logf = log.Infof
	}

	if service == nil {
		logf("Project [%s]: %s %s", p.Name, event, buffer.Bytes())
	} else {
		logf("[%d/%d] [%s]: %s %s", p.upCount, len(p.Services), service.Name(), event, buffer.Bytes())
	}

	for _, l := range p.listeners {
		projectEvent := ProjectEvent{
			Event:   event,
			Service: service,
			Data:    data,
		}
		// Don't ever block
		select {
		case l <- projectEvent:
		default:
		}
	}
}

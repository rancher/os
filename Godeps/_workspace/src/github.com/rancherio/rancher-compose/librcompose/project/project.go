package project

import (
	"bytes"
	"errors"
	"strings"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
)

type ServiceState string

var (
	EXECUTED   ServiceState = ServiceState("executed")
	UNKNOWN    ServiceState = ServiceState("unknown")
	ErrRestart error        = errors.New("Restart execution")
)

type ProjectEvent struct {
	Event       Event
	ServiceName string
	Data        map[string]string
}

func NewProject(name string, factory ServiceFactory) *Project {
	return &Project{
		Name:    name,
		configs: make(map[string]*ServiceConfig),
		factory: factory,
	}
}

func (p *Project) CreateService(name string, config ServiceConfig) (Service, error) {
	if p.EnvironmentLookup != nil {
		parsedEnv := make([]string, 0, len(config.Environment.Slice()))

		for _, env := range config.Environment.Slice() {
			if strings.IndexRune(env, '=') != -1 {
				parsedEnv = append(parsedEnv, env)
				continue
			}

			for _, value := range p.EnvironmentLookup.Lookup(env, name, &config) {
				parsedEnv = append(parsedEnv, value)
			}
		}

		config.Environment = NewMaporslice(parsedEnv)
	}

	return p.factory.Create(p, name, &config)
}

func (p *Project) AddConfig(name string, config *ServiceConfig) error {
	p.Notify(SERVICE_ADD, name, nil)

	p.configs[name] = config
	p.reload = append(p.reload, name)

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

func (p *Project) loadWrappers(wrappers map[string]*serviceWrapper) error {
	for _, name := range p.reload {
		wrapper, err := newServiceWrapper(name, p)
		if err != nil {
			return err
		}
		wrappers[name] = wrapper
	}

	p.reload = []string{}

	return nil
}

func (p *Project) Up() error {
	wrappers := make(map[string]*serviceWrapper)

	p.Notify(PROJECT_UP_START, "", nil)

	err := p.startAll(wrappers)

	if err == nil {
		p.Notify(PROJECT_UP_DONE, "", nil)
	}

	return err
}

func (p *Project) startAll(wrappers map[string]*serviceWrapper) error {
	restart := false

	for _, wrapper := range wrappers {
		if err := wrapper.Reset(); err != nil {
			return err
		}
	}

	p.loadWrappers(wrappers)

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

func (p *Project) AddListener(c chan<- ProjectEvent) {
	p.listeners = append(p.listeners, c)
}

func (p *Project) Notify(event Event, serviceName string, data map[string]string) {
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

	if serviceName == "" {
		logf("Project [%s]: %s %s", p.Name, event, buffer.Bytes())
	} else {
		logf("[%d/%d] [%s]: %s %s", p.upCount, len(p.configs), serviceName, event, buffer.Bytes())
	}

	for _, l := range p.listeners {
		projectEvent := ProjectEvent{
			Event:       event,
			ServiceName: serviceName,
			Data:        data,
		}
		// Don't ever block
		select {
		case l <- projectEvent:
		default:
		}
	}
}

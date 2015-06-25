package project

import (
	"errors"
	"fmt"
	"strings"

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

type wrapperAction func(*serviceWrapper, map[string]*serviceWrapper)

func NewProject(name string, factory ServiceFactory) *Project {
	p := &Project{
		Name:    name,
		Configs: make(map[string]*ServiceConfig),
		factory: factory,
	}

	listener := defaultListener{
		listenChan: make(chan ProjectEvent),
		project:    p,
	}

	p.listeners = []chan<- ProjectEvent{listener.listenChan}

	go listener.start()

	return p
}

func (p *Project) CreateService(name string) (Service, error) {
	existing, ok := p.Configs[name]
	if !ok {
		return nil, fmt.Errorf("Failed to find service: %s", name)
	}

	// Copy because we are about to modify the environment
	config := *existing

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

		config.Environment = NewMaporEqualSlice(parsedEnv)
	}

	return p.factory.Create(p, name, &config)
}

func (p *Project) AddConfig(name string, config *ServiceConfig) error {
	p.Notify(SERVICE_ADD, name, nil)

	p.Configs[name] = config
	p.reload = append(p.reload, name)

	return nil
}

func (p *Project) Load(bytes []byte) error {
	configs := make(map[string]*ServiceConfig)
	configs, err := Merge(p, bytes)
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

func (p *Project) Create(services ...string) error {
	p.Notify(PROJECT_CREATE_START, "", nil)

	err := p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Create(wrappers)
	}))

	if err == nil {
		p.Notify(PROJECT_CREATE_DONE, "", nil)
	}

	return err
}

func (p *Project) Down(services ...string) error {
	p.Notify(PROJECT_DOWN_START, "", nil)

	err := p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Stop(wrappers)
	}))

	if err == nil {
		p.Notify(PROJECT_DOWN_DONE, "", nil)
	}

	return err
}

func (p *Project) Restart(services ...string) error {
	p.Notify(PROJECT_RESTART_START, "", nil)

	err := p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Restart(wrappers)
	}))

	if err == nil {
		p.Notify(PROJECT_RESTART_DONE, "", nil)
	}

	return err
}

func (p *Project) Up(services ...string) error {
	p.Notify(PROJECT_UP_START, "", nil)

	err := p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Start(wrappers)
	}))

	if err == nil {
		p.Notify(PROJECT_UP_DONE, "", nil)
	}

	return err
}

func (p *Project) Log(services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Log(wrappers)
	}))
}

func (p *Project) Delete(services ...string) error {
	p.Notify(PROJECT_DELETE_START, "", nil)

	err := p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Delete(wrappers)
	}))

	if err == nil {
		p.Notify(PROJECT_DELETE_DONE, "", nil)
	}

	return err
}

func isSelected(wrapper *serviceWrapper, selected map[string]bool) bool {
	return len(selected) == 0 || selected[wrapper.name]
}

func (p *Project) forEach(services []string, action wrapperAction) error {
	selected := make(map[string]bool)
	wrappers := make(map[string]*serviceWrapper)

	for _, s := range services {
		selected[s] = true
	}

	return p.traverse(selected, wrappers, action)
}

func (p *Project) traverse(selected map[string]bool, wrappers map[string]*serviceWrapper, action wrapperAction) error {
	restart := false

	for _, wrapper := range wrappers {
		if err := wrapper.Reset(); err != nil {
			return err
		}
	}

	p.loadWrappers(wrappers)

	for _, wrapper := range wrappers {
		if isSelected(wrapper, selected) {
			go action(wrapper, wrappers)
		} else {
			wrapper.Ignore()
		}
	}

	var firstError error

	for _, wrapper := range wrappers {
		if !isSelected(wrapper, selected) {
			continue
		}
		if err := wrapper.Wait(); err == ErrRestart {
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
		return p.traverse(selected, wrappers, action)
	} else {
		return firstError
	}
}

func (p *Project) AddListener(c chan<- ProjectEvent) {
	if !p.hasListeners {
		for _, l := range p.listeners {
			close(l)
		}
		p.hasListeners = true
		p.listeners = []chan<- ProjectEvent{c}
	} else {
		p.listeners = append(p.listeners, c)
	}
}

func (p *Project) Notify(event Event, serviceName string, data map[string]string) {
	projectEvent := ProjectEvent{
		Event:       event,
		ServiceName: serviceName,
		Data:        data,
	}

	for _, l := range p.listeners {
		// Don't ever block
		select {
		case l <- projectEvent:
		default:
		}
	}
}

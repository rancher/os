package project

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/utils"
)

type ServiceState string

var (
	EXECUTED       ServiceState = ServiceState("executed")
	UNKNOWN        ServiceState = ServiceState("unknown")
	ErrRestart     error        = errors.New("Restart execution")
	ErrUnsupported error        = errors.New("UnsupportedOperation")
)

type ProjectEvent struct {
	Event       Event
	ServiceName string
	Data        map[string]string
}

type wrapperAction func(*serviceWrapper, map[string]*serviceWrapper)
type serviceAction func(service Service) error

func NewProject(context *Context) *Project {
	p := &Project{
		context: context,
		Configs: make(map[string]*ServiceConfig),
	}

	if context.LoggerFactory == nil {
		context.LoggerFactory = &logger.NullLogger{}
	}

	context.Project = p

	p.listeners = []chan<- ProjectEvent{NewDefaultListener(p)}

	return p
}

func (p *Project) Parse() error {
	err := p.context.open()
	if err != nil {
		return err
	}

	p.Name = p.context.ProjectName

	if p.context.ComposeFile == "-" {
		p.File = "."
	} else {
		p.File = p.context.ComposeFile
	}

	if p.context.ComposeBytes != nil {
		return p.Load(p.context.ComposeBytes)
	}

	return nil
}

func (p *Project) CreateService(name string) (Service, error) {
	existing, ok := p.Configs[name]
	if !ok {
		return nil, fmt.Errorf("Failed to find service: %s", name)
	}

	// Copy because we are about to modify the environment
	config := *existing

	if p.context.EnvironmentLookup != nil {
		parsedEnv := make([]string, 0, len(config.Environment.Slice()))

		for _, env := range config.Environment.Slice() {
			if strings.IndexRune(env, '=') != -1 {
				parsedEnv = append(parsedEnv, env)
				continue
			}

			for _, value := range p.context.EnvironmentLookup.Lookup(env, name, &config) {
				parsedEnv = append(parsedEnv, value)
			}
		}

		config.Environment = NewMaporEqualSlice(parsedEnv)
	}

	return p.context.ServiceFactory.Create(p, name, &config)
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

func (p *Project) Build(services ...string) error {
	return p.perform(PROJECT_BUILD_START, PROJECT_BUILD_DONE, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, SERVICE_BUILD_START, SERVICE_BUILD, func(service Service) error {
			return service.Build()
		})
	}), nil)
}

func (p *Project) Create(services ...string) error {
	return p.perform(PROJECT_CREATE_START, PROJECT_CREATE_DONE, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, SERVICE_CREATE_START, SERVICE_CREATE, func(service Service) error {
			return service.Create()
		})
	}), nil)
}

func (p *Project) Down(services ...string) error {
	return p.perform(PROJECT_DOWN_START, PROJECT_DOWN_DONE, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, SERVICE_DOWN_START, SERVICE_DOWN, func(service Service) error {
			return service.Down()
		})
	}), nil)
}

func (p *Project) Restart(services ...string) error {
	return p.perform(PROJECT_RESTART_START, PROJECT_RESTART_DONE, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, SERVICE_RESTART_START, SERVICE_RESTART, func(service Service) error {
			return service.Restart()
		})
	}), nil)
}

func (p *Project) Start(services ...string) error {
	return p.perform(PROJECT_START_START, PROJECT_START_DONE, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, SERVICE_START_START, SERVICE_START, func(service Service) error {
			return service.Start()
		})
	}), nil)
}

func (p *Project) Up(services ...string) error {
	return p.perform(PROJECT_UP_START, PROJECT_UP_DONE, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, SERVICE_UP_START, SERVICE_UP, func(service Service) error {
			return service.Up()
		})
	}), func(service Service) error {
		return service.Create()
	})
}

func (p *Project) Log(services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, NO_EVENT, NO_EVENT, func(service Service) error {
			return service.Log()
		})
	}), nil)
}

func (p *Project) Pull(services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, SERVICE_PULL_START, SERVICE_PULL, func(service Service) error {
			return service.Pull()
		})
	}), nil)
}

func (p *Project) Delete(services ...string) error {
	return p.perform(PROJECT_DELETE_START, PROJECT_DELETE_DONE, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, SERVICE_DELETE_START, SERVICE_DELETE, func(service Service) error {
			return service.Delete()
		})
	}), nil)
}

func (p *Project) Kill(services ...string) error {
	return p.perform(PROJECT_KILL_START, PROJECT_KILL_DONE, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, SERVICE_KILL_START, SERVICE_KILL, func(service Service) error {
			return service.Kill()
		})
	}), nil)
}

func (p *Project) perform(start, done Event, services []string, action wrapperAction, cycleAction serviceAction) error {
	p.Notify(start, "", nil)

	err := p.forEach(services, action, cycleAction)

	p.Notify(done, "", nil)
	return err
}

func isSelected(wrapper *serviceWrapper, selected map[string]bool) bool {
	return len(selected) == 0 || selected[wrapper.name]
}

func (p *Project) forEach(services []string, action wrapperAction, cycleAction serviceAction) error {
	selected := make(map[string]bool)
	wrappers := make(map[string]*serviceWrapper)

	for _, s := range services {
		selected[s] = true
	}

	return p.traverse(selected, wrappers, action, cycleAction)
}

func (p *Project) startService(wrappers map[string]*serviceWrapper, history []string, selected, launched map[string]bool, wrapper *serviceWrapper, action wrapperAction, cycleAction serviceAction) error {
	if launched[wrapper.name] {
		return nil
	}

	launched[wrapper.name] = true
	history = append(history, wrapper.name)

	for _, dep := range wrapper.service.DependentServices() {
		target := wrappers[dep.Target]
		if target == nil {
			log.Errorf("Failed to find %s", dep.Target)
			continue
		}

		if utils.Contains(history, dep.Target) {
			cycle := strings.Join(append(history, dep.Target), "->")
			if dep.Optional {
				log.Debugf("Ignoring cycle for %s", cycle)
				wrapper.IgnoreDep(dep.Target)
				if cycleAction != nil {
					var err error
					log.Debugf("Running cycle action for %s", cycle)
					err = cycleAction(target.service)
					if err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf("Cycle detected in path %s", cycle)
			}

			continue
		}

		err := p.startService(wrappers, history, selected, launched, target, action, cycleAction)
		if err != nil {
			return err
		}
	}

	if isSelected(wrapper, selected) {
		log.Debugf("Launching action for %s", wrapper.name)
		go action(wrapper, wrappers)
	} else {
		wrapper.Ignore()
	}

	return nil
}

func (p *Project) traverse(selected map[string]bool, wrappers map[string]*serviceWrapper, action wrapperAction, cycleAction serviceAction) error {
	restart := false

	for _, wrapper := range wrappers {
		if err := wrapper.Reset(); err != nil {
			return err
		}
	}

	p.loadWrappers(wrappers)

	launched := map[string]bool{}

	for _, wrapper := range wrappers {
		p.startService(wrappers, []string{}, selected, launched, wrapper, action, cycleAction)
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
		return p.traverse(selected, wrappers, action, cycleAction)
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
	if event == NO_EVENT {
		return
	}

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

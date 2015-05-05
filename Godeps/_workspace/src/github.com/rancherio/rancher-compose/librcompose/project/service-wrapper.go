package project

import (
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type serviceWrapper struct {
	name    string
	service Service
	done    sync.WaitGroup
	state   ServiceState
	err     error
	project *Project
}

func newServiceWrapper(name string, p *Project) (*serviceWrapper, error) {
	wrapper := &serviceWrapper{
		name:    name,
		state:   UNKNOWN,
		project: p,
	}

	return wrapper, wrapper.Reset()
}

func (s *serviceWrapper) Reset() error {
	if s.state != EXECUTED {
		service, err := s.project.CreateService(s.name, *s.project.configs[s.name])
		if err != nil {
			log.Errorf("Failed to create service for %s : %v", s.name, err)
			return err
		}

		s.service = service
	}

	if s.err == ErrRestart {
		s.err = nil
	}
	s.done.Add(1)

	return nil
}

func (s *serviceWrapper) Start(wrappers map[string]*serviceWrapper) {
	defer s.done.Done()

	if s.state == EXECUTED {
		return
	}

	for _, link := range append(s.service.Config().Links, s.service.Config().VolumesFrom...) {
		name := strings.Split(link, ":")[0]
		if wrapper, ok := wrappers[name]; ok {
			if wrapper.Wait() == ErrRestart {
				s.project.Notify(PROJECT_RELOAD, wrapper.service.Name(), nil)
				s.err = ErrRestart
				return
			}
		} else {
			log.Errorf("Failed to find %s", name)
		}
	}

	s.state = EXECUTED

	s.project.Notify(SERVICE_UP_START, s.service.Name(), nil)

	s.err = s.service.Up()
	if s.err == ErrRestart {
		s.project.Notify(SERVICE_UP, s.service.Name(), nil)
		s.project.Notify(PROJECT_RELOAD_TRIGGER, s.service.Name(), nil)
	} else if s.err != nil {
		log.Errorf("Failed to start %s : %v", s.name, s.err)
	} else {
		s.project.Notify(SERVICE_UP, s.service.Name(), nil)
	}
}

func (s *serviceWrapper) Wait() error {
	s.done.Wait()
	return s.err
}

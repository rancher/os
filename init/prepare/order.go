package prepare

import (
	composeConfig "github.com/docker/libcompose/config"
	"github.com/rancher/os/config"
)

type Tree struct {
	runafter  *Tree
	Name      string
	Reason    string
	runbefore *Tree
	parent    *Tree
}

type ServicesOrder struct {
	tree *Tree
	Map  map[string]*Tree
}

func GetServicesInOrder(serviceSet string) ServicesOrder {
	services := GetServiceSet(serviceSet)

	// Set the service name
	i := 1
	for name, _ := range services {
		//		log.Infof("naming (%d): %s", i, name)
		i = i + 1
		s := services[name]
		s.Name = name
	}
	//order these based on scope labels
	// unfortunantely, its a go hash, so the order (especially of the unconstrained services) will be random
	order := ServicesOrder{Map: make(map[string]*Tree)}
	for name, _ := range services {
		s := services[name]
		//		log.Infof("inserting: %s", s.Name)
		order = order.insert(services, s)
	}

	return order
}

func GetService(serviceSet, name string) *composeConfig.ServiceConfigV1 {
	cfg := config.LoadConfig()

	if serviceSet != "" {
		set := GetServiceSet(serviceSet)
		if set == nil {
			return nil
		}
		return set[name]
	}

	switch {
	case cfg.Rancher.Services[name] != nil:
		return cfg.Rancher.Services[name]
	case cfg.Rancher.BootstrapContainers[name] != nil:
		return cfg.Rancher.BootstrapContainers[name]
	case cfg.Rancher.CloudInitServices[name] != nil:
		return cfg.Rancher.CloudInitServices[name]
	case cfg.Rancher.RecoveryServices[name] != nil:
		return cfg.Rancher.RecoveryServices[name]
	}

	return nil
}

func GetServiceSet(name string) map[string]*composeConfig.ServiceConfigV1 {
	cfg := config.LoadConfig()
	switch name {
	case "services":
		return cfg.Rancher.Services
	case "bootstrap":
		return cfg.Rancher.BootstrapContainers
	case "cloud_init_services":
		return cfg.Rancher.CloudInitServices
	case "recovery_services":
		return cfg.Rancher.RecoveryServices
	}
	return nil
}

func (o ServicesOrder) insert(services map[string]*composeConfig.ServiceConfigV1, s *composeConfig.ServiceConfigV1) ServicesOrder {

	if _, ok := o.Map[s.Name]; ok {
		// lets not add the same service twice
		return o
	}
	e := &Tree{Name: s.Name}
	o.Map[s.Name] = e

	after, _ := s.Labels["io.rancher.os.after"]
	before, _ := s.Labels["io.rancher.os.before"]

	if after != "" {
		e.Reason = e.Reason + "(after)" + after
		a, ok := o.Map[after]
		if !ok {
			s := services[after]
			o = o.insert(services, s)
			a, ok = o.Map[after]
		}
		if a.parent != nil {
			// place the new element where the old one was.
			e.parent = a.parent
			if a.parent.runbefore == a {
				a.parent.runbefore = e
			}
			if a.parent.runafter == a {
				a.parent.runafter = e
			}
		}
		a.parent = e
		e.runafter = a
		if a == o.tree {
			o.tree = e
		}
	}
	if before != "" {
		e.Reason = e.Reason + "(before)" + before
		b, ok := o.Map[before]
		if !ok {
			s := services[before]
			o = o.insert(services, s)
			b, ok = o.Map[before]
		}
		if b.parent != nil {
			// place the new element where the old one was.
			e.parent = b.parent
			if b.parent.runbefore == b {
				b.parent.runbefore = e
			}
			if b.parent.runafter == b {
				b.parent.runafter = e
			}
		}
		b.parent = e
		e.runbefore = b
		if b == o.tree {
			o.tree = e
		}
	}

	if o.tree == nil && e.parent == nil {
		o.tree = e
		return o
	}

	if e.parent == nil && o.tree != e {
		// run last?
		var newParent = o.tree
		for newParent.runafter != nil {
			if newParent.runafter == o.tree {
				newParent.runafter = nil
			}
			newParent = newParent.runafter
		}
		newParent.runafter = e
		e.parent = newParent
	}

	return o
}

func Walk(t *Tree, ch chan *Tree) {
	if t == nil {
		return
	}
	Walk(t.runafter, ch)
	ch <- t
	Walk(t.runbefore, ch)
}

func (o ServicesOrder) Walker() <-chan *Tree {
	ch := make(chan *Tree)
	go func() {
		Walk(o.tree, ch)
		close(ch)
	}()
	return ch
}

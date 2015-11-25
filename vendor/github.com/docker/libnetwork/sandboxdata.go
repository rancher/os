package libnetwork

import (
	"container/heap"
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libnetwork/sandbox"
)

type epHeap []*endpoint

type sandboxData struct {
	sbox      sandbox.Sandbox
	refCnt    int
	endpoints epHeap
	sync.Mutex
}

func (eh epHeap) Len() int { return len(eh) }

func (eh epHeap) Less(i, j int) bool {
	eh[i].Lock()
	eh[j].Lock()
	defer eh[j].Unlock()
	defer eh[i].Unlock()

	if eh[i].container.config.prio == eh[j].container.config.prio {
		return eh[i].network.Name() < eh[j].network.Name()
	}

	return eh[i].container.config.prio > eh[j].container.config.prio
}

func (eh epHeap) Swap(i, j int) { eh[i], eh[j] = eh[j], eh[i] }

func (eh *epHeap) Push(x interface{}) {
	*eh = append(*eh, x.(*endpoint))
}

func (eh *epHeap) Pop() interface{} {
	old := *eh
	n := len(old)
	x := old[n-1]
	*eh = old[0 : n-1]
	return x
}

func (s *sandboxData) updateGateway(ep *endpoint) error {
	sb := s.sandbox()

	sb.UnsetGateway()
	sb.UnsetGatewayIPv6()

	if ep == nil {
		return nil
	}

	ep.Lock()
	joinInfo := ep.joinInfo
	ep.Unlock()

	if err := sb.SetGateway(joinInfo.gw); err != nil {
		return fmt.Errorf("failed to set gateway while updating gateway: %v", err)
	}

	if err := sb.SetGatewayIPv6(joinInfo.gw6); err != nil {
		return fmt.Errorf("failed to set IPv6 gateway while updating gateway: %v", err)
	}

	return nil
}

func (s *sandboxData) addEndpoint(ep *endpoint) error {
	ep.Lock()
	joinInfo := ep.joinInfo
	ifaces := ep.iFaces
	ep.Unlock()

	sb := s.sandbox()
	for _, i := range ifaces {
		var ifaceOptions []sandbox.IfaceOption

		ifaceOptions = append(ifaceOptions, sb.InterfaceOptions().Address(&i.addr),
			sb.InterfaceOptions().Routes(i.routes))
		if i.addrv6.IP.To16() != nil {
			ifaceOptions = append(ifaceOptions,
				sb.InterfaceOptions().AddressIPv6(&i.addrv6))
		}

		if err := sb.AddInterface(i.srcName, i.dstPrefix, ifaceOptions...); err != nil {
			return fmt.Errorf("failed to add interface %s to sandbox: %v", i.srcName, err)
		}
	}

	if joinInfo != nil {
		// Set up non-interface routes.
		for _, r := range ep.joinInfo.StaticRoutes {
			if err := sb.AddStaticRoute(r); err != nil {
				return fmt.Errorf("failed to add static route %s: %v", r.Destination.String(), err)
			}
		}
	}

	s.Lock()
	heap.Push(&s.endpoints, ep)
	highEp := s.endpoints[0]
	s.Unlock()

	if ep == highEp {
		if err := s.updateGateway(ep); err != nil {
			return err
		}
	}

	return nil
}

func (s *sandboxData) rmEndpoint(ep *endpoint) {
	ep.Lock()
	joinInfo := ep.joinInfo
	ep.Unlock()

	sb := s.sandbox()
	for _, i := range sb.Info().Interfaces() {
		// Only remove the interfaces owned by this endpoint from the sandbox.
		if ep.hasInterface(i.SrcName()) {
			if err := i.Remove(); err != nil {
				logrus.Debugf("Remove interface failed: %v", err)
			}
		}
	}

	// Remove non-interface routes.
	for _, r := range joinInfo.StaticRoutes {
		if err := sb.RemoveStaticRoute(r); err != nil {
			logrus.Debugf("Remove route failed: %v", err)
		}
	}

	// We don't check if s.endpoints is empty here because
	// it should never be empty during a rmEndpoint call and
	// if it is we will rightfully panic here
	s.Lock()
	highEpBefore := s.endpoints[0]
	var (
		i int
		e *endpoint
	)
	for i, e = range s.endpoints {
		if e == ep {
			break
		}
	}
	heap.Remove(&s.endpoints, i)
	var highEpAfter *endpoint
	if len(s.endpoints) > 0 {
		highEpAfter = s.endpoints[0]
	}

	s.Unlock()

	if highEpBefore != highEpAfter {
		s.updateGateway(highEpAfter)
	}
}

func (s *sandboxData) sandbox() sandbox.Sandbox {
	s.Lock()
	defer s.Unlock()

	return s.sbox
}

func (c *controller) sandboxAdd(key string, create bool, ep *endpoint) (sandbox.Sandbox, error) {
	c.Lock()
	sData, ok := c.sandboxes[key]
	c.Unlock()

	if !ok {
		sb, err := sandbox.NewSandbox(key, create)
		if err != nil {
			return nil, fmt.Errorf("failed to create new sandbox: %v", err)
		}

		sData = &sandboxData{
			sbox:      sb,
			endpoints: epHeap{},
		}

		heap.Init(&sData.endpoints)
		c.Lock()
		c.sandboxes[key] = sData
		c.Unlock()
	}

	if err := sData.addEndpoint(ep); err != nil {
		return nil, err
	}

	return sData.sandbox(), nil
}

func (c *controller) sandboxRm(key string, ep *endpoint) {
	c.Lock()
	sData := c.sandboxes[key]
	c.Unlock()

	sData.rmEndpoint(ep)
}

func (c *controller) sandboxGet(key string) sandbox.Sandbox {
	c.Lock()
	sData, ok := c.sandboxes[key]
	c.Unlock()

	if !ok {
		return nil
	}

	return sData.sandbox()
}

func (c *controller) LeaveAll(id string) error {
	c.Lock()
	sData, ok := c.sandboxes[sandbox.GenerateKey(id)]
	c.Unlock()

	if !ok {
		return fmt.Errorf("could not find sandbox for container id %s", id)
	}

	sData.Lock()
	eps := make([]*endpoint, len(sData.endpoints))
	for i, ep := range sData.endpoints {
		eps[i] = ep
	}
	sData.Unlock()

	for _, ep := range eps {
		if err := ep.Leave(id); err != nil {
			logrus.Warnf("Failed leaving endpoint id %s: %v\n", ep.ID(), err)
		}
	}

	sData.sandbox().Destroy()
	delete(c.sandboxes, sandbox.GenerateKey(id))

	return nil
}

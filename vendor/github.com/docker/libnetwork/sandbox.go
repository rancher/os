package libnetwork

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libnetwork/etchosts"
	"github.com/docker/libnetwork/osl"
	"github.com/docker/libnetwork/resolvconf"
	"github.com/docker/libnetwork/types"
)

// Sandbox provides the control over the network container entity. It is a one to one mapping with the container.
type Sandbox interface {
	// ID returns the ID of the sandbox
	ID() string
	// Key returns the sandbox's key
	Key() string
	// ContainerID returns the container id associated to this sandbox
	ContainerID() string
	// Labels returns the sandbox's labels
	Labels() map[string]interface{}
	// Statistics retrieves the interfaces' statistics for the sandbox
	Statistics() (map[string]*types.InterfaceStatistics, error)
	// Refresh leaves all the endpoints, resets and re-apply the options,
	// re-joins all the endpoints without destroying the osl sandbox
	Refresh(options ...SandboxOption) error
	// SetKey updates the Sandbox Key
	SetKey(key string) error
	// Rename changes the name of all attached Endpoints
	Rename(name string) error
	// Delete destroys this container after detaching it from all connected endpoints.
	Delete() error
	// ResolveName searches for the service name in the networks to which the sandbox
	// is connected to.
	ResolveName(name string) net.IP
	// ResolveIP returns the service name for the passed in IP. IP is in reverse dotted
	// notation; the format used for DNS PTR records
	ResolveIP(name string) string
}

// SandboxOption is a option setter function type used to pass varios options to
// NewNetContainer method. The various setter functions of type SandboxOption are
// provided by libnetwork, they look like ContainerOptionXXXX(...)
type SandboxOption func(sb *sandbox)

func (sb *sandbox) processOptions(options ...SandboxOption) {
	for _, opt := range options {
		if opt != nil {
			opt(sb)
		}
	}
}

type epHeap []*endpoint

type sandbox struct {
	id            string
	containerID   string
	config        containerConfig
	extDNS        []string
	osSbox        osl.Sandbox
	controller    *controller
	resolver      Resolver
	resolverOnce  sync.Once
	refCnt        int
	endpoints     epHeap
	epPriority    map[string]int
	joinLeaveDone chan struct{}
	dbIndex       uint64
	dbExists      bool
	isStub        bool
	inDelete      bool
	sync.Mutex
}

// These are the container configs used to customize container /etc/hosts file.
type hostsPathConfig struct {
	hostName        string
	domainName      string
	hostsPath       string
	originHostsPath string
	extraHosts      []extraHost
	parentUpdates   []parentUpdate
}

type parentUpdate struct {
	cid  string
	name string
	ip   string
}

type extraHost struct {
	name string
	IP   string
}

// These are the container configs used to customize container /etc/resolv.conf file.
type resolvConfPathConfig struct {
	resolvConfPath       string
	originResolvConfPath string
	resolvConfHashFile   string
	dnsList              []string
	dnsSearchList        []string
	dnsOptionsList       []string
}

type containerConfig struct {
	hostsPathConfig
	resolvConfPathConfig
	generic           map[string]interface{}
	useDefaultSandBox bool
	useExternalKey    bool
	prio              int // higher the value, more the priority
}

func (sb *sandbox) ID() string {
	return sb.id
}

func (sb *sandbox) ContainerID() string {
	return sb.containerID
}

func (sb *sandbox) Key() string {
	if sb.config.useDefaultSandBox {
		return osl.GenerateKey("default")
	}
	return osl.GenerateKey(sb.id)
}

func (sb *sandbox) Labels() map[string]interface{} {
	return sb.config.generic
}

func (sb *sandbox) Statistics() (map[string]*types.InterfaceStatistics, error) {
	m := make(map[string]*types.InterfaceStatistics)

	if sb.osSbox == nil {
		return m, nil
	}

	var err error
	for _, i := range sb.osSbox.Info().Interfaces() {
		if m[i.DstName()], err = i.Statistics(); err != nil {
			return m, err
		}
	}

	return m, nil
}

func (sb *sandbox) Delete() error {
	sb.Lock()
	if sb.inDelete {
		sb.Unlock()
		return types.ForbiddenErrorf("another sandbox delete in progress")
	}
	// Set the inDelete flag. This will ensure that we don't
	// update the store until we have completed all the endpoint
	// leaves and deletes. And when endpoint leaves and deletes
	// are completed then we can finally delete the sandbox object
	// altogether from the data store. If the daemon exits
	// ungracefully in the middle of a sandbox delete this way we
	// will have all the references to the endpoints in the
	// sandbox so that we can clean them up when we restart
	sb.inDelete = true
	sb.Unlock()

	c := sb.controller

	// Detach from all endpoints
	retain := false
	for _, ep := range sb.getConnectedEndpoints() {
		// endpoint in the Gateway network will be cleaned up
		// when when sandbox no longer needs external connectivity
		if ep.endpointInGWNetwork() {
			continue
		}

		// Retain the sanbdox if we can't obtain the network from store.
		if _, err := c.getNetworkFromStore(ep.getNetwork().ID()); err != nil {
			retain = true
			log.Warnf("Failed getting network for ep %s during sandbox %s delete: %v", ep.ID(), sb.ID(), err)
			continue
		}

		if err := ep.Leave(sb); err != nil {
			log.Warnf("Failed detaching sandbox %s from endpoint %s: %v\n", sb.ID(), ep.ID(), err)
		}

		if err := ep.Delete(false); err != nil {
			log.Warnf("Failed deleting endpoint %s: %v\n", ep.ID(), err)
		}
	}

	if retain {
		sb.Lock()
		sb.inDelete = false
		sb.Unlock()
		return fmt.Errorf("could not cleanup all the endpoints in container %s / sandbox %s", sb.containerID, sb.id)
	}
	// Container is going away. Path cache in etchosts is most
	// likely not required any more. Drop it.
	etchosts.Drop(sb.config.hostsPath)

	if sb.resolver != nil {
		sb.resolver.Stop()
	}

	if sb.osSbox != nil && !sb.config.useDefaultSandBox {
		sb.osSbox.Destroy()
	}

	if err := sb.storeDelete(); err != nil {
		log.Warnf("Failed to delete sandbox %s from store: %v", sb.ID(), err)
	}

	c.Lock()
	delete(c.sandboxes, sb.ID())
	c.Unlock()

	return nil
}

func (sb *sandbox) Rename(name string) error {
	var err error

	for _, ep := range sb.getConnectedEndpoints() {
		if ep.endpointInGWNetwork() {
			continue
		}

		oldName := ep.Name()
		lEp := ep
		if err = ep.rename(name); err != nil {
			break
		}

		defer func() {
			if err != nil {
				lEp.rename(oldName)
			}
		}()
	}

	return err
}

func (sb *sandbox) Refresh(options ...SandboxOption) error {
	// Store connected endpoints
	epList := sb.getConnectedEndpoints()

	// Detach from all endpoints
	for _, ep := range epList {
		if err := ep.Leave(sb); err != nil {
			log.Warnf("Failed detaching sandbox %s from endpoint %s: %v\n", sb.ID(), ep.ID(), err)
		}
	}

	// Re-apply options
	sb.config = containerConfig{}
	sb.processOptions(options...)

	// Setup discovery files
	if err := sb.setupResolutionFiles(); err != nil {
		return err
	}

	// Re -connect to all endpoints
	for _, ep := range epList {
		if err := ep.Join(sb); err != nil {
			log.Warnf("Failed attach sandbox %s to endpoint %s: %v\n", sb.ID(), ep.ID(), err)
		}
	}

	return nil
}

func (sb *sandbox) MarshalJSON() ([]byte, error) {
	sb.Lock()
	defer sb.Unlock()

	// We are just interested in the container ID. This can be expanded to include all of containerInfo if there is a need
	return json.Marshal(sb.id)
}

func (sb *sandbox) UnmarshalJSON(b []byte) (err error) {
	sb.Lock()
	defer sb.Unlock()

	var id string
	if err := json.Unmarshal(b, &id); err != nil {
		return err
	}
	sb.id = id
	return nil
}

func (sb *sandbox) startResolver() {
	sb.resolverOnce.Do(func() {
		var err error
		sb.resolver = NewResolver(sb)
		defer func() {
			if err != nil {
				sb.resolver = nil
			}
		}()

		sb.rebuildDNS()
		sb.resolver.SetExtServers(sb.extDNS)

		sb.osSbox.InvokeFunc(sb.resolver.SetupFunc())
		if err := sb.resolver.Start(); err != nil {
			log.Errorf("Resolver Setup/Start failed for container %s, %q", sb.ContainerID(), err)
		}
	})
}

func (sb *sandbox) setupResolutionFiles() error {
	if err := sb.buildHostsFile(); err != nil {
		return err
	}

	if err := sb.updateParentHosts(); err != nil {
		return err
	}

	if err := sb.setupDNS(); err != nil {
		return err
	}

	return nil
}

func (sb *sandbox) getConnectedEndpoints() []*endpoint {
	sb.Lock()
	defer sb.Unlock()

	eps := make([]*endpoint, len(sb.endpoints))
	for i, ep := range sb.endpoints {
		eps[i] = ep
	}

	return eps
}

func (sb *sandbox) getEndpoint(id string) *endpoint {
	sb.Lock()
	defer sb.Unlock()

	for _, ep := range sb.endpoints {
		if ep.id == id {
			return ep
		}
	}

	return nil
}

func (sb *sandbox) updateGateway(ep *endpoint) error {
	sb.Lock()
	osSbox := sb.osSbox
	sb.Unlock()
	if osSbox == nil {
		return nil
	}
	osSbox.UnsetGateway()
	osSbox.UnsetGatewayIPv6()

	if ep == nil {
		return nil
	}

	ep.Lock()
	joinInfo := ep.joinInfo
	ep.Unlock()

	if err := osSbox.SetGateway(joinInfo.gw); err != nil {
		return fmt.Errorf("failed to set gateway while updating gateway: %v", err)
	}

	if err := osSbox.SetGatewayIPv6(joinInfo.gw6); err != nil {
		return fmt.Errorf("failed to set IPv6 gateway while updating gateway: %v", err)
	}

	return nil
}

func (sb *sandbox) ResolveIP(ip string) string {
	var svc string
	log.Debugf("IP To resolve %v", ip)

	for _, ep := range sb.getConnectedEndpoints() {
		n := ep.getNetwork()

		sr, ok := n.getController().svcDb[n.ID()]
		if !ok {
			continue
		}

		nwName := n.Name()
		n.Lock()
		svc, ok = sr.ipMap[ip]
		n.Unlock()
		if ok {
			return svc + "." + nwName
		}
	}
	return svc
}

func (sb *sandbox) ResolveName(name string) net.IP {
	var ip net.IP
	parts := strings.Split(name, ".")
	log.Debugf("To resolve %v", parts)

	reqName := parts[0]
	networkName := ""
	if len(parts) > 1 {
		networkName = parts[1]
	}
	epList := sb.getConnectedEndpoints()
	// First check for local container alias
	ip = sb.resolveName(reqName, networkName, epList, true)
	if ip != nil {
		return ip
	}

	// Resolve the actual container name
	return sb.resolveName(reqName, networkName, epList, false)
}

func (sb *sandbox) resolveName(req string, networkName string, epList []*endpoint, alias bool) net.IP {
	for _, ep := range epList {
		name := req
		n := ep.getNetwork()

		if networkName != "" && networkName != n.Name() {
			continue
		}

		if alias {
			if ep.aliases == nil {
				continue
			}

			var ok bool
			ep.Lock()
			name, ok = ep.aliases[req]
			ep.Unlock()
			if !ok {
				continue
			}
		} else {
			// If it is a regular lookup and if the requested name is an alias
			// dont perform a svc lookup for this endpoint.
			ep.Lock()
			if _, ok := ep.aliases[req]; ok {
				ep.Unlock()
				continue
			}
			ep.Unlock()
		}

		sr, ok := n.getController().svcDb[n.ID()]
		if !ok {
			continue
		}

		n.Lock()
		ip, ok := sr.svcMap[name]
		n.Unlock()
		if ok {
			return ip[0]
		}
	}
	return nil
}

func (sb *sandbox) SetKey(basePath string) error {
	if basePath == "" {
		return types.BadRequestErrorf("invalid sandbox key")
	}

	sb.Lock()
	oldosSbox := sb.osSbox
	sb.Unlock()

	if oldosSbox != nil {
		// If we already have an OS sandbox, release the network resources from that
		// and destroy the OS snab. We are moving into a new home further down. Note that none
		// of the network resources gets destroyed during the move.
		sb.releaseOSSbox()
	}

	osSbox, err := osl.GetSandboxForExternalKey(basePath, sb.Key())
	if err != nil {
		return err
	}

	sb.Lock()
	sb.osSbox = osSbox
	sb.Unlock()
	defer func() {
		if err != nil {
			sb.Lock()
			sb.osSbox = nil
			sb.Unlock()
		}
	}()

	// If the resolver was setup before stop it and set it up in the
	// new osl sandbox.
	if oldosSbox != nil && sb.resolver != nil {
		sb.resolver.Stop()

		sb.osSbox.InvokeFunc(sb.resolver.SetupFunc())
		if err := sb.resolver.Start(); err != nil {
			log.Errorf("Resolver Setup/Start failed for container %s, %q", sb.ContainerID(), err)
		}
	}

	for _, ep := range sb.getConnectedEndpoints() {
		if err = sb.populateNetworkResources(ep); err != nil {
			return err
		}
	}
	return nil
}

func releaseOSSboxResources(osSbox osl.Sandbox, ep *endpoint) {
	for _, i := range osSbox.Info().Interfaces() {
		// Only remove the interfaces owned by this endpoint from the sandbox.
		if ep.hasInterface(i.SrcName()) {
			if err := i.Remove(); err != nil {
				log.Debugf("Remove interface failed: %v", err)
			}
		}
	}

	ep.Lock()
	joinInfo := ep.joinInfo
	ep.Unlock()

	if joinInfo == nil {
		return
	}

	// Remove non-interface routes.
	for _, r := range joinInfo.StaticRoutes {
		if err := osSbox.RemoveStaticRoute(r); err != nil {
			log.Debugf("Remove route failed: %v", err)
		}
	}
}

func (sb *sandbox) releaseOSSbox() {
	sb.Lock()
	osSbox := sb.osSbox
	sb.osSbox = nil
	sb.Unlock()

	if osSbox == nil {
		return
	}

	for _, ep := range sb.getConnectedEndpoints() {
		releaseOSSboxResources(osSbox, ep)
	}

	osSbox.Destroy()
}

func (sb *sandbox) populateNetworkResources(ep *endpoint) error {
	sb.Lock()
	if sb.osSbox == nil {
		sb.Unlock()
		return nil
	}
	inDelete := sb.inDelete
	sb.Unlock()

	ep.Lock()
	joinInfo := ep.joinInfo
	i := ep.iface
	ep.Unlock()

	if ep.needResolver() {
		sb.startResolver()
	}

	if i != nil && i.srcName != "" {
		var ifaceOptions []osl.IfaceOption

		ifaceOptions = append(ifaceOptions, sb.osSbox.InterfaceOptions().Address(i.addr), sb.osSbox.InterfaceOptions().Routes(i.routes))
		if i.addrv6 != nil && i.addrv6.IP.To16() != nil {
			ifaceOptions = append(ifaceOptions, sb.osSbox.InterfaceOptions().AddressIPv6(i.addrv6))
		}

		if err := sb.osSbox.AddInterface(i.srcName, i.dstPrefix, ifaceOptions...); err != nil {
			return fmt.Errorf("failed to add interface %s to sandbox: %v", i.srcName, err)
		}
	}

	if joinInfo != nil {
		// Set up non-interface routes.
		for _, r := range joinInfo.StaticRoutes {
			if err := sb.osSbox.AddStaticRoute(r); err != nil {
				return fmt.Errorf("failed to add static route %s: %v", r.Destination.String(), err)
			}
		}
	}

	for _, gwep := range sb.getConnectedEndpoints() {
		if len(gwep.Gateway()) > 0 {
			if gwep != ep {
				break
			}
			if err := sb.updateGateway(gwep); err != nil {
				return err
			}
		}
	}

	// Only update the store if we did not come here as part of
	// sandbox delete. If we came here as part of delete then do
	// not bother updating the store. The sandbox object will be
	// deleted anyway
	if !inDelete {
		return sb.storeUpdate()
	}

	return nil
}

func (sb *sandbox) clearNetworkResources(origEp *endpoint) error {
	ep := sb.getEndpoint(origEp.id)
	if ep == nil {
		return fmt.Errorf("could not find the sandbox endpoint data for endpoint %s",
			ep.name)
	}

	sb.Lock()
	osSbox := sb.osSbox
	inDelete := sb.inDelete
	sb.Unlock()
	if osSbox != nil {
		releaseOSSboxResources(osSbox, ep)
	}

	sb.Lock()
	if len(sb.endpoints) == 0 {
		// sb.endpoints should never be empty and this is unexpected error condition
		// We log an error message to note this down for debugging purposes.
		log.Errorf("No endpoints in sandbox while trying to remove endpoint %s", ep.Name())
		sb.Unlock()
		return nil
	}

	var (
		gwepBefore, gwepAfter *endpoint
		index                 = -1
	)
	for i, e := range sb.endpoints {
		if e == ep {
			index = i
		}
		if len(e.Gateway()) > 0 && gwepBefore == nil {
			gwepBefore = e
		}
		if index != -1 && gwepBefore != nil {
			break
		}
	}
	heap.Remove(&sb.endpoints, index)
	for _, e := range sb.endpoints {
		if len(e.Gateway()) > 0 {
			gwepAfter = e
			break
		}
	}
	delete(sb.epPriority, ep.ID())
	sb.Unlock()

	if gwepAfter != nil && gwepBefore != gwepAfter {
		sb.updateGateway(gwepAfter)
	}

	// Only update the store if we did not come here as part of
	// sandbox delete. If we came here as part of delete then do
	// not bother updating the store. The sandbox object will be
	// deleted anyway
	if !inDelete {
		return sb.storeUpdate()
	}

	return nil
}

const (
	defaultPrefix = "/var/lib/docker/network/files"
	dirPerm       = 0755
	filePerm      = 0644
)

func (sb *sandbox) buildHostsFile() error {
	if sb.config.hostsPath == "" {
		sb.config.hostsPath = defaultPrefix + "/" + sb.id + "/hosts"
	}

	dir, _ := filepath.Split(sb.config.hostsPath)
	if err := createBasePath(dir); err != nil {
		return err
	}

	// This is for the host mode networking
	if sb.config.originHostsPath != "" {
		if err := copyFile(sb.config.originHostsPath, sb.config.hostsPath); err != nil && !os.IsNotExist(err) {
			return types.InternalErrorf("could not copy source hosts file %s to %s: %v", sb.config.originHostsPath, sb.config.hostsPath, err)
		}
		return nil
	}

	extraContent := make([]etchosts.Record, 0, len(sb.config.extraHosts))
	for _, extraHost := range sb.config.extraHosts {
		extraContent = append(extraContent, etchosts.Record{Hosts: extraHost.name, IP: extraHost.IP})
	}

	return etchosts.Build(sb.config.hostsPath, "", sb.config.hostName, sb.config.domainName, extraContent)
}

func (sb *sandbox) updateHostsFile(ifaceIP string) error {
	var mhost string

	if sb.config.originHostsPath != "" {
		return nil
	}

	if sb.config.domainName != "" {
		mhost = fmt.Sprintf("%s.%s %s", sb.config.hostName, sb.config.domainName,
			sb.config.hostName)
	} else {
		mhost = sb.config.hostName
	}

	extraContent := []etchosts.Record{{Hosts: mhost, IP: ifaceIP}}

	sb.addHostsEntries(extraContent)
	return nil
}

func (sb *sandbox) addHostsEntries(recs []etchosts.Record) {
	if err := etchosts.Add(sb.config.hostsPath, recs); err != nil {
		log.Warnf("Failed adding service host entries to the running container: %v", err)
	}
}

func (sb *sandbox) deleteHostsEntries(recs []etchosts.Record) {
	if err := etchosts.Delete(sb.config.hostsPath, recs); err != nil {
		log.Warnf("Failed deleting service host entries to the running container: %v", err)
	}
}

func (sb *sandbox) updateParentHosts() error {
	var pSb Sandbox

	for _, update := range sb.config.parentUpdates {
		sb.controller.WalkSandboxes(SandboxContainerWalker(&pSb, update.cid))
		if pSb == nil {
			continue
		}
		if err := etchosts.Update(pSb.(*sandbox).config.hostsPath, update.ip, update.name); err != nil {
			return err
		}
	}

	return nil
}

func (sb *sandbox) setupDNS() error {
	var newRC *resolvconf.File

	if sb.config.resolvConfPath == "" {
		sb.config.resolvConfPath = defaultPrefix + "/" + sb.id + "/resolv.conf"
	}

	sb.config.resolvConfHashFile = sb.config.resolvConfPath + ".hash"

	dir, _ := filepath.Split(sb.config.resolvConfPath)
	if err := createBasePath(dir); err != nil {
		return err
	}

	// This is for the host mode networking
	if sb.config.originResolvConfPath != "" {
		if err := copyFile(sb.config.originResolvConfPath, sb.config.resolvConfPath); err != nil {
			return fmt.Errorf("could not copy source resolv.conf file %s to %s: %v", sb.config.originResolvConfPath, sb.config.resolvConfPath, err)
		}
		return nil
	}

	currRC, err := resolvconf.Get()
	if err != nil {
		return err
	}

	if len(sb.config.dnsList) > 0 || len(sb.config.dnsSearchList) > 0 || len(sb.config.dnsOptionsList) > 0 {
		var (
			err            error
			dnsList        = resolvconf.GetNameservers(currRC.Content)
			dnsSearchList  = resolvconf.GetSearchDomains(currRC.Content)
			dnsOptionsList = resolvconf.GetOptions(currRC.Content)
		)
		if len(sb.config.dnsList) > 0 {
			dnsList = sb.config.dnsList
		}
		if len(sb.config.dnsSearchList) > 0 {
			dnsSearchList = sb.config.dnsSearchList
		}
		if len(sb.config.dnsOptionsList) > 0 {
			dnsOptionsList = sb.config.dnsOptionsList
		}
		newRC, err = resolvconf.Build(sb.config.resolvConfPath, dnsList, dnsSearchList, dnsOptionsList)
		if err != nil {
			return err
		}
	} else {
		// Replace any localhost/127.* (at this point we have no info about ipv6, pass it as true)
		if newRC, err = resolvconf.FilterResolvDNS(currRC.Content, true); err != nil {
			return err
		}
		// No contention on container resolv.conf file at sandbox creation
		if err := ioutil.WriteFile(sb.config.resolvConfPath, newRC.Content, filePerm); err != nil {
			return types.InternalErrorf("failed to write unhaltered resolv.conf file content when setting up dns for sandbox %s: %v", sb.ID(), err)
		}
	}

	// Write hash
	if err := ioutil.WriteFile(sb.config.resolvConfHashFile, []byte(newRC.Hash), filePerm); err != nil {
		return types.InternalErrorf("failed to write resolv.conf hash file when setting up dns for sandbox %s: %v", sb.ID(), err)
	}

	return nil
}

func (sb *sandbox) updateDNS(ipv6Enabled bool) error {
	var (
		currHash string
		hashFile = sb.config.resolvConfHashFile
	)

	if len(sb.config.dnsList) > 0 || len(sb.config.dnsSearchList) > 0 || len(sb.config.dnsOptionsList) > 0 {
		return nil
	}

	currRC, err := resolvconf.GetSpecific(sb.config.resolvConfPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		h, err := ioutil.ReadFile(hashFile)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else {
			currHash = string(h)
		}
	}

	if currHash != "" && currHash != currRC.Hash {
		// Seems the user has changed the container resolv.conf since the last time
		// we checked so return without doing anything.
		log.Infof("Skipping update of resolv.conf file with ipv6Enabled: %t because file was touched by user", ipv6Enabled)
		return nil
	}

	// replace any localhost/127.* and remove IPv6 nameservers if IPv6 disabled.
	newRC, err := resolvconf.FilterResolvDNS(currRC.Content, ipv6Enabled)
	if err != nil {
		return err
	}

	// for atomic updates to these files, use temporary files with os.Rename:
	dir := path.Dir(sb.config.resolvConfPath)
	tmpHashFile, err := ioutil.TempFile(dir, "hash")
	if err != nil {
		return err
	}
	tmpResolvFile, err := ioutil.TempFile(dir, "resolv")
	if err != nil {
		return err
	}

	// Change the perms to filePerm (0644) since ioutil.TempFile creates it by default as 0600
	if err := os.Chmod(tmpResolvFile.Name(), filePerm); err != nil {
		return err
	}

	// write the updates to the temp files
	if err = ioutil.WriteFile(tmpHashFile.Name(), []byte(newRC.Hash), filePerm); err != nil {
		return err
	}
	if err = ioutil.WriteFile(tmpResolvFile.Name(), newRC.Content, filePerm); err != nil {
		return err
	}

	// rename the temp files for atomic replace
	if err = os.Rename(tmpHashFile.Name(), hashFile); err != nil {
		return err
	}
	return os.Rename(tmpResolvFile.Name(), sb.config.resolvConfPath)
}

// Embedded DNS server has to be enabled for this sandbox. Rebuild the container's
// resolv.conf by doing the follwing
// - Save the external name servers in resolv.conf in the sandbox
// - Add only the embedded server's IP to container's resolv.conf
// - If the embedded server needs any resolv.conf options add it to the current list
func (sb *sandbox) rebuildDNS() error {
	currRC, err := resolvconf.GetSpecific(sb.config.resolvConfPath)
	if err != nil {
		return err
	}

	// localhost entries have already been filtered out from the list
	sb.extDNS = resolvconf.GetNameservers(currRC.Content)

	var (
		dnsList        = []string{sb.resolver.NameServer()}
		dnsOptionsList = resolvconf.GetOptions(currRC.Content)
		dnsSearchList  = resolvconf.GetSearchDomains(currRC.Content)
	)

	// Resolver returns the options in the format resolv.conf expects
	dnsOptionsList = append(dnsOptionsList, sb.resolver.ResolverOptions()...)

	dir := path.Dir(sb.config.resolvConfPath)
	tmpResolvFile, err := ioutil.TempFile(dir, "resolv")
	if err != nil {
		return err
	}

	// Change the perms to filePerm (0644) since ioutil.TempFile creates it by default as 0600
	if err := os.Chmod(tmpResolvFile.Name(), filePerm); err != nil {
		return err
	}

	_, err = resolvconf.Build(tmpResolvFile.Name(), dnsList, dnsSearchList, dnsOptionsList)
	if err != nil {
		return err
	}

	return os.Rename(tmpResolvFile.Name(), sb.config.resolvConfPath)
}

// joinLeaveStart waits to ensure there are no joins or leaves in progress and
// marks this join/leave in progress without race
func (sb *sandbox) joinLeaveStart() {
	sb.Lock()
	defer sb.Unlock()

	for sb.joinLeaveDone != nil {
		joinLeaveDone := sb.joinLeaveDone
		sb.Unlock()

		select {
		case <-joinLeaveDone:
		}

		sb.Lock()
	}

	sb.joinLeaveDone = make(chan struct{})
}

// joinLeaveEnd marks the end of this join/leave operation and
// signals the same without race to other join and leave waiters
func (sb *sandbox) joinLeaveEnd() {
	sb.Lock()
	defer sb.Unlock()

	if sb.joinLeaveDone != nil {
		close(sb.joinLeaveDone)
		sb.joinLeaveDone = nil
	}
}

// OptionHostname function returns an option setter for hostname option to
// be passed to NewSandbox method.
func OptionHostname(name string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.hostName = name
	}
}

// OptionDomainname function returns an option setter for domainname option to
// be passed to NewSandbox method.
func OptionDomainname(name string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.domainName = name
	}
}

// OptionHostsPath function returns an option setter for hostspath option to
// be passed to NewSandbox method.
func OptionHostsPath(path string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.hostsPath = path
	}
}

// OptionOriginHostsPath function returns an option setter for origin hosts file path
// tbeo  passed to NewSandbox method.
func OptionOriginHostsPath(path string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.originHostsPath = path
	}
}

// OptionExtraHost function returns an option setter for extra /etc/hosts options
// which is a name and IP as strings.
func OptionExtraHost(name string, IP string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.extraHosts = append(sb.config.extraHosts, extraHost{name: name, IP: IP})
	}
}

// OptionParentUpdate function returns an option setter for parent container
// which needs to update the IP address for the linked container.
func OptionParentUpdate(cid string, name, ip string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.parentUpdates = append(sb.config.parentUpdates, parentUpdate{cid: cid, name: name, ip: ip})
	}
}

// OptionResolvConfPath function returns an option setter for resolvconfpath option to
// be passed to net container methods.
func OptionResolvConfPath(path string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.resolvConfPath = path
	}
}

// OptionOriginResolvConfPath function returns an option setter to set the path to the
// origin resolv.conf file to be passed to net container methods.
func OptionOriginResolvConfPath(path string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.originResolvConfPath = path
	}
}

// OptionDNS function returns an option setter for dns entry option to
// be passed to container Create method.
func OptionDNS(dns string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.dnsList = append(sb.config.dnsList, dns)
	}
}

// OptionDNSSearch function returns an option setter for dns search entry option to
// be passed to container Create method.
func OptionDNSSearch(search string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.dnsSearchList = append(sb.config.dnsSearchList, search)
	}
}

// OptionDNSOptions function returns an option setter for dns options entry option to
// be passed to container Create method.
func OptionDNSOptions(options string) SandboxOption {
	return func(sb *sandbox) {
		sb.config.dnsOptionsList = append(sb.config.dnsOptionsList, options)
	}
}

// OptionUseDefaultSandbox function returns an option setter for using default sandbox to
// be passed to container Create method.
func OptionUseDefaultSandbox() SandboxOption {
	return func(sb *sandbox) {
		sb.config.useDefaultSandBox = true
	}
}

// OptionUseExternalKey function returns an option setter for using provided namespace
// instead of creating one.
func OptionUseExternalKey() SandboxOption {
	return func(sb *sandbox) {
		sb.config.useExternalKey = true
	}
}

// OptionGeneric function returns an option setter for Generic configuration
// that is not managed by libNetwork but can be used by the Drivers during the call to
// net container creation method. Container Labels are a good example.
func OptionGeneric(generic map[string]interface{}) SandboxOption {
	return func(sb *sandbox) {
		sb.config.generic = generic
	}
}

func (eh epHeap) Len() int { return len(eh) }

func (eh epHeap) Less(i, j int) bool {
	var (
		cip, cjp int
		ok       bool
	)

	ci, _ := eh[i].getSandbox()
	cj, _ := eh[j].getSandbox()

	epi := eh[i]
	epj := eh[j]

	if epi.endpointInGWNetwork() {
		return false
	}

	if epj.endpointInGWNetwork() {
		return true
	}

	if ci != nil {
		cip, ok = ci.epPriority[eh[i].ID()]
		if !ok {
			cip = 0
		}
	}

	if cj != nil {
		cjp, ok = cj.epPriority[eh[j].ID()]
		if !ok {
			cjp = 0
		}
	}

	if cip == cjp {
		return eh[i].network.Name() < eh[j].network.Name()
	}

	return cip > cjp
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

func createBasePath(dir string) error {
	return os.MkdirAll(dir, dirPerm)
}

func createFile(path string) error {
	var f *os.File

	dir, _ := filepath.Split(path)
	err := createBasePath(dir)
	if err != nil {
		return err
	}

	f, err = os.Create(path)
	if err == nil {
		f.Close()
	}

	return err
}

func copyFile(src, dst string) error {
	sBytes, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, sBytes, filePerm)
}

package overlay

import (
	"net"
	"testing"
	"time"

	"github.com/docker/libnetwork/driverapi"
	_ "github.com/docker/libnetwork/testutils"
)

type driverTester struct {
	t *testing.T
	d *driver
}

const testNetworkType = "overlay"

func setupDriver(t *testing.T) *driverTester {
	dt := &driverTester{t: t}
	if err := Init(dt, nil); err != nil {
		t.Fatal(err)
	}

	if err := dt.d.configure(); err != nil {
		t.Fatal(err)
	}

	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		t.Fatal(err)
	}
	addrs, err := iface.Addrs()
	if err != nil || len(addrs) == 0 {
		t.Fatal(err)
	}
	data := driverapi.NodeDiscoveryData{
		Address: addrs[0].String(),
		Self:    true,
	}
	dt.d.DiscoverNew(driverapi.NodeDiscovery, data)
	return dt
}

func cleanupDriver(t *testing.T, dt *driverTester) {
	ch := make(chan struct{})
	go func() {
		Fini(dt.d)
		close(ch)
	}()

	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		t.Fatal("test timed out because Fini() did not return on time")
	}
}

func (dt *driverTester) RegisterDriver(name string, drv driverapi.Driver,
	cap driverapi.Capability) error {
	if name != testNetworkType {
		dt.t.Fatalf("Expected driver register name to be %q. Instead got %q",
			testNetworkType, name)
	}

	if _, ok := drv.(*driver); !ok {
		dt.t.Fatalf("Expected driver type to be %T. Instead got %T",
			&driver{}, drv)
	}

	dt.d = drv.(*driver)
	return nil
}

func TestOverlayInit(t *testing.T) {
	if err := Init(&driverTester{t: t}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestOverlayFiniWithoutConfig(t *testing.T) {
	dt := &driverTester{t: t}
	if err := Init(dt, nil); err != nil {
		t.Fatal(err)
	}

	cleanupDriver(t, dt)
}

func TestOverlayNilConfig(t *testing.T) {
	dt := &driverTester{t: t}
	if err := Init(dt, nil); err != nil {
		t.Fatal(err)
	}

	if err := dt.d.configure(); err != nil {
		t.Fatal(err)
	}

	cleanupDriver(t, dt)
}

func TestOverlayConfig(t *testing.T) {
	dt := setupDriver(t)

	time.Sleep(1 * time.Second)

	d := dt.d
	if d.notifyCh == nil {
		t.Fatal("Driver notify channel wasn't initialzed after Config method")
	}

	if d.exitCh == nil {
		t.Fatal("Driver serfloop exit channel wasn't initialzed after Config method")
	}

	if d.serfInstance == nil {
		t.Fatal("Driver serfinstance  hasn't been initialized after Config method")
	}

	cleanupDriver(t, dt)
}

func TestOverlayType(t *testing.T) {
	dt := &driverTester{t: t}
	if err := Init(dt, nil); err != nil {
		t.Fatal(err)
	}

	if dt.d.Type() != testNetworkType {
		t.Fatalf("Expected Type() to return %q. Instead got %q", testNetworkType,
			dt.d.Type())
	}
}

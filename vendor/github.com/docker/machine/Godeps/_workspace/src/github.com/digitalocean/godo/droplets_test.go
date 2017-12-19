package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestAction_DropletsServiceOpImplementsActionService(t *testing.T) {
	if !Implements((*DropletsService)(nil), new(DropletsServiceOp)) {
		t.Error("DropletsServiceOp does not implement DropletsService")
	}
}

func TestDroplets_ListDroplets(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"droplets": [{"id":1},{"id":2}]}`)
	})

	droplets, _, err := client.Droplets.List(nil)
	if err != nil {
		t.Errorf("Droplets.List returned error: %v", err)
	}

	expected := []Droplet{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(droplets, expected) {
		t.Errorf("Droplets.List returned %+v, expected %+v", droplets, expected)
	}
}

func TestDroplets_ListDropletsMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		dr := dropletsRoot{
			Droplets: []Droplet{
				Droplet{ID: 1},
				Droplet{ID: 2},
			},
			Links: &Links{
				Pages: &Pages{Next: "http://example.com/v2/droplets/?page=2"},
			},
		}

		b, err := json.Marshal(dr)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprint(w, string(b))
	})

	_, resp, err := client.Droplets.List(nil)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 1)
}

func TestDroplets_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
	{
		"droplets": [{"id":1},{"id":2}],
		"links":{
			"pages":{
				"next":"http://example.com/v2/droplets/?page=3",
				"prev":"http://example.com/v2/droplets/?page=1",
				"last":"http://example.com/v2/droplets/?page=3",
				"first":"http://example.com/v2/droplets/?page=1"
			}
		}
	}`

	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, jBlob)
	})

	opt := &ListOptions{Page: 2}
	_, resp, err := client.Droplets.List(opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestDroplets_GetDroplet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"droplet":{"id":12345}}`)
	})

	droplets, _, err := client.Droplets.Get(12345)
	if err != nil {
		t.Errorf("Droplet.Get returned error: %v", err)
	}

	expected := &DropletRoot{Droplet: &Droplet{ID: 12345}}
	if !reflect.DeepEqual(droplets, expected) {
		t.Errorf("Droplets.Get returned %+v, expected %+v", droplets, expected)
	}
}

func TestDroplets_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &DropletCreateRequest{
		Name:   "name",
		Region: "region",
		Size:   "size",
		Image:  "1",
	}

	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		v := new(DropletCreateRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, createRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, createRequest)
		}

		fmt.Fprintf(w, `{"droplet":{"id":1}, "links":{"actions": [{"id": 1, "href": "http://example.com", "rel": "create"}]}}`)
	})

	root, resp, err := client.Droplets.Create(createRequest)
	if err != nil {
		t.Errorf("Droplets.Create returned error: %v", err)
	}

	if id := root.Droplet.ID; id != 1 {
		t.Errorf("expected id '%d', received '%d'", 1, id)
	}

	if a := resp.Links.Actions[0]; a.ID != 1 {
		t.Errorf("expected action id '%d', received '%d'", 1, a.ID)
	}
}

func TestDroplets_Destroy(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Droplets.Delete(12345)
	if err != nil {
		t.Errorf("Droplet.Delete returned error: %v", err)
	}
}

func TestNetwork_String(t *testing.T) {
	network := &Network{
		IPAddress: "192.168.1.2",
		Netmask:   "255.255.255.0",
		Gateway:   "192.168.1.1",
	}

	stringified := network.String()
	expected := `godo.Network{IPAddress:"192.168.1.2", Netmask:"255.255.255.0", Gateway:"192.168.1.1", Type:""}`
	if expected != stringified {
		t.Errorf("Distribution.String returned %+v, expected %+v", stringified, expected)
	}

}

func TestDroplet_String(t *testing.T) {

	region := &Region{
		Slug:      "region",
		Name:      "Region",
		Sizes:     []string{"1", "2"},
		Available: true,
	}

	image := &Image{
		ID:           1,
		Name:         "Image",
		Distribution: "Ubuntu",
		Slug:         "image",
		Public:       true,
		Regions:      []string{"one", "two"},
	}

	size := &Size{
		Slug:         "size",
		PriceMonthly: 123,
		PriceHourly:  456,
		Regions:      []string{"1", "2"},
	}
	network := &Network{
		IPAddress: "192.168.1.2",
		Netmask:   "255.255.255.0",
		Gateway:   "192.168.1.1",
	}
	networks := &Networks{
		V4: []Network{*network},
	}

	droplet := &Droplet{
		ID:          1,
		Name:        "droplet",
		Memory:      123,
		Vcpus:       456,
		Disk:        789,
		Region:      region,
		Image:       image,
		Size:        size,
		BackupIDs:   []int{1},
		SnapshotIDs: []int{1},
		ActionIDs:   []int{1},
		Locked:      false,
		Status:      "active",
		Networks:    networks,
	}

	stringified := droplet.String()
	expected := `godo.Droplet{ID:1, Name:"droplet", Memory:123, Vcpus:456, Disk:789, Region:godo.Region{Slug:"region", Name:"Region", Sizes:["1" "2"], Available:true}, Image:godo.Image{ID:1, Name:"Image", Distribution:"Ubuntu", Slug:"image", Public:true, Regions:["one" "two"]}, Size:godo.Size{Slug:"size", Memory:0, Vcpus:0, Disk:0, PriceMonthly:123, PriceHourly:456, Regions:["1" "2"]}, BackupIDs:[1], SnapshotIDs:[1], Locked:false, Status:"active", Networks:godo.Networks{V4:[godo.Network{IPAddress:"192.168.1.2", Netmask:"255.255.255.0", Gateway:"192.168.1.1", Type:""}]}, ActionIDs:[1]}`
	if expected != stringified {
		t.Errorf("Droplet.String returned %+v, expected %+v", stringified, expected)
	}
}

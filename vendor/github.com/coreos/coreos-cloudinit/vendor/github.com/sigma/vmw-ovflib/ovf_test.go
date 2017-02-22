package ovf

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var data_vsphere = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Environment
     xmlns="http://schemas.dmtf.org/ovf/environment/1"
     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
     xmlns:oe="http://schemas.dmtf.org/ovf/environment/1"
     xmlns:ve="http://www.vmware.com/schema/ovfenv"
     oe:id=""
     ve:vCenterId="vm-12345">
   <PlatformSection>
      <Kind>VMware ESXi</Kind>
      <Version>5.5.0</Version>
      <Vendor>VMware, Inc.</Vendor>
      <Locale>en</Locale>
   </PlatformSection>
   <PropertySection>
         <Property oe:key="foo" oe:value="42"/>
         <Property oe:key="bar" oe:value="0"/>
   </PropertySection>
   <ve:EthernetAdapterSection>
      <ve:Adapter ve:mac="00:00:00:00:00:00" ve:network="foo" ve:unitNumber="7"/>
   </ve:EthernetAdapterSection>
</Environment>`)

var data_vapprun = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Environment xmlns="http://schemas.dmtf.org/ovf/environment/1"
     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
     xmlns:oe="http://schemas.dmtf.org/ovf/environment/1"
     oe:id="CoreOS-vmw">
   <PlatformSection>
      <Kind>vapprun</Kind>
      <Version>1.0</Version>
      <Vendor>VMware, Inc.</Vendor>
      <Locale>en_US</Locale>
   </PlatformSection>
   <PropertySection>
      <Property oe:key="foo" oe:value="42"/>
      <Property oe:key="bar" oe:value="0"/>
      <Property oe:key="guestinfo.user_data.url" oe:value="https://gist.githubusercontent.com/sigma/5a64aac1693da9ca70d2/raw/plop.yaml"/>
      <Property oe:key="guestinfo.user_data.doc" oe:value=""/>
      <Property oe:key="guestinfo.meta_data.url" oe:value=""/>
      <Property oe:key="guestinfo.meta_data.doc" oe:value=""/>
   </PropertySection>
</Environment>`)

func TestOvfEnvProperties(t *testing.T) {

	var testerFunc = func(env_str []byte) func() {
		return func() {
			env := ReadEnvironment(env_str)
			props := env.Properties

			var val string
			var ok bool
			Convey(`Property "foo"`, func() {
				val, ok = props["foo"]
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, "42")
			})

			Convey(`Property "bar"`, func() {
				val, ok = props["bar"]
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, "0")
			})
		}
	}

	Convey("With vAppRun environment", t, testerFunc(data_vapprun))
	Convey("With vSphere environment", t, testerFunc(data_vsphere))
}

func TestOvfEnvPlatform(t *testing.T) {
	Convey("With vSphere environment", t, func() {
		env := ReadEnvironment(data_vsphere)
		platform := env.Platform

		So(platform.Kind, ShouldEqual, "VMware ESXi")
		So(platform.Version, ShouldEqual, "5.5.0")
		So(platform.Vendor, ShouldEqual, "VMware, Inc.")
		So(platform.Locale, ShouldEqual, "en")
	})
}

func TestVappRunUserDataUrl(t *testing.T) {
	Convey("With vAppRun environment", t, func() {
		env := ReadEnvironment(data_vapprun)
		props := env.Properties

		var val string
		var ok bool

		val, ok = props["guestinfo.user_data.url"]
		So(ok, ShouldBeTrue)
		So(val, ShouldEqual, "https://gist.githubusercontent.com/sigma/5a64aac1693da9ca70d2/raw/plop.yaml")
	})
}

func TestInvalidData(t *testing.T) {
	Convey("With invalid data", t, func() {
		ReadEnvironment(append(data_vsphere, []byte("garbage")...))
	})
}

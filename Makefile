TARGETS := $(shell ls scripts | grep -vE 'clean|run|help|docs|release')

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m|sed 's/v7l//'` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

$(TARGETS): .dapper
	./.dapper $@

trash: .dapper
	./.dapper -m bind trash

trash-keep: .dapper
	./.dapper -m bind trash -k

deps: trash

build/initrd/.id: .dapper
	./.dapper prepare

run: build/initrd/.id .dapper
	./.dapper -m bind build-target
	./scripts/run

docs:
	./scripts/docs

shell-bind: .dapper
	./.dapper -m bind -s

clean:
	@./scripts/clean

release: .dapper release-build qcows

release-build:
	mkdir -p dist
	./.dapper release 2>&1 | tee dist/release.log

itest:
	mkdir -p dist
	./.dapper integration-test 2>&1 | tee dist/itest.log

qcows:
	cp dist/artifacts/rancheros.iso scripts/images/openstack/
	cd scripts/images/openstack && \
		NAME=openstack ../../../.dapper
	cd scripts/images/openstack && \
		APPEND="console=tty1 rancher.debug=true printk.devkmsg=on notsc clocksource=kvm-clock rancher.network.interfaces.eth0.ipv4ll rancher.cloud_init.datasources=[digitalocean] rancher.autologin=tty1 rancher.autologin=ttyS0" NAME=digitalocean ../../../.dapper
	cp ./scripts/images/openstack/dist/*.img dist/artifacts/

rpi: release
	# scripts/images/raspberry-pi-hypriot/dist/rancheros-raspberry-pi.zip
	cp dist/artifacts/rootfs_arm.tar.gz scripts/images/raspberry-pi-hypriot/
	cd scripts/images/raspberry-pi-hypriot/ \
		&& ../../../.dapper

help:
	@./scripts/help

.DEFAULT_GOAL := default

.PHONY: $(TARGETS)

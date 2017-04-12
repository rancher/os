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

release: release-build openstack

release-build:
	./.dapper release 2>&1 | tee dist/release.log

itest:
	./.dapper integration-test 2>&1 | tee dist/itest.log

openstack:
	cp dist/artifacts/rancheros.iso scripts/images/openstack/
	cd scripts/images/openstack && ../../../.dapper
	cp ./scripts/images/openstack/dist/*.img dist/

openstack-run:
	qemu-system-x86_64 -curses \
		-net nic -net user \
		-m 2048M \
		--hdc scripts/images/openstack/dist/rancheros-openstack.img

rpi: release
	# scripts/images/raspberry-pi-hypriot/dist/rancheros-raspberry-pi.zip
	cp dist/artifacts/rootfs_arm.tar.gz scripts/images/raspberry-pi-hypriot/
	cd scripts/images/raspberry-pi-hypriot/ \
		&& ../../../.dapper

help:
	@./scripts/help

.DEFAULT_GOAL := default

.PHONY: $(TARGETS)

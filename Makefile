TARGETS := $(shell ls scripts | grep -vE 'clean|run|help|release*|build-moby|run-moby')

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

build-moby:
	./scripts/build-moby

run-moby:
	./scripts/run-moby

shell-bind: .dapper
	./.dapper -m bind -s

clean:
	@./scripts/clean

release: .dapper release-build

release-build:
	mkdir -p dist
	./.dapper release 2>&1 | tee dist/release.log

rpi64: .dapper
	./scripts/release-rpi64

vmware: .dapper
	mkdir -p dist
	APPEND_SYSTEM_IMAGES="rancher/os-openvmtools:10.3.10-1" \
	./.dapper release-vmware 2>&1 | tee dist/release.log

hyperv: .dapper
	mkdir -p dist
	APPEND_SYSTEM_IMAGES="rancher/os-hypervvmtools:v4.14.128-rancher-1" \
	./.dapper release-hyperv 2>&1 | tee dist/release.log

azurebase: .dapper
	mkdir -p dist
	AZURE_SERVICE="true" \
	APPEND_SYSTEM_IMAGES="rancher/os-hypervvmtools:v4.14.128-rancher-1 rancher/os-waagent:v2.2.34-1" \
	./.dapper release-azurebase 2>&1 | tee dist/release.log

4glte: .dapper
	mkdir -p dist
	APPEND_SYSTEM_IMAGES="rancher/os-modemmanager:v1.6.4-1" \
	./.dapper release-4glte 2>&1 | tee dist/release.log

proxmoxve: .dapper
	mkdir -p dist
	PROXMOXVE_SERVICE="true" \
	APPEND_SYSTEM_IMAGES="rancher/os-qemuguestagent:v2.8.1-2" \
	./.dapper release-proxmoxve 2>&1 | tee dist/release.log

pingan: .dapper
	mkdir -p dist
	APPEND_SYSTEM_IMAGES="cnrancher/os-pingan-amc:v0.0.6-1" \
	./.dapper release-pingan 2>&1 | tee dist/release.log

help:
	@./scripts/help

.DEFAULT_GOAL := default

.PHONY: $(TARGETS)

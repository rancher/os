TARGETS := $(shell ls scripts | grep -vE 'clean|run|help|release*|deps')

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m|sed 's/v7l//'` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

$(TARGETS): .dapper
	./.dapper $@

pr-validation: .dapper
	ARCH="amd64" \
	KERNEL_URL_amd64="skip" \
	SKIP_BUILD="true" \
	./.dapper ci

deps: .dapper
	./.dapper -m bind deps

build/initrd/.id: .dapper
	./.dapper prepare

run: build/initrd/.id .dapper
	./.dapper -m bind build-target
	./scripts/run

shell-bind: .dapper
	./.dapper -m bind -s

clean:
	@./scripts/clean

release: .dapper release-build

release-build:
	mkdir -p dist
	./.dapper release

vmware: .dapper
	mkdir -p dist
	OS_FIRMWARE="false" \
	APPEND_SYSTEM_IMAGES="burmilla/os-openvmtools:12.5-1" \
	./.dapper release

hyperv: .dapper
	mkdir -p dist
	OS_FIRMWARE="false" \
	APPEND_SYSTEM_IMAGES="burmilla/os-hypervvmtools:6.12.114-burmilla-1" \
	./.dapper release

azurebase: .dapper
	mkdir -p dist
	AZURE_SERVICE="true" \
	OS_FIRMWARE="false" \
	APPEND_SYSTEM_IMAGES="burmilla/os-hypervvmtools:6.12.114-burmilla-1 burmilla/os-waagent:v2.2.49.2-1" \
	./.dapper release

proxmoxve: .dapper
	mkdir -p dist
	PROXMOXVE_SERVICE="true" \
	OS_FIRMWARE="false" \
	APPEND_SYSTEM_IMAGES="burmilla/os-qemuguestagent:v7.2-1" \
	./.dapper release

help:
	@./scripts/help

.DEFAULT_GOAL := default

.PHONY: $(TARGETS)

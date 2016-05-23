TARGETS := $(shell ls scripts | grep -vE 'clean|run|help')

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m` > .dapper.tmp
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

build/initrd/.id:
	dapper prepare

run: build/initrd/.id
	dapper -m bind build-target
	./scripts/run

clean:
	@./scripts/clean

help:
	@./scripts/help

.DEFAULT_GOAL := default

.PHONY: $(TARGETS)

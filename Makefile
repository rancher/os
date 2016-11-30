TARGETS := $(shell ls scripts | grep -vE 'clean|run|help|docs')

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

help:
	@./scripts/help

.DEFAULT_GOAL := default

.PHONY: $(TARGETS)

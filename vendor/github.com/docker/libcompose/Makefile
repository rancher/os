.PHONY: all test validate-dco validate-gofmt validate build

LIBCOMPOSE_ENVS := \
	-e OS_PLATFORM_ARG \
	-e OS_ARCH_ARG \
	-e DOCKER_TEST_HOST \
	-e TESTFLAGS

# (default to no bind mount if DOCKER_HOST is set)
BIND_DIR := $(if $(DOCKER_HOST),,bundles)
LIBCOMPOSE_MOUNT := $(if $(BIND_DIR),-v "$(CURDIR)/$(BIND_DIR):/go/src/github.com/docker/libcompose/$(BIND_DIR)")

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
LIBCOMPOSE_IMAGE := libcompose-dev$(if $(GIT_BRANCH),:$(GIT_BRANCH))

DOCKER_RUN_LIBCOMPOSE := docker run --rm -it --privileged $(LIBCOMPOSE_ENVS) $(LIBCOMPOSE_MOUNT) "$(LIBCOMPOSE_IMAGE)"

default: binary

all: build

binary: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh binary

test: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh binary test-unit test-integration

test-unit: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh binary test-unit

test-integration: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh binary test-integration

validate-dco: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh validate-dco

validate-gofmt: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh validate-gofmt

validate-lint: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh validate-lint

validate-vet: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh validate-vet

validate-git-marks: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh validate-git-marks

validate: build
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh validate-dco validate-git-marks validate-gofmt validate-lint validate-vet

shell: build
	$(DOCKER_RUN_LIBCOMPOSE) bash

# Build the docker image, should be prior almost any other goals
build: bundles
	docker build -t "$(LIBCOMPOSE_IMAGE)" .

bundles:
	mkdir bundles

clean: 
	$(DOCKER_RUN_LIBCOMPOSE) ./script/make.sh clean


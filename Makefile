.DEFAULT_GOAL := build
REPO?=rancher/os
TAG?=dev
IMAGE=${REPO}:${TAG}

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-$$(uname -s)-$$(uname -m) > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

.PHONY: validate
ci: .dapper
	./.dapper ci

.PHONY: clean
clean:
	rm -rf build

.PHONY: build
build:
	docker build \
		--build-arg CACHEBUST=${CACHEBUST} \
		--build-arg IMAGE_TAG=${TAG} \
		--build-arg IMAGE_REPO=${REPO} \
		-t ${IMAGE} .

.PHONY: push
push: build
	docker push ${IMAGE}

.PHONY: iso
iso: build
	./ros-image-build ${IMAGE} iso
	@echo "INFO: ISO available at build/output.iso"

.PHONY: qcow
qcow: build
	./ros-image-build ${IMAGE} qcow
	@echo "INFO: QCOW image available at build/output.qcow.gz"

.PHONY: ami-%
ami-%:
	AWS_DEFAULT_REGION=$* ./ros-image-build ${IMAGE} ami

.PHONY: ami
ami:
	./ros-image-build ${IMAGE} ami

.PHONY: run
run:
	./scripts/run

all-amis: \
	ami-us-west-1 \
	ami-us-west-2
	#ami-ap-east-1 \
	#ami-ap-northeast-1 \
	#ami-ap-northeast-2 \
	#ami-ap-northeast-3 \
	#ami-ap-southeast-1 \
	#ami-ap-southeast-2 \
	#ami-ca-central-1 \
	#ami-eu-central-1 \
	#ami-eu-south-1 \
	#ami-eu-west-1 \
	#ami-eu-west-2 \
	#ami-eu-west-3 \
	#ami-me-south-1 \
	#ami-sa-east-1 \
	#ami-us-east-1 \
	#ami-us-east-2 \

.DEFAULT_GOAL := iso
REPO=ibuildthecloud/test
LABEL=latest
IMAGE=${REPO}:${LABEL}
TOOLS=${IMAGE}-tools

.PHONY: build
build:
	docker build \
		--build-arg CACHEBUST="${CACHEBUST}" \
		--build-arg OS_LABEL=${LABEL} \
		--build-arg OS_REPO=${REPO} \
		-t ${IMAGE} .

.PHONY: push
push: build
	docker push ${IMAGE}

.PHONY: tools
tools:
	docker build -t ${TOOLS} --target tools .

.PHONY: iso
iso: tools build
	mkdir -p build
	rm -f build/iso-container
	docker run -v /var/run:/var/run -it --cidfile=build/iso-container ${TOOLS} makeiso ${IMAGE}
	docker cp $$(cat build/iso-container):/output.iso build/output.iso
	docker rm -fv $$(cat build/iso-container)
	rm -f build/iso-container

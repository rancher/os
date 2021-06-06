.DEFAULT_GOAL := iso
IMAGE=ibuildthecloud/test
TOOLS=${IMAGE}-tools

.PHONY: build
build:
	docker build -t ${IMAGE} .

.PHONY: final-build
final-build:
	docker build --build-arg FINALIZE=true -t ${IMAGE} .

.PHONY: push
push: build
	docker push ${IMAGE}

.PHONY: tools
tools:
	docker build -t ${TOOLS} --target tools .

.PHONY: iso
iso: tools final-build
	mkdir -p build
	rm -f build/iso-container
	docker run -v /var/run:/var/run -it --cidfile=build/iso-container ${TOOLS} makeiso ${IMAGE}
	docker cp $$(cat build/iso-container):/output.iso build/output.iso
	docker rm -fv $$(cat build/iso-container)
	rm -f build/iso-container

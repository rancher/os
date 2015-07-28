
DOCKER_BINARY_URL := https://github.com/rancher/docker/releases/download/v1.7.1-ros-1/docker-1.7.1


FORCE_PULL   := 0
DOCKER_BINARY := $(shell basename $(DOCKER_BINARY_URL))
pwd := $(shell pwd)
include scripts/build-common
include scripts/version


compile: bin/rancheros

all: clean ros-build-base package


ros-build-base:
	docker build -t ros-build-base -f Dockerfile.base .

ros-build:
	docker build -t ros-build .

docker-run: ros-build
	docker rm -fv ros-build > /dev/null 2>&1 || :
	docker run -v /var/run/docker.sock:/var/run/docker.sock --name=ros-build -i ros-build


CD := $(BUILD)/cd

assets bin $(DIST)/artifacts $(CD)/boot/isolinux:
	mkdir -p $@


ifdef CONTAINED


assets/$(DOCKER_BINARY): assets
	cd assets && curl -OL "$(DOCKER_BINARY_URL)"

assets/docker: assets/$(DOCKER_BINARY)
	mv assets/$(DOCKER_BINARY) $@
	chmod +x $@

bin/rancheros: bin
	godep go build -tags netgo -ldflags "-X github.com/rancherio/os/config.VERSION $(VERSION) -linkmode external -extldflags -static" -o $@
	strip --strip-all $@

copy-images:
	./scripts/copy-images
.PHONY: copy-images

$(DIST)/artifacts/vmlinuz: $(DIST)/artifacts copy-images
	mv $(BUILD)/kernel/vmlinuz $@


INITRD_DIR := $(BUILD)/initrd

$(INITRD_DIR)/images.tar: bin/rancheros
	ln -sf bin/rancheros ./ros
	for i in `./ros c images -i os-config.yml`; do [ "$(FORCE_PULL)" != "1" ] && docker inspect $$i >/dev/null 2>&1 || docker pull $$i; done
	docker save `./ros c images -i os-config.yml` > $@


$(DIST)/artifacts/initrd: $(DIST)/artifacts bin/rancheros assets/docker copy-images $(INITRD_DIR)/images.tar
	mv $(BUILD)/kernel/lib $(INITRD_DIR)
	mv assets/docker       $(INITRD_DIR)
	cp os-config.yml       $(INITRD_DIR)
	cp bin/rancheros       $(INITRD_DIR)/init
	cd $(INITRD_DIR) && find | cpio -H newc -o | lzma -c > $@

$(DIST)/artifacts/rancheros.iso: $(DIST)/artifacts/initrd $(CD)/boot/isolinux
	cp $(DIST)/artifacts/initrd                   $(CD)/boot
	cp $(DIST)/artifacts/vmlinuz                  $(CD)/boot
	cp scripts/isolinux.cfg                       $(CD)/boot/isolinux
	cp /usr/lib/ISOLINUX/isolinux.bin             $(CD)/boot/isolinux
	cp /usr/lib/syslinux/modules/bios/ldlinux.c32 $(CD)/boot/isolinux
	cd $(CD) && xorriso -publisher "Rancher Labs, Inc." \
		-as mkisofs \
		-l -J -R -V "RancherOS" \
		-no-emul-boot -boot-load-size 4 -boot-info-table \
		-b boot/isolinux/isolinux.bin -c boot/isolinux/boot.cat \
		-isohybrid-mbr /usr/lib/ISOLINUX/isohdpfx.bin \
		-o $@ $(CD)

$(DIST)/artifacts/iso-checksums.txt: $(DIST)/artifacts/rancheros.iso
	cd $(DIST)/artifacts && for algo in 'sha256' 'md5'; do echo "$$algo: `$${algo}sum rancheros.iso`" >> $@; done

package: \
	$(DIST)/artifacts/initrd \
	$(DIST)/artifacts/vmlinuz \
	$(DIST)/artifacts/rancheros.iso \
	$(DIST)/artifacts/iso-checksums.txt


else


bin/rancheros:
	@echo make $@ | make docker-run
	docker cp ros-build:/go/src/github.com/rancherio/os/$@ $(dir $@)
.PHONY: bin/rancheros

package:
	@echo make FORCE_PULL=$(FORCE_PULL) $@ | make docker-run
	docker cp ros-build:/go/src/github.com/rancherio/os/bin/rancheros bin
	docker cp ros-build:/go/src/github.com/rancherio/os/dist/artifacts dist


endif


version:
	@echo $(VERSION)

clean:
	rm -rf bin build dist gopath .dockerfile

.PHONY: all compile clean dist docker-run download package ros-build ros-build-base version

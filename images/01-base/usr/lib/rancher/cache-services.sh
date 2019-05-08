#!/bin/bash
set -e -x

root="/mnt/install"
cache="/var/lib/rancher/cache"
preload="/var/lib/rancher/preload/system-docker"

partition=$1
images=${@:2}
images_arr=(${images// / })

mount_directory() {
	sudo mkdir -p ${root}
	mount ${partition} ${root}
}

cache_services() {
	mkdir -p ${root}${cache}
	cp ${cache}/* ${root}${cache}
}

cache_images() {
	mkdir -p ${root}${preload}
	for i in "${images_arr[@]}"
	do
		system-docker pull $i
	done
	system-docker save ${images} | xz > ${root}${preload}/os-include.xz
}

mount_directory
cache_services
cache_images
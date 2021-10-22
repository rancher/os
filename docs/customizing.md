# Custom Images

RancherOS image can be easily remaster using a docker build.
For example, to add `cowsay` to RancherOS you would use the
following Dockerfile

## Docker image

```Dockerfile
FROM rancher/os2:v0.0.1-test01
RUN zypper install -y cowsay

# IMPORTANT: Setup rancheros-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
ARG IMAGE_REPO=norepo
ARG IMAGE_TAG=latest
RUN echo "IMAGE_REPO=${IMAGE_REPO}"          > /usr/lib/rancheros-release && \
    echo "IMAGE_TAG=${IMAGE_TAG}"           >> /usr/lib/rancheros-release && \
    echo "IMAGE=${IMAGE_REPO}:${IMAGE_TAG}" >> /usr/lib/rancheros-release
```

And then the following commands

```bash
docker build --build-arg IMAGE_REPO=myrepo/custom-build \
             --build-arg IMAGE_TAG=v1.1.1 \
             -t myrepo/custom-build:v1.1.1 .
docker push myrepo/custom-build:v1.1.1
```

## Bootable images

To create bootable images from the docker image you just created
run the below command

```bash
curl -o ros-image-build https://raw.githubusercontent.com/rancher/os2/main/ros-image-build
bash ros-image myrepo/custom-build:v1.1.1 qcow,iso,ami
```

The above command will create an ISO, a qcow image, and publish AMIs. You need not create all
three types and can change to comma seperated list to the types you care for.
#IMAGE_NAME := "cr.imroc.cc/library/cert-manager-webhook-dnspod"
IMAGE_NAME ?= "imroc/cert-manager-webhook-dnspod"
IMAGE_TAG ?= "latest"
IMG ?= $(IMAGE_NAME):$(IMAGE_TAG)

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

build:
	docker buildx build --platform=linux/amd64 -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push_chart:
	rm *.tgz
	helm package charts/cert-manager-webhook-dnspod
	helm push *.tgz oci://registry-1.docker.io/imroc
	rm *.tgz

#IMAGE_NAME := "cr.imroc.cc/library/cert-manager-webhook-dnspod"
IMAGE_NAME ?= "imroc/cert-manager-webhook-dnspod"
IMAGE_TAG ?= "latest"

build:
	docker buildx build --push --platform=linux/amd64 -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push_chart:
	rm *.tgz
	helm package charts/cert-manager-webhook-dnspod
	helm push *.tgz oci://registry-1.docker.io/imroc
	rm *.tgz

#IMAGE_NAME := "cr.imroc.cc/library/cert-manager-webhook-dnspod"
IMAGE_NAME := "imroc/cert-manager-webhook-dnspod"
IMAGE_TAG := "latest"

build:
	docker buildx build --platform=linux/amd64 -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

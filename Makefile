IMAGE_NAME ?= imroc/cert-manager-webhook-dnspod
IMAGE_TAG ?= latest
IMG ?= $(IMAGE_NAME):$(IMAGE_TAG)
PROJECT_NAME := cert-manager-webhook-dnspod

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

.PHONY: docker-buildx-push
docker-buildx-push: docker-buildx docker-push

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	$(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} .
.PHONY: docker-push
docker-push:
	$(CONTAINER_TOOL) push ${IMG}

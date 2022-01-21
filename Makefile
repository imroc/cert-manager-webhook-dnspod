#IMAGE_NAME := "cr.imroc.cc/library/cert-manager-webhook-dnspod"
IMAGE_NAME := "imroc/cert-manager-webhook-dnspod"
IMAGE_TAG := "latest"

build:
	docker build build --platform=linux/amd64 -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

build_fast:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/webhook -ldflags '-w -extldflags "-static"' .
	docker buildx build --push --platform=linux/amd64 -f fast.Dockerfile -t "$(IMAGE_NAME):$(IMAGE_TAG)" .
	rm ./bin/webhook

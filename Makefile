BINARY_NAME=app-restart-controller
IMG ?= ${BINARY_NAME}:latest

run:
	go run main.go

build:
	go build -o bin/${BINARY_NAME} main.go

docker:
	docker build -t ${IMG} .

install:
	kubectl apply -f config/rbac.yaml
	kubectl apply -f config/deploy.yaml

uninstall:
	kubectl delete -f config/deploy.yaml
	kubectl delete -f config/rbac.yaml

kind-load:
	kind load docker-image ${IMG} --name pod-monitor

kind-apply: docker kind-load install
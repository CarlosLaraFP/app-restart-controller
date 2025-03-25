APP_NAME=app-restart-controller
IMAGE_NAME=$(APP_NAME):latest

build:
	go build -o $(APP_NAME) main.go

run:
	go run main.go

test:
	go mod tidy
	go test ./... -v

kind-install:
	curl -Lo kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
	chmod +x kind
	sudo mv kind /usr/local/bin/kind

kind-create:
	kind create cluster --name app-restart

docker:
	docker build -t $(IMAGE_NAME) .

kind-load:
	kind load docker-image $(IMAGE_NAME) --name app-restart

helm-install:
	helm upgrade --install $(APP_NAME) ./chart --namespace controllers --create-namespace

log:
	kubectl describe deployment restart-me -n default
	kubectl get configmap my-config -o yaml

helm-uninstall:
	helm uninstall $(APP_NAME) --namespace controllers

kind-delete:
	kind delete cluster --name app-restart

kind-reset:
	kind delete cluster --name app-restart
	kind create cluster --name app-restart

deploy: kind-create docker kind-load helm-install

destroy: helm-uninstall kind-delete

clean:
	rm -f $(APP_NAME)

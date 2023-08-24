all: agent controller client

BUILD_TAGS := containers_image_openpgp,containers_image_docker_daemon_stub,containers_image_storage_stub

proto:
	@echo "Generating Go files"
	protoc --go_out=./pkg --go-grpc_out=./pkg pkg/proto/*.proto

agent: 
	@echo "Building agent..."
	go build -tags $(BUILD_TAGS) -o bin/k8s-image-warden-agent github.com/surik/k8s-image-warden/cmd/agent

controller: 
	@echo "Building controller..."
	go build -tags $(BUILD_TAGS) -o bin/k8s-image-warden-controller github.com/surik/k8s-image-warden/cmd/controller

client: 
	@echo "Building client..."
	go build -tags $(BUILD_TAGS) -o bin/kiwctl github.com/surik/k8s-image-warden/cmd/client

docker:
	@echo "Building images..."
	docker buildx build -f cmd/agent/Dockerfile -t k8s-image-warden-agent .
	docker buildx build -f cmd/controller/Dockerfile -t k8s-image-warden-controller .

install-deps:
	@echo "Install dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3
	go install gotest.tools/gotestsum@v1.10.1

lint: 
	@echo "Linting..."
	golangci-lint run

test:
	@echo "Testing..."
	@gotestsum -- -tags $(BUILD_TAGS) -timeout 30s -coverpkg=./... -coverprofile=cover.out ./...
	@go tool cover -func cover.out | grep total

.PHONY: all agent controller docker lint test install-deps
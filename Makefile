OUT?=./bin
BINARY_NAME?=$(OUT)/naavik
WORKING_DIR=$(shell pwd)
MAIN_PATH_NAAVIK=./main.go

# Image Build Args
DOCKER_REPO?=admiralproj
IMAGE?=$(DOCKER_REPO)/naavik
DOCKERFILE?=Dockerfile.naavik
GOVERSION=1.21
ISTIO_VERSION=1.21.0

setup-dev: setup-dependencies setup-format-tools setup-swagger

# install go 1.21
# set go 1.21 as default
# validate go version. If go version is less than 1.21, exit with error
install-go:
	brew install go@$(GOVERSION)
	brew unlink go && brew link --overwrite go@$(GOVERSION)
	@go version
	@go version | grep -q "go version go$(GOVERSION)" || (echo "Go version must be $(GOVERSION). Validate PATH is set properly." && exit 1)

# validate go version. If go version is less than 1.25, exit with error
setup-dependencies:
	@go version | grep -q "go version go$(GOVERSION)" || (echo "Go version must be $(GOVERSION). Run make setup-golang" && exit 1)
	go mod tidy

setup-format-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	golangci-lint --version
	go install github.com/segmentio/golines@v0.11.0
	golines --version
	go install github.com/onsi/ginkgo/v2/ginkgo                                                                         
	ginkgo version
	go install github.com/boumenot/gocover-cobertura@latest

setup-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest
	swag --version

generate-swagger:
	swag fmt
	swag init -g ./internal/server/swagger/docs.go -o ./internal/server/swagger
	
lint:
	golangci-lint run --fast --fix -c .golangci.yml

lint-ci: 
	golangci-lint run --fast -c .golangci.yml

format:
	go fmt ./...
	# for f in $(find $(pwd) -name '*.go'); do golines $f -w -m 200; done

run-local:
	go run $(MAIN_PATH_NAAVIK) --kube_config ~/.kube/config --log_level debug --log_color true --config_resolver=secret --state_checker=none --config_path ./config/config.yaml

test:
	ginkgo run -r -v -race --coverprofile=coverage.out -skip "Traffic config processing performance test"
	gocover-cobertura < coverage.out > coverage.xml

perf:
	ginkgo -v -race -focus "Traffic config processing performance test" -r 2>&1 | tee perf_output.log
	go run tests/perf/calc/calc_timetaken.go 2>&1 | tee -a perf_output.log

help:
	go run $(MAIN_PATH_NAAVIK) --help

build-mac:
	go build -o $(BINARY_NAME) -v $(MAIN_PATH_NAAVIK)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) -v $(MAIN_PATH_NAAVIK)

set-tag:
ifndef TAG
ifeq ($(BRANCH),master)
override TAG=latest
endif
endif
ifndef TAG
override TAG=$(SHA)
endif

docker-build: set-tag
	#NOTE: Assumes binary has already been built (admiral)
	docker build -t $(IMAGE):$(TAG) -f ./build/docker/$(DOCKERFILE) .


docker-publish:
ifndef DO_NOT_PUBLISH
ifndef PIPELINE_BUILD
	echo "$(DOCKER_PASS)" | docker login -u ${DOCKER_USERNAME} --password-stdin --storage-driver=overlay
endif
endif
ifeq ($(TAG),)
	echo "This is not a Tag/Release, skipping docker publish"
else
ifndef DO_NOT_PUBLISH
	docker push $(IMAGE):$(TAG) 
	docker pull $(IMAGE):$(TAG) 
endif
endif
#no tag set and its master branch, in this case publish `latest` tag
ifeq ($(TAG),)
ifeq ($(BRANCH),master)
	docker push $(IMAGE):latest 
	docker pull $(IMAGE):$(TAG) 
else
	echo "This is not master branch, skipping to publish 'latest' tag"
endif
endif
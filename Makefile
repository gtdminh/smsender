.PHONY: deps-install check-style test build build-with-docker docker-build

BUILD_EXECUTABLE := smsender
PACKAGES := $(shell go list ./smsender/...)

all: build

deps-install:
	@echo Getting dependencies using Glide
	go get -v -u github.com/Masterminds/glide
	glide install

vet:
	@echo Running go vet
	@go vet $(PACKAGES)

check-style: vet
	@echo Running go fmt
	$(eval GO_FMT_OUTPUT := $(shell go fmt $(PACKAGES)))
	@echo "$(GO_FMT_OUTPUT)"
	@if [ ! "$(GO_FMT_OUTPUT)" ]; then \
		echo "go fmt success"; \
	else \
		echo "go fmt failure"; \
		exit 1; \
	fi

test:
	@echo Testing
	@go test -race -v $(PACKAGES)

build: clean
	@echo Building app
	go build -o ./bin/$(BUILD_EXECUTABLE)

clean:
	@echo Cleaning up previous build data
	rm -f ./bin/$(BUILD_EXECUTABLE)

build-with-docker:
	@echo Building app with Docker
	docker run --rm -v $(PWD):/go/src/github.com/minchao/smsender -w /go/src/github.com/minchao/smsender golang sh -c "make deps-install build"

	cd webroot && make build-with-docker

docker-build: build-with-docker
	@echo Building Docker image
	docker build -t minchao/smsender-preview:latest .
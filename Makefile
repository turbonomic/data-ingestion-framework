BUILD_DIR=./build
SOURCE_DIRS=cmd pkg
PACKAGES=go list ./... | grep -v /vendor | grep -v /out
SHELL='/bin/bash'
REMOTE=github.ibm.com
USER=turbonomic
PROJECT=data-ingestion-framework
BINARY=turbodif
DEFAULT_VERSION=latest
TURBODIF_VERSION=8.15.4-SNAPSHOT
REMOTE_URL=$(shell git config --get remote.origin.url)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
REVISION=$(shell git show -s --format=%cd --date=format:'%Y%m%d%H%M%S000')


.DEFAULT_GOAL:=build

GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_TIME=$(shell date -R)
BUILD_TIMESTAMP=$(shell date +'%Y%m%d%H%M%S000')
PROJECT_PATH=$(REMOTE)/$(USER)/$(PROJECT)
VERSION=$(or $(TURBODIF_VERSION), $(DEFAULT_VERSION))
LDFLAGS='\
 -X "$(PROJECT_PATH)/version.GitCommit=$(GIT_COMMIT)" \
 -X "$(PROJECT_PATH)/version.BuildTime=$(BUILD_TIME)" \
 -X "$(PROJECT_PATH)/version.Version=$(VERSION)"'


LINUX_ARCH=amd64 arm64 ppc64le s390x

$(LINUX_ARCH): clean
	env GOOS=linux GOARCH=$@ go build -ldflags $(LDFLAGS) -o $(BUILD_DIR)/linux/$@/$(BINARY) ./cmd

product: $(LINUX_ARCH)

debug-product: clean
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -gcflags "-N -l" -o $(BUILD_DIR)/$(BINARY)_debug.linux ./cmd

build: clean
	go build -ldflags $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd

buildInfo:
			$(shell test -f git.properties && rm -rf git.properties)
			@echo 'turbo-version.remote.origin.url=$(REMOTE_URL)' >> git.properties
			@echo 'turbo-version.commit.id=$(GIT_COMMIT)' >> git.properties
			@echo 'turbo-version.branch=$(BRANCH)' >> git.properties
			@echo 'turbo-version.branch.version=$(VERSION)' >> git.properties
			@echo 'turbo-version.commit.time=$(REVISION)' >> git.properties
			@echo 'turbo-version.build.time=$(BUILD_TIMESTAMP)' >> git.properties

debug: clean
	go build -ldflags $(LDFLAGS) -gcflags "-N -l" -o $(BUILD_DIR)/$(BINARY).debug ./cmd

docker: product
	DOCKER_BUILDKIT=1 docker build -f build/Dockerfile -t turbonomic/turbodif .

test: clean
	@go test -v -race ./pkg/...

.PHONY: fmtcheck
fmtcheck:
	@gofmt -s -l $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

.PHONY: vet
vet:
	@go vet $(shell $(PACKAGES))

clean:
	@rm -rf $(BUILD_DIR)/$(BINARY)* ${BUILD_DIR}/linux


PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
REPO_NAME ?= icr.io/cpopen/turbonomic
multi-archs:
	env GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd
.PHONY: docker-buildx
docker-buildx:
	docker buildx create --name turbodif-builder
	- docker buildx use turbodif-builder
	- docker buildx build --platform=$(PLATFORMS) --label "git-commit=$(GIT_COMMIT)" --label "git-version=$(VERSION)" --provenance=false --push --tag $(REPO_NAME)/turbodif:$(VERSION) -f build/Dockerfile.multi-archs --build-arg VERSION=$(VERSION) .
	docker buildx rm turbodif-builder

check-upstream-dependencies:
	./hack/check_upstream_dependencies.sh

.PHONY:public-repo-update
public-repo-update:
	@if [[ "$(TURBODIF_VERSION)" =~ ^[0-9]+\.[0-9]+\.[0-9]+$$ ]] ; then \
		./scripts/travis/public_repo_update.sh ${TURBODIF_VERSION}; \
	fi

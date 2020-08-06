SOURCE_DIRS=.
PACKAGES:=go list ./... | grep -v /vendor
BIN=azureloganalytics
IMAGE=azure-loganalytics-go

.DEFAULT_GOAL := build

build:
	go build -o $(BIN) ./...

product: clean
	env GOOS=linux GOARCH=amd64 go build -o $(BIN).linux ./...

docker: product
	docker build -t turbonomic/$(IMAGE) --build-arg GIT_COMMIT=$(shell git rev-parse --short HEAD) .

.PHONY: fmtcheck
fmtcheck:
	@gofmt -l $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

.PHONY: vet
 vet:
	@go vet $(shell $(PACKAGES))

clean:
	@rm -rf $(BIN) $(BIN).linux


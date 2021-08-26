export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GO111MODULE=on
export GOFLAGS=-mod=vendor

BIN_DIR = $(CURDIR)/build/bin
export GOROOT=$(BIN_DIR)/go
export GOBIN = $(GOROOT)/bin
export PATH := $(GOBIN):$(PATH)

GO := $(GOBIN)/go

$(GO):
	hack/install-go.sh $(BIN_DIR)

format: $(GO)
	$(GO)fmt -w cmd/

build: $(GO) format
	$(GO) build -o build/bin/templating-device cmd/templating-device.go

clean:
	rm -rf deployment/* build/*

test:
	$(GO) test -v cmd/templating-device.go cmd/templating-device_test.go

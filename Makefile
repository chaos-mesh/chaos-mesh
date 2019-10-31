# Set DEBUGGER=1 to build debug symbols
LDFLAGS = $(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh)

# SET DOCKER_REGISTRY to change the docker registry
DOCKER_REGISTRY := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY),localhost:5000)

GOVER_MAJOR := $(shell go version | sed -E -e "s/.*go([0-9]+)[.]([0-9]+).*/\1/")
GOVER_MINOR := $(shell go version | sed -E -e "s/.*go([0-9]+)[.]([0-9]+).*/\2/")
GO111 := $(shell [ $(GOVER_MAJOR) -gt 1 ] || [ $(GOVER_MAJOR) -eq 1 ] && [ $(GOVER_MINOR) -ge 11 ]; echo $$?)
ifeq ($(GO111), 1)
$(error Please upgrade your Go compiler to 1.11 or higher version)
endif

# Enable GO111MODULE=on explicitly, disable it with GO111MODULE=off when necessary.
export GO111MODULE := on
GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)
GOENV  := GO15VENDOREXPERIMENT="1" CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go build
GOTEST := CGO_ENABLED=0 go test -v -cover

PACKAGE_LIST := go list ./... | grep -vE "pkg/client" | grep -vE "zz_generated"
PACKAGE_DIRECTORIES := $(PACKAGE_LIST) | sed 's|github.com/pingcap/chaos-operator/||'
FILES := $$(find $$($(PACKAGE_DIRECTORIES)) -name "*.go")
FAIL_ON_STDOUT := awk '{ print } END { if (NR > 0) { exit 1 } }'
TEST_COVER_PACKAGES:=go list ./pkg/... | grep -vE "pkg/client" | grep -vE "pkg/apis" | sed 's|github.com/pingcap/chaos-operator/|./|' | tr '\n' ','

default: build

docker-push: docker
	docker push "${DOCKER_REGISTRY}/pingcap/chaos-operator:latest"

docker: build
	docker build --tag "${DOCKER_REGISTRY}/pingcap/chaos-operator:latest" images/chaos-operator

build: controller-manager admission-controller

controller-manager:
	$(GO) -ldflags '$(LDFLAGS)' -o images/chaos-operator/bin/chaos-controller-manager cmd/controller-manager/*.go

admission-controller:
	$(GO) -ldflags '$(LDFLAGS)' -o images/chaos-operator/bin/chaos-admission-manager cmd/admission-controller/*.go

test:
	@echo "Run unit tests"
	@$(GOTEST) ./pkg/... -coverpkg=$$($(TEST_COVER_PACKAGES)) -coverprofile=coverage.txt -covermode=atomic && echo "\nUnit tests run successfully!"

check-all: check tidy check-shadow check-gosec staticcheck

check-setup:
	@which retool >/dev/null 2>&1 || go get github.com/twitchtv/retool
	@GO111MODULE=off retool sync

check: check-setup lint check-static

check-static:
	@ # Not running vet and fmt through metalinter becauase it ends up looking at vendor
	@echo "gofmt checking"
	gofmt -s -l -w $(FILES) 2>&1| $(FAIL_ON_STDOUT)
	@echo "go vet check"
	@go vet -all $$($(PACKAGE_LIST)) 2>&1
	@echo "mispell and ineffassign checking"
	CGO_ENABLED=0 retool do gometalinter.v2 --disable-all \
	  --enable misspell \
	  --enable ineffassign \
	  $$($(PACKAGE_DIRECTORIES))

# TODO: staticcheck is too slow currently
staticcheck:
	@echo "gometalinter staticcheck"
	CGO_ENABLED=0 retool do gometalinter.v2 --disable-all --deadline 120s \
	  --enable staticcheck \
	  $$($(PACKAGE_DIRECTORIES))

# TODO: errcheck is too slow currently
errcheck:
	@echo "gometalinter errcheck"
	CGO_ENABLED=0 retool do gometalinter.v2 --disable-all --deadline 120s \
	  --enable errcheck \
	  $$($(PACKAGE_DIRECTORIES))

# TODO: shadow check fails at the moment
check-shadow:
	@echo "go vet shadow checking"
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	@go vet -vettool=$(which shadow) $$($(PACKAGE_LIST))

lint:
	@echo "linting"
	CGO_ENABLED=0 retool do revive -formatter friendly -config revive.toml $$($(PACKAGE_LIST))


tidy:
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff --quiet go.mod go.sum

check-gosec:
	@echo "security checking"
	CGO_ENABLED=0 retool do gosec $$($(PACKAGE_DIRECTORIES))


.PHONY: check check-all build tidy admission-controller
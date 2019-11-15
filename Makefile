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
GO     := $(GOENV) go
GOTEST := CGO_ENABLED=0 go test -v -cover

PACKAGE_LIST := go list ./... | grep -vE "pkg/client" | grep -vE "zz_generated"
PACKAGE_DIRECTORIES := $(PACKAGE_LIST) | sed 's|github.com/pingcap/chaos-operator/||'
FILES := $$(find $$($(PACKAGE_DIRECTORIES)) -name "*.go")
FAIL_ON_STDOUT := awk '{ print } END { if (NR > 0) { exit 1 } }'
TEST_COVER_PACKAGES:=go list ./pkg/... | grep -vE "pkg/client" | grep -vE "pkg/apis" | sed 's|github.com/pingcap/chaos-operator/|./|' | tr '\n' ','

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: yaml build image

build: chaosdaemon manager chaosfs

# Run tests
test: generate fmt vet manifests
	$(GO) test ./... -coverprofile cover.out

# Build chaos-daemon binary
chaosdaemon: generate fmt vet
	$(GO) build -ldflags '$(LDFLAGS)' -o images/chaos-operator/bin/chaos-daemon ./cmd/chaos-daemon/main.go

# Build manager binary
manager: generate fmt vet
	$(GO) build -ldflags '$(LDFLAGS)' -o images/chaos-operator/bin/chaos-controller-manager ./cmd/controller-manager/*.go

chaosfs: generate fmt vet
	$(GO) build -ldflags '$(LDFLAGS)' -o images/chaosfs/bin/chaosfs ./cmd/chaosfs/*.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	$(GO) run ./cmd/controller-manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f manifests/
	helm install helm/chaos-operator --name=chaos-operator --namespace=chaos-testing

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	$(GO) fmt ./...

# Run go vet against code
vet:
	$(GO) vet ./...

tidy:
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff --quiet go.mod go.sum

image:
	docker build -t pingcap/chaos-operator images/chaos-operator
	docker build -t pingcap/chaos-fs images/chaosfs
	cp -R hack images/chaos-scripts
	docker build -t pingcap/chaos-scripts images/chaos-scripts
	rm -rf images/hack

docker-push:
	docker push "${DOCKER_REGISTRY}/pingcap/chaos-operator:latest"
	docker push "${DOCKER_REGISTRY}/pingcap/chaos-fs:latest"
	docker push "${DOCKER_REGISTRY}/pingcap/chaos-scripts:latest"

lint:
	@echo "linting"
	CGO_ENABLED=0 retool do revive -formatter friendly -config revive.toml $$($(PACKAGE_LIST))

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.1
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

yaml: manifests
	kustomize build config/default > manifests/crd.yaml


.PHONY: all build test install manifests fmt vet tidy image \
	docker-push lint generate controller-gen yaml \
	manager chaosfs chaosdaemon
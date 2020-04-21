# Set DEBUGGER=1 to build debug symbols
LDFLAGS = $(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh)

# SET DOCKER_REGISTRY to change the docker registry
DOCKER_REGISTRY_PREFIX := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)
DOCKER_BUILD_ARGS := --build-arg HTTP_PROXY=${HTTP_PROXY} --build-arg HTTPS_PROXY=${HTTPS_PROXY}

GOVER_MAJOR := $(shell go version | sed -E -e "s/.*go([0-9]+)[.]([0-9]+).*/\1/")
GOVER_MINOR := $(shell go version | sed -E -e "s/.*go([0-9]+)[.]([0-9]+).*/\2/")
GO111 := $(shell [ $(GOVER_MAJOR) -gt 1 ] || [ $(GOVER_MAJOR) -eq 1 ] && [ $(GOVER_MINOR) -ge 11 ]; echo $$?)

IMAGE_TAG := $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)

ROOT=$(shell pwd)
OUTPUT_BIN=$(ROOT)/output/bin
KUSTOMIZE_BIN=$(OUTPUT_BIN)/kustomize
KUBEBUILDER_BIN=$(OUTPUT_BIN)/kubebuilder
KUBECTL_BIN=$(OUTPUT_BIN)/kubectl
HELM_BIN=$(OUTPUT_BIN)/helm

ifeq ($(GO111), 1)
$(error Please upgrade your Go compiler to 1.11 or higher version)
endif

# Enable GO111MODULE=on explicitly, disable it with GO111MODULE=off when necessary.
export GO111MODULE := on
GOOS := $(if $(GOOS),$(GOOS),"")
GOARCH := $(if $(GOARCH),$(GOARCH),"")
GOENV  := GO15VENDOREXPERIMENT="1" CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
CGOENV  := GO15VENDOREXPERIMENT="1" CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go
GOTEST := TEST_USE_EXISTING_CLUSTER=false NO_PROXY="${NO_PROXY},testhost" go test
SHELL    := /usr/bin/env bash

PACKAGE_LIST := go list ./... | grep -vE "pkg/client" | grep -vE "zz_generated" | grep -vE "vendor"
PACKAGE_DIRECTORIES := $(PACKAGE_LIST) | sed 's|github.com/pingcap/chaos-mesh/||'
FILES := $$(find $$($(PACKAGE_DIRECTORIES)) -name "*.go")
FAIL_ON_STDOUT := awk '{ print } END { if (NR > 0) { exit 1 } }'

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

FAILPOINT_ENABLE  := $$(find $$PWD/ -type d | grep -vE "(\.git|bin)" | xargs $(GOBIN)/failpoint-ctl enable)
FAILPOINT_DISABLE := $$(find $$PWD/ -type d | grep -vE "(\.git|bin)" | xargs $(GOBIN)/failpoint-ctl disable)

all: yaml build image

build: dashboard-server-frontend

check: fmt vet lint generate yaml tidy gosec-scan

# Run tests
test: failpoint-enable generate manifests test-utils
	rm -rf cover.* cover
	mkdir -p cover
	$(GOTEST) ./api/... ./controllers/... ./pkg/... -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out
	@$(FAILPOINT_DISABLE)

test-utils: timer multithread_tracee

timer:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/test/timer ./test/cmd/timer/*.go

multithread_tracee: test/cmd/multithread_tracee/main.c
	cc test/cmd/multithread_tracee/main.c -lpthread -O2 -o ./bin/test/multithread_tracee

coverage:
ifeq ("$(CI)", "1")
	@bash <(curl -s https://codecov.io/bash) -f cover.out -t $(CODECOV_TOKEN)
else
	gocov convert cover.out > cover.json
	gocov-xml < cover.json > cover.xml
	gocov-html < cover.json > cover/index.html
endif

# Build chaos-daemon binary
chaosdaemon: generate
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -o bin/chaos-daemon ./cmd/chaos-daemon/main.go

# Build manager binary
manager: generate
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaos-controller-manager ./cmd/controller-manager/*.go

chaosfs: generate
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaosfs ./cmd/chaosfs/*.go

dashboard:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaos-dashboard ./cmd/chaos-dashboard/*.go

binary: chaosdaemon manager chaosfs dashboard

watchmaker:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -o bin/watchmaker ./cmd/watchmaker/...

dashboard-server-frontend:
	cd images/chaos-dashboard; yarn install; yarn build

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	$(GO) run ./cmd/controller-manager/main.go

# Install CRDs into a cluster
install: manifests
	$(KUBECTL_BIN) apply -f manifests/crd.yaml
	bash -c '[[ `$(HELM_BIN) version --client --short` == "Client: v2"* ]] && $(HELM_BIN) install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing || $(HELM_BIN) install chaos-mesh helm/chaos-mesh --namespace=chaos-testing;'

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(GOBIN)/controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt: groupimports
	$(CGOENV) go fmt ./...

gosec-scan: gosec
	$(GOENV) $(GOBIN)/gosec ./api/... ./controllers/... ./pkg/...

groupimports: goimports
	$(GOBIN)/goimports -w -l -local github.com/pingcap/chaos-mesh $$($(PACKAGE_DIRECTORIES))

failpoint-enable: failpoint-ctl
# Converting gofail failpoints...
	@$(FAILPOINT_ENABLE)

failpoint-disable: failpoint-ctl
# Restoring gofail failpoints...
	@$(FAILPOINT_DISABLE)

# Run go vet against code
vet:
	$(CGOENV) go vet ./...

tidy:
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff --quiet go.mod go.sum

image: image-chaos-daemon image-chaos-mesh image-chaos-fs image-chaos-scripts image-chaos-grafana image-chaos-dashboard

image-binary:
	docker build -t pingcap/binary ${DOCKER_BUILD_ARGS} .

image-chaos-daemon: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-daemon:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} images/chaos-daemon

image-chaos-mesh: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} images/chaos-mesh

image-chaos-fs: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-fs:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} images/chaosfs

image-chaos-scripts: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-scripts:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} images/chaos-scripts

image-chaos-grafana:
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-grafana:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} images/grafana

image-chaos-dashboard: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-dashboard:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} images/chaos-dashboard

image-chaos-kernel:
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-kernel ${DOCKER_BUILD_ARGS} --build-arg MAKE_JOBS=${MAKE_JOBS} --build-arg MIRROR=${UBUNTU_MIRROR} images/chaos-kernel

docker-push:
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-fs:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-daemon:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-scripts:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-grafana:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-dashboard:${IMAGE_TAG}"

docker-push-chaos-kernel:
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-kernel:${IMAGE_TAG}"

controller-gen:
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5
revive:
	$(GO) get github.com/mgechev/revive@v1.0.2-0.20200225072153-6219ca02fffb
failpoint-ctl:
	$(GO) get github.com/pingcap/failpoint/failpoint-ctl@v0.0.0-20200210140405-f8f9fb234798
goimports:
	$(GO) get golang.org/x/tools/cmd/goimports@v0.0.0-20200309202150-20ab64c0d93f
gosec:
	$(GO) get github.com/securego/gosec/cmd/gosec

lint: revive
	@echo "linting"
	$(GOBIN)/revive -formatter friendly -config revive.toml $$($(PACKAGE_LIST))

# Generate code
generate: controller-gen
	$(GOBIN)/controller-gen object:headerFile=./hack/boilerplate.go.txt paths="./..."

yaml: manifests ensure-kustomize
	$(KUSTOMIZE_BIN) build config/default > manifests/crd.yaml

e2e-build:
	$(GO) build -trimpath  -o test/image/e2e/bin/ginkgo github.com/onsi/ginkgo/ginkgo
	$(GO) test -c  -o ./test/image/e2e/bin/e2e.test ./test/e2e

ifeq ($(NO_BUILD),y)
e2e-docker:
	@echo "NO_BUILD=y, skip build for $@"
else
e2e-docker: e2e-build
endif
	[ -d test/image/e2e/chaos-mesh ] && rm -r test/image/e2e/chaos-mesh || true
	cp -r helm/chaos-mesh test/image/e2e
	cp -r manifests test/image/e2e
	docker build -t "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh-e2e:${IMAGE_TAG}" test/image/e2e

ensure-kind:
	@echo "ensuring kind"
	$(shell ./hack/tools.sh kind)

ensure-kubebuilder:
	@echo "ensuring kubebuilder"
	$(shell ./hack/tools.sh kubebuilder)

ensure-kustomize:
	@echo "ensuring kustomize"
	$(shell ./hack/tools.sh kustomize)

ensure-kubectl:
	@echo "ensuring kubectl"
	$(shell ./hack/tools.sh kubectl)

ensure-all:
	@echo "ensuring all"
	$(shell ./hack/tools.sh all)

.PHONY: all build test install manifests groupimports fmt vet tidy image \
	docker-push lint generate controller-gen yaml \
	manager chaosfs chaosdaemon ensure-all \
	dashboard dashboard-server-frontend \
	gosec-scan

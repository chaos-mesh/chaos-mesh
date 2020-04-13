# Set DEBUGGER=1 to build debug symbols
LDFLAGS = $(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh)

# SET DOCKER_REGISTRY to change the docker registry
DOCKER_REGISTRY_PREFIX := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)
DOCKER_BUILD_ARGS := --build-arg HTTP_PROXY=${HTTP_PROXY} --build-arg HTTPS_PROXY=${HTTPS_PROXY}

GOVER_MAJOR := $(shell go version | sed -E -e "s/.*go([0-9]+)[.]([0-9]+).*/\1/")
GOVER_MINOR := $(shell go version | sed -E -e "s/.*go([0-9]+)[.]([0-9]+).*/\2/")
GO111 := $(shell [ $(GOVER_MAJOR) -gt 1 ] || [ $(GOVER_MAJOR) -eq 1 ] && [ $(GOVER_MINOR) -ge 11 ]; echo $$?)

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

all: yaml image

# Run tests
test: failpoint-enable generate fmt vet lint manifests test-utils
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
ifeq ("$(JenkinsCI)", "1")
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

chaos-server: generate
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaos-server ./cmd/chaos-server/*.go

binary: chaosdaemon manager chaosfs chaos-server

watchmaker:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -o bin/watchmaker ./cmd/watchmaker/...

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	$(GO) run ./cmd/controller-manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f manifests/crd.yaml
	bash -c '[[ `helm version --client --short` == "Client: v2"* ]] && helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing || helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing;'

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(GOBIN)/controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt: groupimports
	$(CGOENV) go fmt ./...

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

image: image-chaos-daemon image-chaos-mesh image-chaos-fs image-chaos-scripts image-chaos-grafana image-chaos-kernel

image-binary:
	docker build -t pingcap/binary ${DOCKER_BUILD_ARGS} .

image-chaos-daemon: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-daemon ${DOCKER_BUILD_ARGS} images/chaos-daemon

image-chaos-mesh: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh ${DOCKER_BUILD_ARGS} images/chaos-mesh

image-chaos-fs: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-fs ${DOCKER_BUILD_ARGS} images/chaosfs

image-chaos-scripts: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-scripts ${DOCKER_BUILD_ARGS} images/chaos-scripts

image-chaos-server: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-server ${DOCKER_BUILD_ARGS} images/chaos-server

image-chaos-grafana:
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-grafana ${DOCKER_BUILD_ARGS} images/grafana

image-chaos-kernel:
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-kernel ${DOCKER_BUILD_ARGS} --build-arg MAKE_JOBS=${MAKE_JOBS} --build-arg MIRROR=${UBUNTU_MIRROR} images/chaos-kernel

docker-push:
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh:latest"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-fs:latest"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-daemon:latest"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-scripts:latest"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-grafana:latest"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-kernel:latest"

controller-gen:
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5
revive:
	$(GO) get github.com/mgechev/revive@v1.0.2-0.20200225072153-6219ca02fffb
failpoint-ctl:
	$(GO) get github.com/pingcap/failpoint/failpoint-ctl@v0.0.0-20200210140405-f8f9fb234798
goimports:
	$(GO) get golang.org/x/tools/cmd/goimports@v0.0.0-20200309202150-20ab64c0d93f

lint: revive
	@echo "linting"
	$(GOBIN)/revive -formatter friendly -config revive.toml $$($(PACKAGE_LIST))

# Generate code
generate: controller-gen
	$(GOBIN)/controller-gen object:headerFile=./hack/boilerplate.go.txt paths="./..."

yaml: manifests
	kustomize build config/default > manifests/crd.yaml

e2e-build:
	$(GO) build -trimpath  -o test/image/e2e/bin/ginkgo github.com/onsi/ginkgo/ginkgo
	$(GO) test -c  -o ./test/image/e2e/bin/e2e.test ./test/e2e

e2e-docker: e2e-build
	[ -d test/image/e2e/chaos-mesh ] && rm -r test/image/e2e/chaos-mesh || true
	cp -r helm/chaos-mesh test/image/e2e
	cp -r manifests test/image/e2e
	docker build -t "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh-e2e:latest" test/image/e2e

check: fmt vet lint

install-kind:
ifeq (,$(shell which kind))
	@echo "installing kind"
	GO111MODULE="on" go get sigs.k8s.io/kind@v0.7.0
else
	@echo "kind has been installed"
endif

install-kubebuilder:
ifeq (,$(shell which kubebuilder))
	@echo "installing kubebuilder"
	# download kubebuilder and extract it to tmp
	curl -sL https://go.kubebuilder.io/dl/2.2.0/$(shell go env GOOS)/$(shell go env GOARCH) | tar -zx -C /tmp/
	# move to a long-term location and put it on your path
	# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
	sudo mv /tmp/kubebuilder_2.2.0_$(shell go env GOOS)_$(shell go env GOARCH) /usr/local/kubebuilder
	export PATH="${PATH}:/usr/local/kubebuilder/bin"
else
	@echo "kubebuilder has been installed"
endif

install-kustomize:
ifeq (,$(shell which kustomize))
	@echo "installing kustomize"
	# download kustomize
	curl -o /usr/local/kubebuilder/bin/kustomize -sL "https://go.kubebuilder.io/kustomize/$(shell go env GOOS)/$(shell go env GOARCH)"
	# set permission
	sudo chmod a+x /usr/local/kubebuilder/bin/kustomize
	$(shell which kustomize)
else
	@echo "kustomize has been installed"
endif

.PHONY: all build test install manifests groupimports fmt vet tidy image \
	docker-push lint generate controller-gen yaml \
	manager chaosfs chaosdaemon chaos-server \
	install-kind install-kubebuilder install-kustomize

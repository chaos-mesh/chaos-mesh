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
GOTEST := TEST_USE_EXISTING_CLUSTER=false go test

PACKAGE_LIST := go list ./... | grep -vE "pkg/client" | grep -vE "zz_generated"
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

all: yaml build image

build: chaosdaemon manager chaosfs dashboard

# Run tests
test: generate fmt vet lint manifests
	rm -rf cover.* cover
	mkdir -p cover
	$(GOTEST) ./api/... ./controllers/... ./pkg/... -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out
	gocov convert cover.out > cover.json
	gocov-xml < cover.json > cover.xml
	gocov-html < cover.json > cover/index.html
	rm -rf cover.out cover.out.tmp cover.json

# Build chaos-daemon binary
chaosdaemon: generate fmt vet
	$(GO) build -ldflags '$(LDFLAGS)' -o images/chaos-daemon/bin/chaos-daemon ./cmd/chaos-daemon/main.go

# Build manager binary
manager: generate fmt vet
	$(GO) build -ldflags '$(LDFLAGS)' -o images/chaos-mesh/bin/chaos-controller-manager ./cmd/controller-manager/*.go

chaosfs: generate fmt vet
	$(GO) build -ldflags '$(LDFLAGS)' -o images/chaosfs/bin/chaosfs ./cmd/chaosfs/*.go

dashboard: fmt vet
	$(GO) build -ldflags '$(LDFLAGS)' -o images/chaos-dashboard/bin/chaos-dashboard ./cmd/chaos-dashboard/*.go

dashboard-server-frontend:
	cd images/chaos-dashboard; yarn build

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	$(GO) run ./cmd/controller-manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f manifests/crd.yaml
	helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing

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

image: dashboard-server-frontend
	docker build -t ${DOCKER_REGISTRY}/pingcap/chaos-daemon images/chaos-daemon
	docker build -t ${DOCKER_REGISTRY}/pingcap/chaos-mesh images/chaos-mesh
	docker build -t ${DOCKER_REGISTRY}/pingcap/chaos-fs images/chaosfs
	cp -R hack images/chaos-scripts
	docker build -t ${DOCKER_REGISTRY}/pingcap/chaos-scripts images/chaos-scripts
	rm -rf images/chaos-scripts/hack
	docker build -t ${DOCKER_REGISTRY}/pingcap/chaos-grafana images/grafana
	docker build -t ${DOCKER_REGISTRY}/pingcap/chaos-dashboard images/chaos-dashboard

docker-push:
	docker push "${DOCKER_REGISTRY}/pingcap/chaos-mesh:latest"
	docker push "${DOCKER_REGISTRY}/pingcap/chaos-fs:latest"
	docker push "${DOCKER_REGISTRY}/pingcap/chaos-daemon:latest"
	docker push "${DOCKER_REGISTRY}/pingcap/chaos-scripts:latest"

bin/revive:
	GO111MODULE="on" go build -o bin/revive github.com/mgechev/revive

lint: bin/revive
	@echo "linting"
	bin/revive -formatter friendly -config revive.toml $$($(PACKAGE_LIST))

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

yaml: manifests
	kustomize build config/default > manifests/crd.yaml

install-kind:
ifeq (,$(shell which kind))
	@echo "installing kind"
	GO111MODULE="on" go get sigs.k8s.io/kind@v0.4.0
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
	mv /tmp/kubebuilder_2.2.0_$(shell go env GOOS)_$(shell go env GOARCH) /usr/local/kubebuilder
	export PATH=${PATH}:/usr/local/kubebuilder/bin
else
	@echo "kubebuilder has been installed"
endif

install-kustomize:
ifeq (,$(shell which kustomize))
	@echo "installing kustomize"
	# download kustomize
	curl -o /usr/local/kubebuilder/bin/kustomize -sL "https://go.kubebuilder.io/kustomize/$(shell go env GOOS)/$(shell go env GOARCH)"
	# set permission
	chmod a+x /usr/local/kubebuilder/bin/kustomize
	$(shell which kustomize)
else
	@echo "kustomize has been installed"
endif

install-test-dependency:
	go get -u github.com/jstemmer/go-junit-report \
	&& go get github.com/axw/gocov/gocov \
	&& go get github.com/AlekSi/gocov-xml \
	&& go get github.com/onsi/ginkgo/ginkgo \
	&& go get golang.org/x/tools/cmd/cover \
	&& go get -u github.com/matm/gocov-html


.PHONY: all build test install manifests fmt vet tidy image \
	docker-push lint generate controller-gen yaml \
	manager chaosfs chaosdaemon install-kind install-kubebuilder \
	install-kustomize install-test-dependency dashboard dashboard-server-frontend
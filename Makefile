# Set DEBUGGER=1 to build debug symbols
LDFLAGS = $(if $(IMG_LDFLAGS),$(IMG_LDFLAGS),$(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh))
DOCKER_REGISTRY ?= "localhost:5000"

# SET DOCKER_REGISTRY to change the docker registry
DOCKER_REGISTRY_PREFIX := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)
DOCKER_BUILD_ARGS := --build-arg HTTP_PROXY=${HTTP_PROXY} --build-arg HTTPS_PROXY=${HTTPS_PROXY} --build-arg UI=${UI} --build-arg SWAGGER=${SWAGGER} --build-arg LDFLAGS="${LDFLAGS}" --build-arg CRATES_MIRROR="${CRATES_MIRROR}"

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
CGO    := $(CGOENV) go
GOTEST := TEST_USE_EXISTING_CLUSTER=false NO_PROXY="${NO_PROXY},testhost" go test
SHELL    := /usr/bin/env bash

PACKAGE_LIST := go list ./... | grep -vE "chaos-mesh/test|pkg/ptrace|zz_generated|vendor"
PACKAGE_DIRECTORIES := $(PACKAGE_LIST) | sed 's|github.com/chaos-mesh/chaos-mesh/||'

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

FAILPOINT_ENABLE  := $$(find $$PWD/ -type d | grep -vE "(\.git|bin)" | xargs $(GOBIN)/failpoint-ctl enable)
FAILPOINT_DISABLE := $$(find $$PWD/ -type d | grep -vE "(\.git|bin)" | xargs $(GOBIN)/failpoint-ctl disable)

BUILD_TAGS ?=

ifeq ($(SWAGGER),1)
	BUILD_TAGS += swagger_server
endif

ifeq ($(UI),1)
	BUILD_TAGS += ui_server
endif

all: yaml image

build: binary

check: fmt vet boilerplate lint generate yaml tidy

# Run tests
test: failpoint-enable generate manifests test-utils
	rm -rf cover.* cover
	$(GOTEST) $$($(PACKAGE_LIST)) -coverprofile cover.out.tmp
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
	mkdir -p cover
	gocov convert cover.out > cover.json
	gocov-xml < cover.json > cover.xml
	gocov-html < cover.json > cover/index.html
endif

# Build chaos-daemon binary
chaosdaemon:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -o bin/chaos-daemon ./cmd/chaos-daemon/main.go

bin/pause: ./hack/pause.c
	cc ./hack/pause.c -o bin/pause

# Build manager binary
manager:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaos-controller-manager ./cmd/controller-manager/*.go

chaosfs:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaosfs ./cmd/chaosfs/*.go

chaos-dashboard:
ifeq ($(SWAGGER),1)
	make swagger_spec
endif
ifeq ($(UI),1)
	make ui
	hack/embed_ui_assets.sh
endif
	$(CGO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o bin/chaos-dashboard cmd/chaos-dashboard/*.go

swagger_spec:
	hack/generate_swagger_spec.sh

yarn_dependencies:
	cd ui &&\
	yarn install --frozen-lockfile

ui: yarn_dependencies
	cd ui &&\
	yarn build

binary: chaosdaemon manager chaosfs chaos-dashboard bin/pause

watchmaker:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -o bin/watchmaker ./cmd/watchmaker/...

# Build chaosctl
chaosctl:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaosctl ./cmd/chaosctl/*.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	$(GO) run ./cmd/controller-manager/main.go

# Install CRDs into a cluster
install: manifests
	$(KUBECTL_BIN) apply -f manifests/crd.yaml
	bash -c '[[ `$(HELM_BIN) version --client --short` == "Client: v2"* ]] && $(HELM_BIN) install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing || $(HELM_BIN) install chaos-mesh helm/chaos-mesh --namespace=chaos-testing;'

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(GOBIN)/controller-gen
	$< $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt: groupimports
	$(CGOENV) go fmt ./...

gosec-scan: $(GOBIN)/gosec
	$(GOENV) $< ./api/... ./controllers/... ./pkg/... || echo "*** sec-scan failed: known-issues ***"

groupimports: $(GOBIN)/goimports
	$< -w -l -local github.com/chaos-mesh/chaos-mesh .

failpoint-enable: $(GOBIN)/failpoint-ctl
# Converting gofail failpoints...
	@$(FAILPOINT_ENABLE)

failpoint-disable: $(GOBIN)/failpoint-ctl
# Restoring gofail failpoints...
	@$(FAILPOINT_DISABLE)

# Run go vet against code
vet:
	$(CGOENV) go vet ./...

tidy: clean
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff -U --exit-code go.mod go.sum

install.sh:
	./hack/update_install_script.sh

check-install-script: install.sh
	git diff -U --exit-code install.sh

clean:
	rm -rf docs/docs.go

boilerplate:
	./hack/verify-boilerplate.sh

image: image-chaos-daemon image-chaos-mesh image-chaos-dashboard

define COPY_TEMPLATE =
images/$(1)/bin/$(2): image-chaos-binary
	docker run --rm --volume $(shell pwd)/images/$(1)/bin:/mnt ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-binary cp /bin/$(2) /mnt/$(2)

image-$(1)-dependencies := $(image-$(1)-dependencies) images/$(1)/bin/$(2)
endef

$(eval $(call COPY_TEMPLATE,chaos-daemon,chaos-daemon))
$(eval $(call COPY_TEMPLATE,chaos-daemon,toda))
$(eval $(call COPY_TEMPLATE,chaos-daemon,nsexec))
$(eval $(call COPY_TEMPLATE,chaos-daemon,libnsenter.so))
$(eval $(call COPY_TEMPLATE,chaos-daemon,pause))
$(eval $(call COPY_TEMPLATE,chaos-dashboard,chaos-dashboard))
$(eval $(call COPY_TEMPLATE,chaos-mesh,chaos-controller-manager))

define IMAGE_TEMPLATE =
image-$(1):$(image-$(1)-dependencies)
ifeq ($(DOCKER_CACHE),1)
	DOCKER_BUILDKIT=1 DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build --load --cache-to type=local,dest=$(CACHE_DIR)/$(1) --cache-from type=local,src=$(CACHE_DIR)/$(1) -t ${DOCKER_REGISTRY_PREFIX}pingcap/$(1):${IMAGE_TAG} ${DOCKER_BUILD_ARGS} $(2)
else
	DOCKER_BUILDKIT=1 docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/$(1):${IMAGE_TAG} ${DOCKER_BUILD_ARGS} $(2)
endif
endef

$(eval $(call IMAGE_TEMPLATE,chaos-daemon,images/chaos-daemon))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh,images/chaos-mesh))
$(eval $(call IMAGE_TEMPLATE,chaos-dashboard,images/chaos-dashboard))
$(eval $(call IMAGE_TEMPLATE,chaos-binary,.))
$(eval $(call IMAGE_TEMPLATE,e2e-helper,test/cmd/e2e_helper))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh-protoc,./hack/protoc))

image-chaos-kernel:
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-kernel:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} --build-arg MAKE_JOBS=${MAKE_JOBS} --build-arg MIRROR=${UBUNTU_MIRROR} images/chaos-kernel

docker-push:
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-dashboard:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-daemon:${IMAGE_TAG}"

docker-push-chaos-kernel:
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-kernel:${IMAGE_TAG}"

$(GOBIN)/controller-gen:
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5
$(GOBIN)/revive:
	$(GO) get github.com/mgechev/revive@v1.0.2-0.20200225072153-6219ca02fffb
$(GOBIN)/failpoint-ctl:
	$(GO) get github.com/pingcap/failpoint/failpoint-ctl@v0.0.0-20200210140405-f8f9fb234798
$(GOBIN)/goimports:
	$(GO) get golang.org/x/tools/cmd/goimports@v0.0.0-20200309202150-20ab64c0d93f
$(GOBIN)/gosec:
	$(GO) get github.com/securego/gosec/cmd/gosec@v0.0.0-20200401082031-e946c8c39989

lint: $(GOBIN)/revive
	@echo "linting"
	$< -formatter friendly -config revive.toml $$($(PACKAGE_LIST))

bin/chaos-builder:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -o bin/chaos-builder ./cmd/chaos-builder/...

chaos-build: bin/chaos-builder
	bin/chaos-builder

# Generate code
generate: $(GOBIN)/controller-gen chaos-build
	$< object:headerFile=./hack/boilerplate/boilerplate.generatego.txt paths="./..."

yaml: manifests ensure-kustomize
	$(KUSTOMIZE_BIN) build config/default > manifests/crd.yaml

# Generate Go files from Chaos Mesh proto files.
ifeq ($(IN_DOCKER),1)
proto:
	for dir in pkg/chaosdaemon pkg/chaosfs; do\
		protoc -I $$dir/pb $$dir/pb/*.proto --go_out=plugins=grpc:$$dir/pb --go_out=./$$dir/pb ;\
	done
else
proto: image-chaos-mesh-protoc
	docker run --rm --workdir /mnt/ --volume $(shell pwd):/mnt \
		--user $(shell id -u):$(shell id -g) --env IN_DOCKER=1 ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh-protoc \
		/usr/bin/make proto

	make fmt
endif

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

tools := kubectl helm kind kubebuilder kustomize kubetest2
define DOWNLOAD_TOOL =
ensure-$(1):
	@echo "ensuring $(1)"
	ROOT=$(ROOT) && source ./hack/lib.sh && hack::ensure_$(1)

all-tool-dependencies := $(all-tool-dependencies) ensure-$(1)
endef

$(foreach tool, $(tools), $(eval $(call DOWNLOAD_TOOL,$(tool))))

ensure-all: $(all-tool-dependencies)

install-local-coverage-tools:
	go get github.com/axw/gocov/gocov \
	&& go get github.com/AlekSi/gocov-xml \
	&& go get -u github.com/matm/gocov-html

.PHONY: all build test install manifests groupimports fmt vet tidy image \
	binary docker-push lint generate yaml \
	$(all-tool-dependencies) install.sh \
	manager chaosfs chaosdaemon chaos-dashboard \
	dashboard dashboard-server-frontend gosec-scan \
	proto bin/chaos-builder

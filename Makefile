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
SHELL    := bash

PACKAGE_LIST := echo $$(go list ./... | grep -vE "chaos-mesh/test|pkg/ptrace|zz_generated|vendor") github.com/chaos-mesh/chaos-mesh/api/v1alpha1

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

GO_BUILD_CACHE ?= $(HOME)/.cache/chaos-mesh

go_build_cache_directory:
	mkdir -p $(GO_BUILD_CACHE)/chaos-mesh-gobuild
	mkdir -p $(GO_BUILD_CACHE)/chaos-mesh-gopath

BUILD_TAGS ?=

ifeq ($(SWAGGER),1)
	BUILD_TAGS += swagger_server
endif

ifeq ($(UI),1)
	BUILD_TAGS += ui_server
endif

CLEAN_TARGETS :=

all: manifests/crd.yaml image

check: fmt vet boilerplate lint generate manifests/crd.yaml tidy

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

swagger_spec:
ifeq (${SWAGGER},1)
	hack/generate_swagger_spec.sh
endif

yarn_dependencies:
ifeq (${UI},1)
	cd ui &&\
	yarn install --frozen-lockfile
endif

ui: yarn_dependencies
ifeq (${UI},1)
	cd ui &&\
	yarn build
	hack/embed_ui_assets.sh
endif

watchmaker:
	$(CGO) build -ldflags '$(LDFLAGS)' -o bin/watchmaker ./cmd/watchmaker/...

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
config: $(GOBIN)/controller-gen
	cd ./api/v1alpha1 ;\
		$< $(CRD_OPTIONS) rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../../config/crd/bases ;\
		$< $(CRD_OPTIONS) rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../../helm/chaos-mesh/crds ;

# Run go fmt against code
fmt: groupimports
	$(CGO) fmt ./...

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
	rm -rf docs/docs.go $(CLEAN_TARGETS)

boilerplate:
	./hack/verify-boilerplate.sh

image: image-chaos-daemon image-chaos-mesh image-chaos-dashboard

GO_TARGET_PHONY :=

BINARIES :=

define COMPILE_GO_TEMPLATE
ifeq ($(IN_DOCKER),1)

$(1): $(4)
ifeq ($(3),1)
	$(CGO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o $(1) $(2)
else
	$(GO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o $(1) $(2)
endif

endif
GO_TARGET_PHONY += $(1)
endef

BUILD_INDOCKER_ARG := --env IN_DOCKER=1
ifeq ($(DOCKER_HOST),)
    BUILD_INDOCKER_ARG += --volume $(ROOT):/mnt --user $(shell id -u):$(shell id -g)
else
	CLEAN_TARGETS += Dockerfile
endif

ifneq ($(GO_BUILD_CACHE),)
	BUILD_INDOCKER_ARG += --volume $(GO_BUILD_CACHE)/chaos-mesh-gopath:/tmp/go
	BUILD_INDOCKER_ARG += --volume $(GO_BUILD_CACHE)/chaos-mesh-gobuild:/tmp/go-build
endif

define BUILD_IN_DOCKER_TEMPLATE
CLEAN_TARGETS += $(2)
ifneq ($(IN_DOCKER),1)

$(2): image-build-env go_build_cache_directory
	[[ "$(DOCKER_HOST)" == "" ]] || (printf "\
	FROM ${DOCKER_REGISTRY_PREFIX}pingcap/build-env:${IMAGE_TAG} \n\
	RUN rm -rf /mnt \n\
	COPY ./ /mnt \n"\
	> Dockerfile; docker build . -t ${DOCKER_REGISTRY_PREFIX}pingcap/build-env:${IMAGE_TAG})

	DOCKER_ID=$$$$(docker run -d \
		$(BUILD_INDOCKER_ARG) \
		${DOCKER_REGISTRY_PREFIX}pingcap/build-env:${IMAGE_TAG} \
		sleep infinity); \
	docker exec --workdir /mnt/ \
		--env IMG_LDFLAGS="${LDFLAGS}" \
		--env UI=${UI} --env SWAGGER=${SWAGGER} \
		$$$$DOCKER_ID /usr/bin/make $(2); \
	[[ "$(DOCKER_HOST)" == "" ]] || docker cp $$$$DOCKER_ID:/mnt/$(2) $(2); \
	docker rm -f $$$$DOCKER_ID
endif

image-$(1)-dependencies := $(image-$(1)-dependencies) $(2)
BINARIES := $(BINARIES) $(2)
endef

ifeq ($(IN_DOCKER),1)
images/chaos-daemon/bin/pause: hack/pause.c
	cc ./hack/pause.c -o images/chaos-daemon/bin/pause
endif
$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-daemon,images/chaos-daemon/bin/pause))

$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-daemon,images/chaos-daemon/bin/chaos-daemon))
$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-daemon/bin/chaos-daemon,./cmd/chaos-daemon/main.go,1))

$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-dashboard,images/chaos-dashboard/bin/chaos-dashboard))
$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-dashboard/bin/chaos-dashboard,./cmd/chaos-dashboard/main.go,1,ui swagger_spec))

$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-mesh,images/chaos-mesh/bin/chaos-controller-manager))
$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-mesh/bin/chaos-controller-manager,./cmd/chaos-controller-manager/main.go,0))

$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-mesh-e2e,test/image/e2e/bin/ginkgo))
$(eval $(call COMPILE_GO_TEMPLATE,test/image/e2e/bin/ginkgo,github.com/onsi/ginkgo/ginkgo,0))

$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-mesh-e2e,test/image/e2e/bin/e2e.test))
ifeq ($(IN_DOCKER),1)
test/image/e2e/bin/e2e.test:
	$(GO) test -c  -o ./test/image/e2e/bin/e2e.test ./test/e2e
$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-mesh-e2e,test/image/e2e/bin/e2e.test))

GO_TARGET_PHONY += test/image/e2e/bin/e2e.test
endif

image-chaos-mesh-e2e-dependencies += test/image/e2e/manifests test/image/e2e/chaos-mesh

e2e-build: test/image/e2e/bin/ginkgo test/image/e2e/bin/e2e.test

test/image/e2e/manifests: manifests
	cp -r manifests test/image/e2e

test/image/e2e/chaos-mesh: helm/chaos-mesh
	cp -r helm/chaos-mesh test/image/e2e

define IMAGE_TEMPLATE
CLEAN_TARGETS += $(2)/.dockerbuilt

image-$(1): $(2)/.dockerbuilt

$(2)/.dockerbuilt:$(image-$(1)-dependencies) $(2)/Dockerfile
ifeq ($(DOCKER_CACHE),1)

ifneq ($(DISABLE_CACHE_FROM),1)
	DOCKER_BUILDKIT=1 DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build --load --cache-to type=local,dest=$(DOCKER_CACHE_DIR)/image-$(1) --cache-from type=local,src=$(DOCKER_CACHE_DIR)/image-$(1) -t ${DOCKER_REGISTRY_PREFIX}pingcap/$(1):${IMAGE_TAG} ${DOCKER_BUILD_ARGS} $(2)
else
	DOCKER_BUILDKIT=1 DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build --load --cache-to type=local,dest=$(DOCKER_CACHE_DIR)/image-$(1) -t ${DOCKER_REGISTRY_PREFIX}pingcap/$(1):${IMAGE_TAG} ${DOCKER_BUILD_ARGS} $(2)
endif

else
	DOCKER_BUILDKIT=1 docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/$(1):${IMAGE_TAG} ${DOCKER_BUILD_ARGS} $(2)
endif
	touch $(2)/.dockerbuilt
endef

$(eval $(call IMAGE_TEMPLATE,chaos-daemon,images/chaos-daemon))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh,images/chaos-mesh))
$(eval $(call IMAGE_TEMPLATE,chaos-dashboard,images/chaos-dashboard))
$(eval $(call IMAGE_TEMPLATE,build-env,images/build-env))
$(eval $(call IMAGE_TEMPLATE,e2e-helper,test/cmd/e2e_helper))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh-protoc,hack/protoc))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh-e2e,test/image/e2e))
$(eval $(call IMAGE_TEMPLATE,chaos-kernel,images/chaos-kernel))
$(eval $(call IMAGE_TEMPLATE,chaos-jvm,images/chaos-jvm))

binary: $(BINARIES)

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
	cd ./api/v1alpha1 ;\
		$< object:headerFile=../../hack/boilerplate/boilerplate.generatego.txt paths="./..." ;

manifests/crd.yaml: config ensure-kustomize
	$(KUSTOMIZE_BIN) build config/default > manifests/crd.yaml

yaml: manifests/crd.yaml

# Generate Go files from Chaos Mesh proto files.
ifeq ($(IN_DOCKER),1)
proto:
	for dir in pkg/chaosdaemon ; do\
		protoc -I $$dir/pb $$dir/pb/*.proto --go_out=plugins=grpc:$$dir/pb --go_out=./$$dir/pb ;\
	done
else
proto: image-chaos-mesh-protoc
	docker run --rm --workdir /mnt/ --volume $(shell pwd):/mnt \
		--user $(shell id -u):$(shell id -g) --env IN_DOCKER=1 ${DOCKER_REGISTRY_PREFIX}pingcap/chaos-mesh-protoc \
		/usr/bin/make proto

	make fmt
endif

tools := kubectl helm kind kubebuilder kustomize kubetest2
define DOWNLOAD_TOOL
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

.PHONY: all clean test install manifests groupimports fmt vet tidy image \
	docker-push lint generate config \
	$(all-tool-dependencies) install.sh $(GO_TARGET_PHONY) \
	manager chaosfs chaosdaemon chaos-dashboard \
	dashboard dashboard-server-frontend gosec-scan \
	proto bin/chaos-builder go_build_cache_directory

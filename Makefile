# Set DEBUGGER=1 to build debug symbols
export LDFLAGS := $(if $(LDFLAGS),$(LDFLAGS),$(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh))
export DOCKER_REGISTRY ?= "localhost:5000"

# SET DOCKER_REGISTRY to change the docker registry
DOCKER_REGISTRY_PREFIX := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)

export IMAGE_TAG := $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)
export IMAGE_PROJECT := $(if $(IMAGE_PROJECT),$(IMAGE_PROJECT),pingcap)
export IMAGE_CHAOS_MESH_PROJECT := chaos-mesh
export IMAGE_CHAOS_DAEMON_PROJECT := chaos-mesh
export IMAGE_CHAOS_DASHBOARD_PROJECT := chaos-mesh

ROOT=$(shell pwd)
HELM_BIN=$(ROOT)/output/bin/helm

# Every branch should have its own image tag for build-env and dev-env
export IMAGE_BUILD_ENV_PROJECT ?= chaos-mesh
export IMAGE_BUILD_ENV_REGISTRY ?= ghcr.io
export IMAGE_BUILD_ENV_BUILD ?= 0
export IMAGE_BUILD_ENV_TAG ?= latest
export IMAGE_DEV_ENV_PROJECT ?= chaos-mesh
export IMAGE_DEV_ENV_REGISTRY ?= ghcr.io
export IMAGE_DEV_ENV_BUILD ?= 0
export IMAGE_DEV_ENV_TAG ?= latest

export GOPROXY  := $(if $(GOPROXY),$(GOPROXY),https://proxy.golang.org,direct)
GOENV  	:= CGO_ENABLED=0
CGOENV 	:= CGO_ENABLED=1
GO     	:= $(GOENV) go
CGO    	:= $(CGOENV) go
GOTEST 	:= USE_EXISTING_CLUSTER=false NO_PROXY="${NO_PROXY},testhost" go test
SHELL  	:= bash

PACKAGE_LIST := echo $$(go list ./... | grep -vE "chaos-mesh/test|pkg/ptrace|zz_generated|vendor") github.com/chaos-mesh/chaos-mesh/api/v1alpha1

# no version conversion
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false,crdVersions=v1"

export GO_BUILD_CACHE ?= $(ROOT)/.cache/chaos-mesh

BUILD_TAGS ?=

ifeq ($(UI),1)
	BUILD_TAGS += ui_server
endif

CLEAN_TARGETS :=

all: yaml image

test-utils: timer multithread_tracee pkg/time/fakeclock/fake_clock_gettime.o

timer:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/test/timer ./test/cmd/timer/*.go

multithread_tracee: test/cmd/multithread_tracee/main.c
	cc test/cmd/multithread_tracee/main.c -lpthread -O2 -o ./bin/test/multithread_tracee

yarn_dependencies:
ifeq (${UI},1)
	cd ui &&\
	yarn install --frozen-lockfile --network-timeout 500000
endif

ui: yarn_dependencies
ifeq (${UI},1)
	cd ui &&\
	yarn build
	hack/embed_ui_assets.sh
endif

watchmaker: pkg/time/fakeclock/fake_clock_gettime.o
	$(CGO) build -ldflags '$(LDFLAGS)' -o bin/watchmaker ./cmd/watchmaker/...

# Build chaosctl
chaosctl:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaosctl ./cmd/chaosctl/*.go

# Build schedule-migration
schedule-migration:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/schedule-migration ./tools/schedule-migration/*.go

schedule-migration.tar.gz: schedule-migration
	cp ./bin/schedule-migration ./schedule-migration
	cp ./tools/schedule-migration/migrate.sh ./migrate.sh
	tar -czvf schedule-migration.tar.gz schedule-migration migrate.sh
	rm ./migrate.sh
	rm ./schedule-migration

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	$(GO) run ./cmd/controller-manager/main.go

NAMESPACE ?= chaos-testing
# Install CRDs into a cluster
install: manifests
	$(HELM_BIN) upgrade --install chaos-mesh helm/chaos-mesh --namespace=${NAMESPACE} --set registry=${DOCKER_REGISTRY} --set dnsServer.create=true --set dashboard.create=true;

clean:
	rm -rf $(CLEAN_TARGETS)

SKYWALKING_EYES_HEADER = $(RUN_IN_DEV) /bin/license-eye header -c ./.github/.licenserc.yaml
boilerplate: image-dev-env
	$(SKYWALKING_EYES_HEADER) check

boilerplate-fix: image-dev-env
	$(SKYWALKING_EYES_HEADER) fix

image: image-chaos-daemon image-chaos-mesh image-chaos-dashboard $(if $(DEBUGGER), image-chaos-dlv)

e2e-image: image-e2e-helper

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

ifneq ($(GO_BUILD_CACHE),)
	BUILD_INDOCKER_ARG += --volume $(GO_BUILD_CACHE)/chaos-mesh-gopath:/tmp/go
	BUILD_INDOCKER_ARG += --volume $(GO_BUILD_CACHE)/chaos-mesh-gobuild:/tmp/go-build
endif

define BUILD_IN_DOCKER_TEMPLATE
CLEAN_TARGETS += $(2)

ifneq ($(IN_DOCKER),1)
$(2): image-build-env
	$(ROOT)/build/run_in_docker.py build-env make $(2)
endif

image-$(1)-dependencies := $(image-$(1)-dependencies) $(2)
BINARIES := $(BINARIES) $(2)
endef

enter-buildenv: image-build-env
	$(ROOT)/build/run_in_docker.py --interactive --no-check build-env bash

enter-devenv: image-dev-env
	$(ROOT)/build/run_in_docker.py --interactive --no-check dev-env bash

ifeq ($(IN_DOCKER),1)
images/chaos-daemon/bin/pause: hack/pause.c
	cc ./hack/pause.c -o images/chaos-daemon/bin/pause

pkg/time/fakeclock/fake_clock_gettime.o: pkg/time/fakeclock/fake_clock_gettime.c
	cc -c ./pkg/time/fakeclock/fake_clock_gettime.c -fPIE -O2 -o pkg/time/fakeclock/fake_clock_gettime.o
endif
$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-daemon,images/chaos-daemon/bin/pause))

$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-daemon,pkg/time/fakeclock/fake_clock_gettime.o))
$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-daemon,images/chaos-daemon/bin/chaos-daemon))
$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-daemon/bin/chaos-daemon,./cmd/chaos-daemon/main.go,1,pkg/time/fakeclock/fake_clock_gettime.o))

$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-dashboard,images/chaos-dashboard/bin/chaos-dashboard))
$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-dashboard/bin/chaos-dashboard,./cmd/chaos-dashboard/main.go,1,ui))

$(eval $(call BUILD_IN_DOCKER_TEMPLATE,chaos-mesh,images/chaos-mesh/bin/chaos-controller-manager))
$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-mesh/bin/chaos-controller-manager,./cmd/chaos-controller-manager/main.go,0))

prepare-install: all docker-push docker-push-dns-server

prepare-e2e: e2e-image docker-push-e2e

GINKGO_FLAGS ?=
e2e: e2e-build
	./e2e-test/image/e2e/bin/ginkgo ${GINKGO_FLAGS} ./e2e-test/image/e2e/bin/e2e.test -- --e2e-image ${DOCKER_REGISTRY_PREFIX}pingcap/e2e-helper:${IMAGE_TAG}

image-chaos-mesh-e2e-dependencies += e2e-test/image/e2e/manifests e2e-test/image/e2e/chaos-mesh e2e-build
CLEAN_TARGETS += e2e-test/image/e2e/manifests e2e-test/image/e2e/chaos-mesh

e2e-build: e2e-test/image/e2e/bin/ginkgo e2e-test/image/e2e/bin/e2e.test

e2e-test/image/e2e/manifests: manifests
	rm -rf e2e-test/image/e2e/manifests
	cp -r manifests e2e-test/image/e2e

e2e-test/image/e2e/chaos-mesh: helm/chaos-mesh
	rm -rf e2e-test/image/e2e/chaos-mesh
	cp -r helm/chaos-mesh e2e-test/image/e2e

# $(1): the name of the image
# $(2): the path of the Dockerfile build directory
define IMAGE_TEMPLATE
CLEAN_TARGETS += $(2)/.dockerbuilt

image-$(1): $(2)/.dockerbuilt

$(2)/.dockerbuilt:$(image-$(1)-dependencies) $(2)/Dockerfile
	$(ROOT)/build/build_image.py $(1) $(2)
	touch $(2)/.dockerbuilt
endef

$(eval $(call IMAGE_TEMPLATE,chaos-daemon,images/chaos-daemon,0,CHAOS_DAEMON))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh,images/chaos-mesh,0,CHAOS_MESH))
$(eval $(call IMAGE_TEMPLATE,chaos-dashboard,images/chaos-dashboard,0,CHAOS_DASHBOARD))
$(eval $(call IMAGE_TEMPLATE,build-env,images/build-env,0,BUILD_ENV))
$(eval $(call IMAGE_TEMPLATE,dev-env,images/dev-env,0,DEV_ENV))
$(eval $(call IMAGE_TEMPLATE,e2e-helper,e2e-test/cmd/e2e_helper,0,E2E_HELPER))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh-e2e,e2e-test/image/e2e,0,CHAOS_MESH_E2E))
$(eval $(call IMAGE_TEMPLATE,chaos-kernel,images/chaos-kernel,0,CHAOS_KERNEL))
$(eval $(call IMAGE_TEMPLATE,chaos-jvm,images/chaos-jvm,0,CHAOS_JVM))
$(eval $(call IMAGE_TEMPLATE,chaos-dlv,images/chaos-dlv,0,CHAOS_DLV))

binary: $(BINARIES)

docker-push:
	docker push "${DOCKER_REGISTRY_PREFIX}chaos-mesh/chaos-mesh:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}chaos-mesh/chaos-dashboard:${IMAGE_TAG}"
	docker push "${DOCKER_REGISTRY_PREFIX}chaos-mesh/chaos-daemon:${IMAGE_TAG}"

docker-push-e2e:
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/e2e-helper:${IMAGE_TAG}"

# the version of dns server should keep consistent with helm
DNS_SERVER_VERSION ?= v0.2.0
docker-push-dns-server:
	docker pull pingcap/coredns:${DNS_SERVER_VERSION}
	docker tag pingcap/coredns:${DNS_SERVER_VERSION} "${DOCKER_REGISTRY_PREFIX}pingcap/coredns:${DNS_SERVER_VERSION}"
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/coredns:${DNS_SERVER_VERSION}"

docker-push-chaos-kernel:
	docker push "${DOCKER_REGISTRY_PREFIX}pingcap/chaos-kernel:${IMAGE_TAG}"

RUN_IN_DEV=@$(ROOT)/build/run_in_docker.py dev-env -- 

bin/chaos-builder: image-dev-env
	$(RUN_IN_DEV) $(CGOENV) go build -ldflags \'$(LDFLAGS)\' -o bin/chaos-builder ./cmd/chaos-builder/...

chaos-build: bin/chaos-builder image-dev-env
	$(RUN_IN_DEV) bin/chaos-builder

proto: image-dev-env
ifeq ($(IN_DOCKER),1)
	for dir in pkg/chaosdaemon pkg/chaoskernel ; do\
		protoc -I $$dir/pb $$dir/pb/*.proto -I /usr/local/include --go_out=plugins=grpc:$$dir/pb --go_out=./$$dir/pb ;\
	done
else
	$(RUN_IN_DEV) make proto
endif

manifests/crd.yaml: config image-dev-env
ifeq ($(IN_DOCKER),1)
	kustomize build config/default > manifests/crd.yaml
else
	$(RUN_IN_DEV) make manifests/crd.yaml
endif

manifests/crd-v1beta1.yaml: config image-dev-env
ifeq ($(IN_DOCKER),1)
	mkdir -p ./output
	cp -Tr ./config ./output/config-v1beta1
	cd ./api/v1alpha1 ;\
		controller-gen "crd:trivialVersions=true,preserveUnknownFields=false,crdVersions=v1beta1" rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../../output/config-v1beta1/crd/bases ;
	kustomize build output/config-v1beta1/default > manifests/crd-v1beta1.yaml
else
	$(RUN_IN_DEV) make manifests/crd-v1beta1.yaml
endif

yaml: manifests/crd.yaml manifests/crd-v1beta1.yaml

config: image-dev-env
ifeq ($(IN_DOCKER),1)
	cd ./api/v1alpha1 ;\
		controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../../config/crd/bases ;\
		controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../../helm/chaos-mesh/crds ;
else
	$(RUN_IN_DEV) make config
endif

lint: image-dev-env
ifeq ($(IN_DOCKER),1)
	revive -formatter friendly -config revive.toml $$($(PACKAGE_LIST))
else
	$(RUN_IN_DEV) make lint
endif

failpoint-enable: image-dev-env
ifeq ($(IN_DOCKER),1)
	find $(ROOT)/* -type d | grep -vE "(\.git|bin|\.cache|ui)" | xargs failpoint-ctl enable
else
	$(RUN_IN_DEV) make failpoint-enable
endif

failpoint-disable: image-dev-env
ifeq ($(IN_DOCKER),1)
	find $(ROOT)/* -type d | grep -vE "(\.git|bin|\.cache|ui)" | xargs failpoint-ctl disable
else
	$(RUN_IN_DEV) make failpoint-disable
endif

groupimports: image-dev-env
ifeq ($(IN_DOCKER),1)
	find . -type f -name '*.go' -not -path '**/zz_generated.*.go' -not -path './.cache/**' | xargs \
		-d $$'\n' -n 10 goimports -w -l -local github.com/chaos-mesh/chaos-mesh
else
	$(RUN_IN_DEV) make groupimports
endif

fmt: groupimports image-dev-env
ifeq ($(IN_DOCKER),1)
	$(CGO) fmt $$($(PACKAGE_LIST))
else
	$(RUN_IN_DEV) make fmt
endif

vet: image-dev-env
	$(RUN_IN_DEV) $(CGOENV) go vet ./...

tidy: clean image-dev-env
ifeq ($(IN_DOCKER),1)
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff -U --exit-code go.mod go.sum
	cd api/v1alpha1; GO111MODULE=on go mod tidy; git diff -U --exit-code go.mod go.sum
	cd e2e-test; GO111MODULE=on go mod tidy; git diff -U --exit-code go.mod go.sum
	cd e2e-test/cmd/e2e_helper; GO111MODULE=on go mod tidy; git diff -U --exit-code go.mod go.sum
else
	$(RUN_IN_DEV) make tidy
endif

generate-ctrl: image-dev-env image-dev-env
	$(RUN_IN_DEV) $(GO) generate ./pkg/ctrlserver/graph

generate-deepcopy: image-dev-env
ifeq ($(IN_DOCKER),1)
	cd ./api/v1alpha1 ;\
		controller-gen object:headerFile=../../hack/boilerplate/boilerplate.generatego.txt paths="./..." ;
else
	$(RUN_IN_DEV) make generate-deepcopy
endif

generate: generate-deepcopy chaos-build generate-ctrl swagger_spec

check: generate yaml vet boilerplate lint tidy install.sh fmt

CLEAN_TARGETS+=e2e-test/image/e2e/bin/ginkgo
e2e-test/image/e2e/bin/ginkgo: image-dev-env
	mkdir -p e2e-test/image/e2e/bin
	$(RUN_IN_DEV) cp /go/bin/ginkgo e2e-test/image/e2e/bin/ginkgo

CLEAN_TARGETS+=e2e-test/image/e2e/bin/e2e.test
e2e-test/image/e2e/bin/e2e.test: image-dev-env
	$(RUN_IN_DEV) "cd e2e-test && $(GO) test -c  -o ./image/e2e/bin/e2e.test ./e2e"

# Run tests
CLEAN_TARGETS += cover.out cover.out.tmp
test: failpoint-enable generate generate-mock manifests test-utils
	$(RUN_IN_DEV) CGO_ENABLED=1 $(GOTEST) -p 1 $$($(PACKAGE_LIST)) -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out
	make failpoint-disable

gosec-scan:
	$(RUN_IN_DEV) gosec ./api/... ./controllers/... ./pkg/... || echo "*** sec-scan failed: known-issues ***"

coverage:
ifeq ("$(CI)", "1")
	@bash <(curl -s https://codecov.io/bash) -f cover.out -t $(CODECOV_TOKEN)
else
	mkdir -p cover
	$(RUN_IN_DEV) gocov convert cover.out > cover.json
	$(RUN_IN_DEV) gocov-xml < cover.json > cover.xml
	$(RUN_IN_DEV) gocov-html < cover.json > cover/index.html
endif

install.sh:
	$(RUN_IN_DEV) ./hack/update_install_script.sh

swagger_spec:
	$(RUN_IN_DEV) swag init -g cmd/chaos-dashboard/main.go --output pkg/dashboard/swaggerdocs

generate-mock:
	$(RUN_IN_DEV) $(GO) generate ./pkg/workflow

.PHONY: all clean test install manifests groupimports fmt vet tidy image \
	docker-push lint generate config mockgen generate-mock \
	install.sh $(GO_TARGET_PHONY) \
	manager chaosfs chaosdaemon chaos-dashboard \
	dashboard dashboard-server-frontend gosec-scan \
	failpoint-enable failpoint-disable swagger_spec \
	e2e-test/image/e2e/bin/e2e.test \
	proto bin/chaos-builder go_build_cache_directory schedule-migration enter-buildenv enter-devenv \
	manifests/crd.yaml generate-deepcopy boilerplate boilerplate-fix

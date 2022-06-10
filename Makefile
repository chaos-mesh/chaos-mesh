# Set DEBUGGER=1 to build debug symbols
export LDFLAGS := $(if $(LDFLAGS),$(LDFLAGS),$(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh))
export IMAGE_REGISTRY ?= ghcr.io

# SET IMAGE_REGISTRY to change the docker registry
IMAGE_REGISTRY_PREFIX := $(if $(IMAGE_REGISTRY),$(IMAGE_REGISTRY)/,)

export IMAGE_TAG ?= latest
export IMAGE_PROJECT ?= chaos-mesh
export IMAGE_BUILD ?= 1

# todo: rename the project/repository of e2e-helper to chaos-mesh
export IMAGE_E2E_HELPER_PROJECT ?= pingcap
export IMAGE_CHAOS_MESH_E2E_PROJECT ?= pingcap

ROOT=$(shell pwd)
HELM_BIN=$(ROOT)/output/bin/helm

export IMAGE_BUILD_ENV_BUILD ?= 0
export IMAGE_DEV_ENV_BUILD ?= 0

# Every branch should have its own image tag for build-env and dev-env
# using := with ifeq instead of ?= for performance issue
ifeq ($(IMAGE_BUILD_ENV_TAG),)
export IMAGE_BUILD_ENV_TAG := $(shell ./hack/env-image-tag.sh build-env)
endif
ifeq ($(IMAGE_DEV_ENV_TAG),)
export IMAGE_DEV_ENV_TAG := $(shell ./hack/env-image-tag.sh dev-env)
endif

export GOPROXY  := $(if $(GOPROXY),$(GOPROXY),https://proxy.golang.org,direct)
GOENV  	:= CGO_ENABLED=0
CGOENV 	:= CGO_ENABLED=1
GO     	:= $(GOENV) go
CGO    	:= $(CGOENV) go
GOTEST 	:= USE_EXISTING_CLUSTER=false NO_PROXY="${NO_PROXY},testhost" go test
SHELL  	:= bash

PACKAGE_LIST := echo $$(go list ./... | grep -vE "chaos-mesh/test|pkg/ptrace|zz_generated|vendor") $$(cd api && go list ./... && cd ../)

# no version conversion
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false,crdVersions=v1"

export GO_BUILD_CACHE ?= $(ROOT)/.cache/chaos-mesh
export YARN_BUILD_CACHE ?= $(ROOT)/.cache/yarn

BUILD_TAGS ?=

ifeq ($(UI),1)
	BUILD_TAGS += ui_server
endif

BASIC_IMAGE_ENV=IMAGE_DEV_ENV_PROJECT=$(IMAGE_DEV_ENV_PROJECT) IMAGE_DEV_ENV_REGISTRY=$(IMAGE_DEV_ENV_REGISTRY) \
	IMAGE_DEV_ENV_TAG=$(IMAGE_DEV_ENV_TAG) \
	IMAGE_BUILD_ENV_PROJECT=$(IMAGE_BUILD_ENV_PROJECT) IMAGE_BUILD_ENV_REGISTRY=$(IMAGE_BUILD_ENV_REGISTRY) \
	IMAGE_BUILD_ENV_TAG=$(IMAGE_BUILD_ENV_TAG) IN_DOCKER=$(IN_DOCKER) \
	IMAGE_TAG=$(IMAGE_TAG) IMAGE_PROJECT=$(IMAGE_PROJECT) IMAGE_REGISTRY=$(IMAGE_REGISTRY) \
	TARGET_PLATFORM=$(TARGET_PLATFORM) \
	GO_BUILD_CACHE=$(GO_BUILD_CACHE) YARN_BUILD_CACHE=$(YARN_BUILD_CACHE)

RUN_IN_DEV_SHELL=$(shell $(BASIC_IMAGE_ENV)\
	$(ROOT)/build/get_env_shell.py dev-env)
RUN_IN_BUILD_SHELL=$(shell $(BASIC_IMAGE_ENV)\
	$(ROOT)/build/get_env_shell.py build-env)

CLEAN_TARGETS :=

all: yaml image

test-utils: timer multithread_tracee pkg/time/fakeclock/fake_clock_gettime.o pkg/time/fakeclock/fake_gettimeofday.o

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

watchmaker: pkg/time/fakeclock/fake_clock_gettime.o pkg/time/fakeclock/fake_gettimeofday.o
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
	$(HELM_BIN) upgrade --install chaos-mesh helm/chaos-mesh --namespace=${NAMESPACE} --set images.registry=${IMAGE_REGISTRY} --set dnsServer.create=true --set dashboard.create=true;

clean:
	rm -rf $(CLEAN_TARGETS)

SKYWALKING_EYES_HEADER = /go/bin/license-eye header -c ./.github/.licenserc.yaml
boilerplate: SHELL:=$(RUN_IN_DEV_SHELL)
boilerplate: images/dev-env/.dockerbuilt
	$(SKYWALKING_EYES_HEADER) check

boilerplate-fix: SHELL:=$(RUN_IN_DEV_SHELL)
boilerplate-fix: images/dev-env/.dockerbuilt
	$(SKYWALKING_EYES_HEADER) fix

image: image-chaos-daemon image-chaos-mesh image-chaos-dashboard $(if $(DEBUGGER), image-chaos-dlv)

e2e-image: image-e2e-helper

GO_TARGET_PHONY :=

define COMPILE_GO_TEMPLATE

$(1): SHELL:=$(RUN_IN_BUILD_SHELL)
$(1): $(4) image-build-env
ifeq ($(3),1)
	$(CGO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o $(1) $(2)
else
	$(GO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o $(1) $(2)
endif

GO_TARGET_PHONY += $(1)
CLEAN_TARGETS += $(1)
endef

enter-buildenv: SHELL:=$(shell $(BASIC_IMAGE_ENV) $(ROOT)/build/get_env_shell.py --interactive build-env)
enter-buildenv: image-build-env
	@bash

enter-devenv: SHELL:=$(shell $(BASIC_IMAGE_ENV) $(ROOT)/build/get_env_shell.py --interactive dev-env)
enter-devenv: images/dev-env/.dockerbuilt
	@bash

images/chaos-daemon/bin/pause: SHELL:=$(RUN_IN_BUILD_SHELL)
images/chaos-daemon/bin/pause: hack/pause.c images/build-env/.dockerbuilt
	cc ./hack/pause.c -o images/chaos-daemon/bin/pause

pkg/time/fakeclock/fake_clock_gettime.o: SHELL:=$(RUN_IN_BUILD_SHELL)
pkg/time/fakeclock/fake_clock_gettime.o: pkg/time/fakeclock/fake_clock_gettime.c images/build-env/.dockerbuilt
	[[ "$$TARGET_PLATFORM" == "arm64" ]] && CFLAGS="-mcmodel=tiny" ;\
	cc -c ./pkg/time/fakeclock/fake_clock_gettime.c -fPIE -O2 -o pkg/time/fakeclock/fake_clock_gettime.o $$CFLAGS
pkg/time/fakeclock/fake_gettimeofday.o: SHELL:=$(RUN_IN_BUILD_SHELL)
pkg/time/fakeclock/fake_gettimeofday.o: pkg/time/fakeclock/fake_gettimeofday.c images/build-env/.dockerbuilt
	[[ "$$TARGET_PLATFORM" == "arm64" ]] && CFLAGS="-mcmodel=tiny" ;\
	cc -c ./pkg/time/fakeclock/fake_gettimeofday.c -fPIE -O2 -o pkg/time/fakeclock/fake_gettimeofday.o $$CFLAGS

$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-daemon/bin/chaos-daemon,./cmd/chaos-daemon/main.go,1,pkg/time/fakeclock/fake_clock_gettime.o pkg/time/fakeclock/fake_gettimeofday.o))
$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-dashboard/bin/chaos-dashboard,./cmd/chaos-dashboard/main.go,1,ui))
$(eval $(call COMPILE_GO_TEMPLATE,images/chaos-mesh/bin/chaos-controller-manager,./cmd/chaos-controller-manager/main.go,0))

prepare-install: all docker-push docker-push-dns-server

prepare-e2e: e2e-image docker-push-e2e

GINKGO_FLAGS ?=
e2e: e2e-build
	./e2e-test/image/e2e/bin/ginkgo ${GINKGO_FLAGS} ./e2e-test/image/e2e/bin/e2e.test -- --e2e-image ${IMAGE_REGISTRY_PREFIX}pingcap/e2e-helper:${IMAGE_TAG}

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
# $(3): the dependency of the image
define IMAGE_TEMPLATE
CLEAN_TARGETS += $(2)/.dockerbuilt

image-$(1): $(2)/.dockerbuilt

$(2)/.dockerbuilt:SHELL=bash
$(2)/.dockerbuilt:$(3) $(2)/Dockerfile
	$(ROOT)/build/build_image.py $(1) $(2)
	touch $(2)/.dockerbuilt
endef

$(eval $(call IMAGE_TEMPLATE,chaos-daemon,images/chaos-daemon,images/chaos-daemon/bin/chaos-daemon images/chaos-daemon/bin/pause))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh,images/chaos-mesh,images/chaos-mesh/bin/chaos-controller-manager))
$(eval $(call IMAGE_TEMPLATE,chaos-dashboard,images/chaos-dashboard,images/chaos-dashboard/bin/chaos-dashboard))
$(eval $(call IMAGE_TEMPLATE,build-env,images/build-env))
$(eval $(call IMAGE_TEMPLATE,dev-env,images/dev-env))
$(eval $(call IMAGE_TEMPLATE,e2e-helper,e2e-test/cmd/e2e_helper))
$(eval $(call IMAGE_TEMPLATE,chaos-mesh-e2e,e2e-test/image/e2e,e2e-test/image/e2e/manifests e2e-test/image/e2e/chaos-mesh e2e-build))
$(eval $(call IMAGE_TEMPLATE,chaos-kernel,images/chaos-kernel))
$(eval $(call IMAGE_TEMPLATE,chaos-jvm,images/chaos-jvm))
$(eval $(call IMAGE_TEMPLATE,chaos-dlv,images/chaos-dlv))

docker-push:
	docker push "${IMAGE_REGISTRY_PREFIX}chaos-mesh/chaos-mesh:${IMAGE_TAG}"
	docker push "${IMAGE_REGISTRY_PREFIX}chaos-mesh/chaos-dashboard:${IMAGE_TAG}"
	docker push "${IMAGE_REGISTRY_PREFIX}chaos-mesh/chaos-daemon:${IMAGE_TAG}"

docker-push-e2e:
	docker push "${IMAGE_REGISTRY_PREFIX}pingcap/e2e-helper:${IMAGE_TAG}"

# the version of dns server should keep consistent with helm
DNS_SERVER_VERSION ?= v0.2.0
docker-push-dns-server:
	docker pull pingcap/coredns:${DNS_SERVER_VERSION}
	docker tag pingcap/coredns:${DNS_SERVER_VERSION} "${IMAGE_REGISTRY_PREFIX}pingcap/coredns:${DNS_SERVER_VERSION}"
	docker push "${IMAGE_REGISTRY_PREFIX}pingcap/coredns:${DNS_SERVER_VERSION}"

docker-push-chaos-kernel:
	docker push "${IMAGE_REGISTRY_PREFIX}pingcap/chaos-kernel:${IMAGE_TAG}"

bin/chaos-builder: SHELL:=$(RUN_IN_DEV_SHELL)
bin/chaos-builder: images/dev-env/.dockerbuilt
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -o bin/chaos-builder ./cmd/chaos-builder/...

chaos-build: SHELL:=$(RUN_IN_DEV_SHELL)
chaos-build: bin/chaos-builder images/dev-env/.dockerbuilt
	bin/chaos-builder

proto: SHELL:=$(RUN_IN_DEV_SHELL)
proto: images/dev-env/.dockerbuilt
	for dir in pkg/chaosdaemon pkg/chaoskernel ; do\
		protoc -I $$dir/pb $$dir/pb/*.proto -I /usr/local/include --go_out=plugins=grpc:$$dir/pb --go_out=./$$dir/pb ;\
	done

manifests/crd.yaml: SHELL:=$(RUN_IN_DEV_SHELL)
manifests/crd.yaml: config images/dev-env/.dockerbuilt
	kustomize build config/default > manifests/crd.yaml

manifests/crd-v1beta1.yaml: SHELL:=$(RUN_IN_DEV_SHELL)
manifests/crd-v1beta1.yaml: config images/dev-env/.dockerbuilt
	mkdir -p ./output
	cp -Tr ./config ./output/config-v1beta1
	cd ./api ;\
		controller-gen "crd:trivialVersions=true,preserveUnknownFields=false,crdVersions=v1beta1" rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../output/config-v1beta1/crd/bases ;
	kustomize build output/config-v1beta1/default > manifests/crd-v1beta1.yaml

yaml: manifests/crd.yaml manifests/crd-v1beta1.yaml

config: SHELL:=$(RUN_IN_DEV_SHELL)
config: images/dev-env/.dockerbuilt
	cd ./api ;\
		controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../config/crd/bases ;\
		controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../helm/chaos-mesh/crds ;

lint: SHELL:=$(RUN_IN_DEV_SHELL)
lint: images/dev-env/.dockerbuilt
	revive -formatter friendly -config revive.toml $$($(PACKAGE_LIST))

failpoint-enable: SHELL:=$(RUN_IN_DEV_SHELL)
failpoint-enable: images/dev-env/.dockerbuilt
	find $(ROOT)/* -type d | grep -vE "(\.git|bin|\.cache|ui)" | xargs failpoint-ctl enable

failpoint-disable: SHELL:=$(RUN_IN_DEV_SHELL)
failpoint-disable: images/dev-env/.dockerbuilt
	find $(ROOT)/* -type d | grep -vE "(\.git|bin|\.cache|ui)" | xargs failpoint-ctl disable

groupimports: SHELL:=$(RUN_IN_DEV_SHELL)
groupimports: images/dev-env/.dockerbuilt
	find . -type f -name '*.go' -not -path '**/zz_generated.*.go' -not -path './.cache/**' | xargs \
		-d $$'\n' -n 10 goimports -combine -w -l -local github.com/chaos-mesh/chaos-mesh

fmt: SHELL:=$(RUN_IN_DEV_SHELL)
fmt: groupimports images/dev-env/.dockerbuilt
	$(CGO) fmt $$($(PACKAGE_LIST))

vet: SHELL:=$(RUN_IN_DEV_SHELL)
vet: images/dev-env/.dockerbuilt
	$(CGOENV) go vet ./...

tidy: SHELL:=$(RUN_IN_DEV_SHELL)
tidy: images/dev-env/.dockerbuilt
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff -U --exit-code go.mod go.sum
	cd api; GO111MODULE=on go mod tidy; git diff -U --exit-code go.mod go.sum
	cd e2e-test; GO111MODULE=on go mod tidy; git diff -U --exit-code go.mod go.sum
	cd e2e-test/cmd/e2e_helper; GO111MODULE=on go mod tidy; git diff -U --exit-code go.mod go.sum

generate-ctrl: SHELL:=$(RUN_IN_DEV_SHELL)
generate-ctrl: images/dev-env/.dockerbuilt generate-deepcopy
	$(GO) generate ./pkg/ctrl/server

generate-deepcopy: SHELL:=$(RUN_IN_DEV_SHELL)
generate-deepcopy: images/dev-env/.dockerbuilt chaos-build
	cd ./api ;\
		controller-gen object:headerFile=../hack/boilerplate/boilerplate.generatego.txt paths="./..." ;

generate: generate-ctrl swagger_spec generate-deepcopy chaos-build

check: generate yaml vet boilerplate lint tidy install.sh fmt

CLEAN_TARGETS+=e2e-test/image/e2e/bin/ginkgo
e2e-test/image/e2e/bin/ginkgo: SHELL:=$(RUN_IN_DEV_SHELL)
e2e-test/image/e2e/bin/ginkgo: images/dev-env/.dockerbuilt
	mkdir -p e2e-test/image/e2e/bin
	cp /go/bin/ginkgo e2e-test/image/e2e/bin/ginkgo

CLEAN_TARGETS+=e2e-test/image/e2e/bin/e2e.test
e2e-test/image/e2e/bin/e2e.test: SHELL:=$(RUN_IN_DEV_SHELL)
e2e-test/image/e2e/bin/e2e.test: images/dev-env/.dockerbuilt
	cd e2e-test && $(GO) test -c  -o ./image/e2e/bin/e2e.test ./e2e

# Run tests
CLEAN_TARGETS += cover.out cover.out.tmp

test: SHELL:=$(RUN_IN_DEV_SHELL)
test: failpoint-enable generate manifests test-utils images/dev-env/.dockerbuilt
	CGO_ENABLED=1 $(GOTEST) -p 1 $$($(PACKAGE_LIST)) -coverprofile cover.out.tmp -covermode=atomic
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out
	make failpoint-disable

gosec-scan: SHELL:=$(RUN_IN_DEV_SHELL)
gosec-scan: images/dev-env/.dockerbuilt
	gosec ./api/... ./controllers/... ./pkg/... || echo "*** sec-scan failed: known-issues ***"

coverage: SHELL:=$(RUN_IN_DEV_SHELL)
coverage: images/dev-env/.dockerbuilt
ifeq ("$(CI)", "1")
	@bash <(curl -s https://codecov.io/bash) -f cover.out -t $(CODECOV_TOKEN)
else
	mkdir -p cover
	gocov convert cover.out > cover.json
	gocov-xml < cover.json > cover.xml
	gocov-html < cover.json > cover/index.html
endif

install.sh: SHELL:=$(RUN_IN_DEV_SHELL)
install.sh: images/dev-env/.dockerbuilt
	./hack/update_install_script.sh

swagger_spec:SHELL:=$(RUN_IN_DEV_SHELL)
swagger_spec: images/dev-env/.dockerbuilt
	swag init -g cmd/chaos-dashboard/main.go --output pkg/dashboard/swaggerdocs

.PHONY: all clean test install manifests groupimports fmt vet tidy image \
	docker-push lint generate config \
	install.sh $(GO_TARGET_PHONY) \
	manager chaosfs chaosdaemon chaos-dashboard \
	dashboard dashboard-server-frontend gosec-scan \
	failpoint-enable failpoint-disable swagger_spec \
	e2e-test/image/e2e/bin/e2e.test \
	proto bin/chaos-builder go_build_cache_directory schedule-migration enter-buildenv enter-devenv \
	manifests/crd.yaml generate-deepcopy boilerplate boilerplate-fix

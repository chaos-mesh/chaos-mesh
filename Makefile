# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

## --------------------------------------
## General
## --------------------------------------

.DEFAULT_GOAL:=help

# Set DEBUGGER=1 to build debug symbols
export LDFLAGS := $(if $(LDFLAGS),$(LDFLAGS),$(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh))

export IMAGE_TAG ?= latest
export IMAGE_BUILD ?= 1

ROOT=$(shell pwd)

export IMAGE_BUILD_ENV_BUILD ?= 0
export IMAGE_DEV_ENV_BUILD ?= 0

export GOPROXY := $(if $(GOPROXY),$(GOPROXY),https://proxy.golang.org,direct)
GOENV  := CGO_ENABLED=0
CGOENV := CGO_ENABLED=1
GO     := $(GOENV) go
CGO    := $(CGOENV) go
GOTEST := USE_EXISTING_CLUSTER=false NO_PROXY="$(NO_PROXY),testhost" go test
SHELL  := bash

PACKAGE_LIST := echo $$(go list ./... | grep -vE "chaos-mesh/test|pkg/ptrace|zz_generated|vendor") $$(cd api && go list ./... && cd ../)

export GO_BUILD_CACHE ?= $(ROOT)/.cache/chaos-mesh

BUILD_TAGS ?=

ifeq ($(UI),1)
	BUILD_TAGS += ui_server
endif

# See https://github.com/chaos-mesh/chaos-mesh/pull/4004 for more details.
ifeq (,$(findstring local/,$(MAKECMDGOALS)))

# Each branch should have its own image tag for build-env and dev-env
# Use := with ifeq instead of = for performance issues (simply expanded)
ifeq ($(IMAGE_BUILD_ENV_TAG),)
export IMAGE_BUILD_ENV_TAG := $(shell ./hack/env-image-tag.sh build-env)
endif
ifeq ($(IMAGE_DEV_ENV_TAG),)
export IMAGE_DEV_ENV_TAG := $(shell ./hack/env-image-tag.sh dev-env)
endif

endif

BASIC_IMAGE_ENV= IMAGE_DEV_ENV_TAG=$(IMAGE_DEV_ENV_TAG) \
	IMAGE_BUILD_ENV_TAG=$(IMAGE_BUILD_ENV_TAG) \
	IMAGE_TAG=$(IMAGE_TAG) TARGET_PLATFORM=$(TARGET_PLATFORM) \
	GO_BUILD_CACHE=$(GO_BUILD_CACHE)

RUN_IN_DEV_SHELL=$(shell $(BASIC_IMAGE_ENV)\
	$(ROOT)/build/get_env_shell.py dev-env)
RUN_IN_BUILD_SHELL=$(shell $(BASIC_IMAGE_ENV)\
	$(ROOT)/build/get_env_shell.py build-env)

# See https://github.com/chaos-mesh/chaos-mesh/pull/4004 for more details.
ifeq (,$(findstring local/,$(MAKECMDGOALS)))

# Include generated makefiles.
# These sub makefiles depend on RUN_IN_DEV_SHELL and RUN_IN_BUILD_SHELL, so it should be included after them.
include binary.generated.mk container-image.generated.mk

endif

include local-binary.generated.mk

export CLEAN_TARGETS :=

# The help will print out all targets with their descriptions organized bellow their categories. The categories are represented by `##@` and the target descriptions by `##`.
# The awk commands is responsible to read the entire set of makefiles included in this invocation, looking for lines of the file as xyz: ## something, and then pretty-format the target and help. Then, if there's a line with ##@ something, that gets pretty-printed as a category.
# More info over the usage of ANSI control characters for terminal formatting: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info over awk command: http://linuxcommand.org/lc3_adv_awk.php
#
# Notice that we have a little modification on the awk command to support slash in the recipe name:
# origin: /^[a-zA-Z_0-9-]+:.*?##/
# modified /^[a-zA-Z_0-9\/\.-]+:.*?##/
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9\/\.-]+:.*?##/ { printf "  \033[36m%-21s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Code generation

config: SHELL:=$(RUN_IN_DEV_SHELL)
config: images/dev-env/.dockerbuilt ## Generate CRD manifests with controller-gen
	cd ./api ;\
		controller-gen crd:trivialVersions=true,preserveUnknownFields=false,crdVersions=v1 rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../config/crd/bases ;\
		controller-gen crd:trivialVersions=true,preserveUnknownFields=false,crdVersions=v1 rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=../helm/chaos-mesh/crds ;

chaos-build: SHELL:=$(RUN_IN_DEV_SHELL)
chaos-build: bin/chaos-builder images/dev-env/.dockerbuilt ## Generate codes for CustomResource Kinds under api/v1alpha1
	bin/chaos-builder

generate: manifests/crd.yaml generate-ctrl swagger_spec generate-deepcopy chaos-build ## Generate codes for codebase, including CRD manifests, chaosctl GraphQL code generation, chaos mesh controller code generation, deepcopy, swager spec.

generate-ctrl: SHELL:=$(RUN_IN_DEV_SHELL)
generate-ctrl: images/dev-env/.dockerbuilt generate-deepcopy ## Generate GraphQL schema for chaosctl
	$(GO) generate ./pkg/ctrl/server

.PHONY: generate-makefile
generate-makefile: ## Generate makefile (binary.generated.mk, container-image.generated.mk)
	@$(GO) run ./cmd/generate-makefile

generate-deepcopy: SHELL:=$(RUN_IN_DEV_SHELL)
generate-deepcopy: images/dev-env/.dockerbuilt chaos-build ## Generate deepcopy files for CRD Kind with controller-gen
	cd ./api ;\
		controller-gen object:headerFile=../hack/boilerplate/boilerplate.generatego.txt paths="./..." ;

install.sh: SHELL:=$(RUN_IN_DEV_SHELL)
install.sh: images/dev-env/.dockerbuilt ## Generate install.sh
	./hack/update_install_script.sh

manifests/crd.yaml: SHELL:=$(RUN_IN_DEV_SHELL)
manifests/crd.yaml: config images/dev-env/.dockerbuilt ## Generate the combined CRD manifests
	kustomize build config/default > manifests/crd.yaml

proto: SHELL:=$(RUN_IN_DEV_SHELL)
proto: images/dev-env/.dockerbuilt ## Generate .go files from .proto files
	for dir in pkg/chaosdaemon pkg/chaoskernel ; do\
		protoc -I $$dir/pb $$dir/pb/*.proto -I /usr/local/include --go_out=plugins=grpc:$$dir/pb --go_out=./$$dir/pb ;\
	done

swagger_spec: SHELL:=$(RUN_IN_DEV_SHELL)
swagger_spec: images/dev-env/.dockerbuilt ## Generate OpenAPI/Swagger spec for frontend
	swag init -g cmd/chaos-dashboard/main.go --output pkg/dashboard/swaggerdocs --pd --parseInternal

##@ Linters, formatters and others

check: generate manifests/crd.yaml vet boilerplate lint tidy install.sh fmt ## Run prerequisite checks for PR

SKYWALKING_EYES_HEADER = /go/bin/license-eye header -c ./.github/.licenserc.yaml
boilerplate: SHELL:=$(RUN_IN_DEV_SHELL)
boilerplate: images/dev-env/.dockerbuilt
	$(SKYWALKING_EYES_HEADER) check

boilerplate-fix: SHELL:=$(RUN_IN_DEV_SHELL)
boilerplate-fix: images/dev-env/.dockerbuilt ## Fix boilerplate
	$(SKYWALKING_EYES_HEADER) fix

fmt: SHELL:=$(RUN_IN_DEV_SHELL)
fmt: groupimports images/dev-env/.dockerbuilt ## Reformat go files with gofmt and goimports
	$(CGO) fmt $$($(PACKAGE_LIST))

gosec-scan: SHELL:=$(RUN_IN_DEV_SHELL)
gosec-scan: images/dev-env/.dockerbuilt
	gosec ./api/... ./controllers/... ./pkg/... || echo "*** sec-scan failed: known-issues ***"

groupimports: SHELL:=$(RUN_IN_DEV_SHELL)
groupimports: images/dev-env/.dockerbuilt ## Reformat go files with goimports
	find . -type f -name '*.go' -not -path '**/zz_generated.*.go' -not -path './.cache/**' | xargs \
		-d $$'\n' -n 10 goimports -combine -w -l -local github.com/chaos-mesh/chaos-mesh

lint: SHELL:=$(RUN_IN_DEV_SHELL)
lint: images/dev-env/.dockerbuilt ## Lint go files with revive
	revive -formatter friendly -config revive.toml $$($(PACKAGE_LIST))

tidy: SHELL:=$(RUN_IN_DEV_SHELL)
tidy: images/dev-env/.dockerbuilt ## Run go mod tidy in all submodules
	go mod tidy
	git diff -U --exit-code go.mod go.sum
	cd api; go mod tidy; git diff -U --exit-code go.mod go.sum
	cd e2e-test; go mod tidy; git diff -U --exit-code go.mod go.sum
	cd e2e-test/cmd/e2e_helper; go mod tidy; git diff -U --exit-code go.mod go.sum

vet: SHELL:=$(RUN_IN_DEV_SHELL)
vet: images/dev-env/.dockerbuilt ## Lint go files with go vet
	$(CGOENV) go vet ./...

##@ Common used building targets

all: manifests/crd.yaml image ## Build all CRD yaml manifests and components container images

chaosctl: ## Build chaosctl
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/chaosctl ./cmd/chaosctl/*.go

image: image-chaos-daemon image-chaos-mesh image-chaos-dashboard $(if $(DEBUGGER), image-chaos-dlv) ## Build container images for Chaos Mesh components (chaos-controller-manager, chaos-daemon, chaos-dashboard)

ui: pnpm_install_dependencies ## Build the frontend UI of Chaos Dashboard
ifeq (${UI},1)
	cd ui &&\
	pnpm build
	hack/embed_ui_assets.sh
endif

##@ Cleanup

.PHONY: clean
clean: clean-binary clean-image-built ## Cleanup artifacts
	rm -rf $(CLEAN_TARGETS)

##@ Tests

CLEAN_TARGETS += cover.out cover.out.tmp

coverage: SHELL:=$(RUN_IN_DEV_SHELL)
coverage: images/dev-env/.dockerbuilt ## Generate coverage report
ifeq ("$(CI)", "1")
	@bash <(curl -s https://codecov.io/bash) -f cover.out -t $(CODECOV_TOKEN)
else
	mkdir -p cover
	gocov convert cover.out > cover.json
	gocov-xml < cover.json > cover.xml
	gocov-html < cover.json > cover/index.html
endif

GINKGO_FLAGS ?=
PAUSE_IMAGE ?= gcr.io/google-containers/pause:latest
e2e: e2e-build ## Run e2e tests in current kubernetes cluster
	./e2e-test/image/e2e/bin/ginkgo ${GINKGO_FLAGS} ./e2e-test/image/e2e/bin/e2e.test -- --e2e-image ghcr.io/chaos-mesh/e2e-helper:${IMAGE_TAG} --pause-image ${PAUSE_IMAGE}

test: SHELL:=$(RUN_IN_DEV_SHELL)
test: generate manifests test-utils images/dev-env/.dockerbuilt ## Run unit tests
	make failpoint-enable
	CGO_ENABLED=1 $(GOTEST) -p 1 $$($(PACKAGE_LIST)) -coverprofile cover.out.tmp -covermode=atomic
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out
	make failpoint-disable

##@ Advanced building targets

test-utils: timer multithread_tracee pkg/time/fakeclock/fake_clock_gettime.o pkg/time/fakeclock/fake_gettimeofday.o

timer:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/test/timer ./test/cmd/timer/*.go

multithread_tracee: test/cmd/multithread_tracee/main.c
	cc test/cmd/multithread_tracee/main.c -lpthread -O2 -o ./bin/test/multithread_tracee

pnpm_install_dependencies:
ifeq (${UI},1)
	cd ui &&\
	pnpm install --frozen-lockfile
endif

watchmaker: pkg/time/fakeclock/fake_clock_gettime.o pkg/time/fakeclock/fake_gettimeofday.o
	$(CGO) build -ldflags '$(LDFLAGS)' -o bin/watchmaker ./cmd/watchmaker/...

# Build schedule-migration
schedule-migration:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/schedule-migration ./tools/schedule-migration/*.go

schedule-migration.tar.gz: schedule-migration
	cp ./bin/schedule-migration ./schedule-migration
	cp ./tools/schedule-migration/migrate.sh ./migrate.sh
	tar -czvf schedule-migration.tar.gz schedule-migration migrate.sh
	rm ./migrate.sh
	rm ./schedule-migration

e2e-image: image-e2e-helper ## Build e2e test helper image

enter-buildenv: SHELL:=$(shell $(BASIC_IMAGE_ENV) $(ROOT)/build/get_env_shell.py --interactive build-env)
enter-buildenv: image-build-env
	@bash

enter-devenv: SHELL:=$(shell $(BASIC_IMAGE_ENV) $(ROOT)/build/get_env_shell.py --interactive dev-env)
enter-devenv: images/dev-env/.dockerbuilt
	@bash

images/chaos-daemon/bin/pause: SHELL:=$(RUN_IN_BUILD_SHELL)
images/chaos-daemon/bin/pause: hack/pause.c images/build-env/.dockerbuilt ## Build binary pause
	cc ./hack/pause.c -o images/chaos-daemon/bin/pause

.PHONY: pkg/time/fakeclock/fake_clock_gettime.o
pkg/time/fakeclock/fake_clock_gettime.o: SHELL:=$(RUN_IN_BUILD_SHELL)
pkg/time/fakeclock/fake_clock_gettime.o: pkg/time/fakeclock/fake_clock_gettime.c images/build-env/.dockerbuilt
	[[ "$$TARGET_PLATFORM" == "arm64" ]] && CFLAGS="-mcmodel=tiny" ;\
	cc -c ./pkg/time/fakeclock/fake_clock_gettime.c -fPIE -O2 -o pkg/time/fakeclock/fake_clock_gettime.o $$CFLAGS
pkg/time/fakeclock/fake_gettimeofday.o: SHELL:=$(RUN_IN_BUILD_SHELL)
pkg/time/fakeclock/fake_gettimeofday.o: pkg/time/fakeclock/fake_gettimeofday.c images/build-env/.dockerbuilt
	[[ "$$TARGET_PLATFORM" == "arm64" ]] && CFLAGS="-mcmodel=tiny" ;\
	cc -c ./pkg/time/fakeclock/fake_gettimeofday.c -fPIE -O2 -o pkg/time/fakeclock/fake_gettimeofday.o $$CFLAGS


CLEAN_TARGETS += e2e-test/image/e2e/manifests e2e-test/image/e2e/chaos-mesh

e2e-test/image/e2e/manifests: manifests ## Copy CRD manifests to e2e image build directory
	rm -rf e2e-test/image/e2e/manifests
	cp -r manifests e2e-test/image/e2e

e2e-test/image/e2e/chaos-mesh: helm/chaos-mesh ## Copy helm chart to e2e image build directory
	rm -rf e2e-test/image/e2e/chaos-mesh
	cp -r helm/chaos-mesh e2e-test/image/e2e

CLEAN_TARGETS+=e2e-test/image/e2e/bin/ginkgo
e2e-test/image/e2e/bin/ginkgo: SHELL:=$(RUN_IN_DEV_SHELL)
e2e-test/image/e2e/bin/ginkgo: images/dev-env/.dockerbuilt
	mkdir -p e2e-test/image/e2e/bin
	cp /go/bin/ginkgo e2e-test/image/e2e/bin/ginkgo

CLEAN_TARGETS+=e2e-test/image/e2e/bin/e2e.test
e2e-test/image/e2e/bin/e2e.test: SHELL:=$(RUN_IN_DEV_SHELL)
e2e-test/image/e2e/bin/e2e.test: images/dev-env/.dockerbuilt
	cd e2e-test && $(GO) test -c  -o ./image/e2e/bin/e2e.test ./e2e

e2e-build: e2e-test/image/e2e/bin/ginkgo e2e-test/image/e2e/bin/e2e.test ## Build e2e test binary

bin/chaos-builder: SHELL:=$(RUN_IN_DEV_SHELL)
bin/chaos-builder: images/dev-env/.dockerbuilt
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -o bin/chaos-builder ./cmd/chaos-builder/...

failpoint-enable: SHELL:=$(RUN_IN_DEV_SHELL)
failpoint-enable: images/dev-env/.dockerbuilt ## Enable failpoint stub for testing
	find $(ROOT)/* -type d | grep -vE "(\.git|bin|\.cache|ui)" | xargs failpoint-ctl enable

failpoint-disable: SHELL:=$(RUN_IN_DEV_SHELL)
failpoint-disable: images/dev-env/.dockerbuilt ## Disable failpoint stub for testing
	find $(ROOT)/* -type d | grep -vE "(\.git|bin|\.cache|ui)" | xargs failpoint-ctl disable

.PHONY: all image clean test manifests manifests/crd.yaml \
	boilerplate tidy groupimports fmt vet lint install.sh schedule-migration \
	config proto \
	generate generate-deepcopy swagger_spec bin/chaos-builder \
	gosec-scan \
	failpoint-enable failpoint-disable \

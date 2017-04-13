# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

all: images

# Some env vars that devs might find useful:
#  GOFLAGS      : extra "go build" flags to use - e.g. -v   (for verbose)
#  TEST_DIRS=   : only run the unit tests from the specified dirs
#  UNIT_TESTS=  : only run the unit tests matching the specified regexp

# Define some constants
#######################
ROOT           = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BINDIR        ?= bin
BUILD_DIR     ?= build
COVERAGE      ?= $(CURDIR)/coverage.html
BROKER_PKG     = github.com/openshift/brokersdk
TOP_SRC_DIRS   = cmd pkg
SRC_DIRS       = $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*.go \
                   -exec dirname {} \\; | sort | uniq")
TEST_DIRS     ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                   -exec dirname {} \\; | sort | uniq")
VERSION       ?= $(shell git describe --tags --always --abbrev=7 --dirty)
ifeq ($(shell uname -s),Darwin)
STAT           = stat -f '%c %N'
else
STAT           = stat -c '%Y %n'
endif
NEWEST_GO_FILE = $(shell find $(SRC_DIRS) -name \*.go -exec $(STAT) {} \; \
                   | sort -r | head -n 1 | sed "s/.* //")
TYPES_FILES    = $(shell find pkg/apis -name types.go)
GO_VERSION     = 1.7.3

PLATFORM?=linux
ARCH?=amd64

GO_BUILD       = env GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -i $(GOFLAGS) \
                   -ldflags "-X $(BROKER_PKG)/pkg.VERSION=$(VERSION)"
BASE_PATH      = $(ROOT:/src/github.com/openshift/brokersdk/=)
export GOPATH  = $(BASE_PATH):$(ROOT)/vendor

REGISTRY      ?= quay.io/kubernetes-service-catalog/

# precheck to avoid kubernetes-incubator/service-catalog#361
$(if $(realpath vendor/k8s.io/kubernetes/vendor), \
	$(error the vendor directory exists in the kubernetes \
		vendored source and must be flattened. \
		run 'glide i -v'))

ifdef UNIT_TESTS
	UNIT_TEST_FLAGS=-run $(UNIT_TESTS) -v
endif

NON_VENDOR_DIRS = $(shell glide nv)

# This section builds the output binaries.
# Some will have dedicated targets to make it easier to type, for example
# "apiserver" instead of "bin/apiserver".
#########################################################################
build: .init .generate_files broker

broker: $(BINDIR)/broker
$(BINDIR)/broker: .init .generate_files $(NEWEST_GO_FILE)
$(BINDIR)/broker: .init $(NEWEST_GO_FILE)
	$(GO_BUILD) -o $@ $(BROKER_PKG)/cmd/broker

# This section contains the code generation stuff
#################################################
.generate_exes: $(BINDIR)/defaulter-gen \
                $(BINDIR)/deepcopy-gen \
                $(BINDIR)/conversion-gen \
                $(BINDIR)/client-gen \
                $(BINDIR)/lister-gen \
                $(BINDIR)/informer-gen
	touch $@

$(BINDIR)/defaulter-gen: .init
	go build -o $@ $(BROKER_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen

$(BINDIR)/deepcopy-gen: .init
	go build -o $@ $(BROKER_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen

$(BINDIR)/conversion-gen: .init
	go build -o $@ $(BROKER_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen

$(BINDIR)/client-gen: .init
	go build -o $@ $(BROKER_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/client-gen

$(BINDIR)/lister-gen: .init
	go build -o $@ $(BROKER_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/lister-gen

$(BINDIR)/informer-gen: .init
	go build -o $@ $(BROKER_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/informer-gen

#$(BINDIR)/openapi-gen: vendor/k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen
#	$(DOCKER_CMD) go build -o $@ $(BROKER_PKG)/$^

# Regenerate all files if the gen exes changed or any "types.go" files changed
.generate_files: .init .generate_exes $(TYPES_FILES)
	# Generate defaults
	$(BINDIR)/defaulter-gen \
		--v 1 --logtostderr \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(BROKER_PKG)/pkg/apis/broker" \
		--input-dirs "$(BROKER_PKG)/pkg/apis/broker/v1alpha1" \
	  	--extra-peer-dirs "$(BROKER_PKG)/pkg/apis/broker" \
		--extra-peer-dirs "$(BROKER_PKG)/pkg/apis/broker/v1alpha1" \
		--output-file-base "zz_generated.defaults"
	# Generate deep copies
	$(BINDIR)/deepcopy-gen \
		--v 1 --logtostderr \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(BROKER_PKG)/pkg/apis/broker" \
		--input-dirs "$(BROKER_PKG)/pkg/apis/broker/v1alpha1" \
		--bounding-dirs "github.com/openshift/brokersdk" \
		--output-file-base zz_generated.deepcopy
	# Generate conversions
	$(BINDIR)/conversion-gen \
		--v 1 --logtostderr \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(BROKER_PKG)/pkg/apis/broker" \
		--input-dirs "$(BROKER_PKG)/pkg/apis/broker/v1alpha1" \
		--output-file-base zz_generated.conversion
	# the previous three directories will be changed from kubernetes to apimachinery in the future
	# generate all pkg/client contents
	$(BUILD_DIR)/update-client-gen.sh

# Some prereq stuff
###################
.init: glide.yaml
	glide install --strip-vendor --strip-vcs --update-vendored
	touch $@

# Util targets
##############
.PHONY: verify verify-client-gen 
verify: .init .generate_files verify-client-gen
	@echo Running gofmt:
	@gofmt -l -s $(TOP_SRC_DIRS) > .out 2>&1 || true
	@bash -c '[ "`cat .out`" == "" ] || \
	  (echo -e "\n*** Please 'gofmt' the following:" ; cat .out ; echo ; false)'
	@rm .out
	@#
	@echo Running golint and go vet:
	@# Exclude the generated (zz) files for now, as well as defaults.go (it
	@# observes conventions from upstream that will not pass lint checks).
	@sh -c \
	  'for i in $$(find $(TOP_SRC_DIRS) -name *.go \
	    | grep -v generated \
	    | grep -v ^pkg/client/ \
	    | grep -v v1alpha1/defaults.go); \
	  do \
	   golint --set_exit_status $$i || exit 1; \
	  done'
	@#
	go vet $(NON_VENDOR_DIRS)
	@echo Running repo-infra verify scripts
	@vendor/github.com/kubernetes/repo-infra/verify/verify-boilerplate.sh --rootdir=. | grep -v generated > .out 2>&1 || true
	@bash -c '[ "`cat .out`" == "" ] || (cat .out ; false)'
	@rm .out
	@#
	@echo Running href checker:
	@build/verify-links.sh
	@echo Running errexit checker:
	@build/verify-errexit.sh

verify-client-gen: .init .generate_files
	$(BUILD_DIR)/verify-client-gen.sh

format: .init
	gofmt -w -s $(TOP_SRC_DIRS)

test: .init build test-unit

test-unit: .init build
	@echo Running tests:
	go test $(UNIT_TEST_FLAGS) \
	  $(addprefix $(BROKER_PKG)/,$(TEST_DIRS))

clean: clean-bin clean-deps clean-build-image clean-generated clean-coverage

clean-bin:
	rm -rf $(BINDIR)
	rm -f .generate_exes

clean-deps:
	rm -f .init

# Building Docker Images for our executables
############################################
images: broker-image

broker-image: image/Dockerfile $(BINDIR)/broker build
	mkdir -p tmp
	cp $(BINDIR)/broker tmp
	cp image/Dockerfile tmp
	docker build -t brokersdk:$(VERSION) tmp
	docker tag brokersdk:$(VERSION) brokersdk:latest
	rm -rf tmp

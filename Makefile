COMMIT=$(shell git rev-parse --short HEAD)$(shell [[ $$(git status --porcelain --ignored) = "" ]] && echo -clean || echo -dirty)

# all is the default target to build everything
all: clean build sync e2e-bin logbridge customer-admin-controller

build: generate
	go build ./...

clean:
	rm -f azure-reader.log coverage.out end-user.log e2e.test logbridge sync customer-admin-controller

test: unit e2e

generate:
	go generate ./...

TAG ?= $(shell git rev-parse --short HEAD)
SYNC_IMAGE ?= quay.io/openshift-on-azure/sync:$(TAG)
LOGBRIDGE_IMAGE ?= quay.io/openshift-on-azure/logbridge:$(TAG)
E2E_IMAGE ?= quay.io/openshift-on-azure/e2e-tests:$(TAG)
CUSTOMER_ADMIN_CONTROLLER_IMAGE ?= quay.io/openshift-on-azure/customer-admin-controller:$(TAG)


logbridge: generate
	go build -ldflags "-X main.gitCommit=$(COMMIT)" ./cmd/logbridge

logbridge-image: logbridge
	go get github.com/openshift/imagebuilder/cmd/imagebuilder
	imagebuilder -f Dockerfile.logbridge -t $(LOGBRIDGE_IMAGE) .

logbridge-push: logbridge-image
	docker push $(LOGBRIDGE_IMAGE)

sync: generate
	go build -ldflags "-X main.gitCommit=$(COMMIT)" ./cmd/sync

sync-image: sync
	go get github.com/openshift/imagebuilder/cmd/imagebuilder
	imagebuilder -f Dockerfile.sync -t $(SYNC_IMAGE) .

sync-push: sync-image
	docker push $(SYNC_IMAGE)

customer-admin-controller: generate
	go build -ldflags "-X main.gitCommit=$(COMMIT)" ./cmd/customer-admin-controller

customer-admin-controller-image: customer-admin-controller
	go get github.com/openshift/imagebuilder/cmd/imagebuilder
	imagebuilder -f Dockerfile.customer-admin-controller -t $(CUSTOMER_ADMIN_CONTROLLER_IMAGE) .

customer-admin-controller-push: customer-admin-controller-image
	docker push $(CUSTOMER_ADMIN_CONTROLLER_IMAGE)

verify:
	./hack/validate-generated.sh
	go vet ./...
	./hack/verify-code-format.sh

unit: generate
	go test ./... -coverprofile=coverage.out
ifneq ($(ARTIFACT_DIR),)
	mkdir -p $(ARTIFACT_DIR)
	cp coverage.out $(ARTIFACT_DIR)
endif

cover: unit
	go tool cover -html=coverage.out

e2e: generate
	./hack/e2e.sh

e2e-prod:
	go test ./test/e2erp -tags e2erp -test.v -ginkgo.v -ginkgo.randomizeAllSpecs -ginkgo.noColor -ginkgo.focus=Real -timeout 4h

e2e-bin: generate
	go test -tags e2e -ldflags "-X github.com/openshift/openshift-azure/test/e2e.gitCommit=$(COMMIT)" -i -c -o e2e.test ./test/e2e

e2e-image: e2e-bin
	go get github.com/openshift/imagebuilder/cmd/imagebuilder
	imagebuilder -f Dockerfile.e2e -t $(E2E_IMAGE) .

e2e-push: e2e-image
	docker push $(E2E_IMAGE)

.PHONY: clean sync-image sync-push verify unit e2e e2e-bin e2e-prod

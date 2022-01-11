# Dirs
BUILDDIR		:= build
CMDDIR			:= cmd
TESTDIR			:= test

# Git
GIT_REPOSITORY		:= github.com/fydrah/loginapp
GIT_COMMIT_ID		:= $(shell git log -n 1 --pretty=format:%h)
GIT_TAG			:= $(shell git describe --tags --exclude="chart-*")

# Go
GOFLAGS			:=
LDFLAGS			= -w -s -X "$(GIT_REPOSITORY)/cmd.GitVersion=$(GIT_TAG)" -X "$(GIT_REPOSITORY)/cmd.GitHash=$(GIT_COMMIT_ID)" -extldflags "-static"
PACKR_VERSION		:= $(shell awk '/packr/ {print $$2}' go.mod)

# Docker
DOCKERFILE		:= Dockerfile
DOCKER_REPOSITORY	:= quay.io/fydrah/loginapp
DOCKER_BIN		:= $(shell which docker || which podman || echo "docker")
DOCKER_BUILD		:= $(DOCKER_BIN) build -f $(DOCKERFILE) .

.PHONY: all
all: go_build

.PHONY: go_packr2
go_packr2:
	@echo "[Go] install packr2"
	@which packr2 >/dev/null || go install github.com/gobuffalo/packr/v2/packr2@$(PACKR_VERSION)
	@# packr2 still requires GO111MODULE var
	@echo "[Go] run packr2 (embded assets)"
	@packr2 clean
	@GO111MODULE=on packr2

.PHONY: go_fmt
go_fmt:
	@echo "[Go] fmt"
	@go fmt ./...

.PHONY: go_mod_vendor
go_mod_vendor:
	@echo "[Go] vendor"
	@go mod vendor

.PHONY: go_build
go_build: go_mod_vendor go_packr2
	@echo "[Go] build"
	@CGO_ENABLED=0 go build -mod=vendor -o $(BUILDDIR)/loginapp $(GOFLAGS) -ldflags '$(LDFLAGS)' $(GIT_REPOSITORY)

.PHONY: docker_build
docker_build:
	@echo "[Docker] build image"
	@$(DOCKER_BUILD) -t $(DOCKER_REPOSITORY):$(GIT_COMMIT_ID)

.PHONY: docker_push
docker_push: docker_build
	@echo "[Docker] push image"
	@$(DOCKER_BIN) push $(DOCKER_REPOSITORY):$(GIT_COMMIT_ID)

.PHONY: helm_doc
helm_doc:
	@echo "[Helm] doc"
	@chart-doc-gen -d docs/chart.yaml -v=helm/loginapp/values.yaml > ./helm/loginapp/README.md

.PHONY: helm_package
helm_package: helm_doc
	@echo "[Helm] package chart"
	@helm package -u helm/loginapp -d $(BUILDDIR)

.PHONY: clean
clean:
	@echo "[Clean] binaries"
	@rm -f $(BUILDDIR)/loginapp
	@rm -f $(BUILDDIR)/loginapp-*.tgz
	@rm -f $(BUILDDIR)/index.yaml

.PHONY: clean_test
clean_test:
	@echo "[Clean] dev env"
	@rm -rf $(TESTDIR)/generated
	@rm -rf $(TESTDIR)/kubernetes/generated
	@rm -rf $(TESTDIR)/helm/generated

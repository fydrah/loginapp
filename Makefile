# Dirs
BINDIR			:= build
CMDDIR			:= cmd

# Git
GIT_REPOSITORY		:= github.com/fydrah/loginapp
GIT_COMMIT_ID		:= $(shell git log -n 1 --pretty=format:%h)
GIT_TAG				:= $(shell git describe --tags)

# Go
GOFLAGS			:=
LDFLAGS			= -w -s -X "$(GIT_REPOSITORY)/internal/app/loginapp.GitVersion=$(GIT_TAG)" -X "$(GIT_REPOSITORY)/internal/app/loginapp.GitHash=$(GIT_COMMIT_ID)"

# Docker
DOCKERFILE		:= Dockerfile
DOCKER_REPOSITORY	:= quay.io/fydrah/loginapp
DOCKER_BIN		:= $(shell which docker || which podman || echo "docker")
DOCKER_BUILD		:= $(DOCKER_BIN) build -f $(DOCKERFILE) .

.PHONY: all
all: build

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: build
build: vendor
	go build -mod=vendor -o $(BINDIR)/loginapp $(GOFLAGS) -ldflags '$(LDFLAGS)' $(GIT_REPOSITORY)/$(CMDDIR)/loginapp

.PHONY: build-static
build-static: LDFLAGS += -extldflags "-static"
build-static: vendor
	CGO_ENABLED=0 go build -mod=vendor -o $(BINDIR)/loginapp-static $(GOFLAGS) -ldflags '$(LDFLAGS)' $(GIT_REPOSITORY)/$(CMDDIR)/loginapp

.PHONY: docker-tmp
docker-tmp:
	$(DOCKER_BUILD) -t $(DOCKER_REPOSITORY):$(GIT_COMMIT_ID)

.PHONY: gofmt
gofmt:
	go fmt ./...

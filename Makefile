# Dirs
BINDIR				:= build/bin
CMDDIR				:= cmd

# Git
GIT_REPOSITORY		:= github.com/fydrah/loginapp
GIT_COMMIT_ID		:= $(shell git log -n 1 --pretty=format:%h)
GIT_TAG				:= $(shell git describe --tags)

# Go
GOFLAGS				:=
LDFLAGS				= -w -s -X main.GitVersion=$(GIT_TAG) -X main.GitHash=$(GIT_COMMIT_ID)

# Docker
DOCKERFILE			:= build/docker/Dockerfile
DOCKER_REPOSITORY	:= quay.io/fydrah/loginapp
DOCKER_BUILD		:= docker build -f build/docker/Dockerfile .

# Tests
CYCLO_MAX			:= 15

# Others
SRC_FILES			:= $(shell find . -name "*.go" -not -path "./vendor*")


.PHONY: all
all: build

.PHONY: build
build:
	go build -o $(BINDIR)/loginapp $(GOFLAGS) -ldflags '$(LDFLAGS)' $(GIT_REPOSITORY)/$(CMDDIR)/loginapp

.PHONY: build-static
build-static: LDFLAGS += -extldflags "-static"
build-static:
	CGO_ENABLED=0 go build -o $(BINDIR)/loginapp-static $(GOFLAGS) -ldflags '$(LDFLAGS)' $(GIT_REPOSITORY)/$(CMDDIR)/loginapp

.PHONY: docker-tmp
docker-tmp:
	$(DOCKER_BUILD) -t $(DOCKER_REPOSITORY):$(GIT_COMMIT_ID)

.PHONY: checks
checks: gocyclo staticcheck

.PHONY: gofmt
gofmt:
	go fmt ./...

.PHONY: gocyclo
gocyclo:
	@echo
	@echo "############ Run cyclomatic complexity check"
	which gocyclo || go get github.com/fzipp/gocyclo
	gocyclo -over $(CYCLO_MAX) $(SRC_FILES)

.PHONY: staticcheck
staticcheck:
	@echo
	@echo "############ Run simplifying code check (codes reference at https://staticcheck.io/docs/checks)"
	which staticcheck || go get honnef.co/go/tools/cmd/staticcheck
	staticcheck ./...

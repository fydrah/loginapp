BINDIR			:= bin
GOFLAGS			:=
DOCKER_REPOSITORY	:= quay.io/fydrah/loginapp
GIT_REPOSITORY		:= github.com/fydrah/loginapp
GIT_COMMIT_ID		:= $(shell git log -n 1 --pretty=format:%h)
GIT_TAG			:= $(shell git describe --tags)
DOCKERFILES		:= dockerfiles

LDFLAGS			= -w -s -X main.GitVersion=$(GIT_TAG) -X main.GitHash=$(GIT_COMMIT_ID)

.PHONY: all
all: build

.PHONY: build
build:
	go build -o $(BINDIR)/loginapp $(GOFLAGS) -ldflags '$(LDFLAGS)'

.PHONY: build-static
build-static: LDFLAGS += -extldflags "-static"
build-static:
	CGO_ENABLED=0 go build -o $(BINDIR)/loginapp-static $(GOFLAGS) -ldflags '$(LDFLAGS)'

.PHONY: docker-tmp
docker-tmp:
	docker build -t $(DOCKER_REPOSITORY):$(GIT_COMMIT_ID) .

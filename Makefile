BINDIR			:= bin
GOFLAGS			:=
DOCKER_REPOSITORY	:= devopyio/loginapp
GIT_REPOSITORY		:= github.com/devopyio/loginapp
GIT_COMMIT_ID		:= $(shell git log -n 1 --pretty=format:%h)
GIT_TAG			:= $(shell git describe --tags)
DOCKERFILES		:= dockerfiles
CYCLO_MAX		:= 15
SRC_FILES		:= $(shell find . -name "*.go" -not -path "./vendor*")

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

.PHONY: checks
checks: errcheck gocyclo gosimple

.PHONY: gofmt
gofmt:
	gofmt -w -s $(SRC_FILES)

.PHONY: errcheck
errcheck:
	@echo
	@echo "############ Run unchecked errors check"
	which errcheck || go get github.com/kisielk/errcheck
	errcheck $(GIT_REPOSITORY)

.PHONY: gocyclo
gocyclo:
	@echo
	@echo "############ Run cyclomatic complexity check"
	which gocyclo || go get github.com/fzipp/gocyclo
	gocyclo -over $(CYCLO_MAX) $(SRC_FILES)

.PHONY: gosimple
gosimple:
	@echo
	@echo "############ Run simplifying code check (codes reference at https://staticcheck.io/docs/gosimple)"
	which gosimple || go get honnef.co/go/tools/cmd/gosimple
	gosimple $(GIT_REPOSITORY)

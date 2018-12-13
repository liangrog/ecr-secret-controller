# Makefile for ecm build and cross-compiling

APPNAME=kctlr-docker-auth
IMAGE_NAME=liangrog/kctlr-docker-auth

DOCKERFILE=Dockerfile

VERSION_TAG=`git describe 2>/dev/null | cut -f 1 -d '-' 2>/dev/null`
COMMIT_HASH=`git rev-parse --short=8 HEAD 2>/dev/null`
BUILD_TIME=`date +%FT%T%z`
LDFLAGS=-ldflags "-s -w \
    -X main.CommitHash=${COMMIT_HASH} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.Tag=${VERSION_TAG}"

all: fast

test:
	GO111MODULE=on go test ./...
clean:
	go clean
	rm ./${APPNAME} || true

docker_binary:
	# Build static binary and disable cgo dependancy
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(APPNAME) ${LDFLAGS} .

docker_build:
	docker build . -f $(DOCKERFILE) -t $(IMAGE_NAME) --no-cache 

fast:
	GO111MODULE=on go build -o ${APPNAME} ${LDFLAGS}

build: all

docker: clean docker_binary docker_build


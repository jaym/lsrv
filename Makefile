GIT_SHA := $(shell git rev-parse --short HEAD)
GIT_REV := $(shell git rev-list --count HEAD)

default: 
	go build -ldflags "-X main.version=r.$(GIT_REV).$(GIT_SHA)" -o bin/lsrv cmd/lsrv.go


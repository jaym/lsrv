GIT_SHA := $(shell git rev-parse --short HEAD)
GIT_REV := $(shell git rev-list --count HEAD)

default: 
	# $(eval VERSION=$("r.$(shell git rev-list --count HEAD).$(shell git rev-parse --short HEAD)"))
	go build -ldflags "-X main.version=r.$(GIT_REV).$(GIT_SHA)" -o bin/lsrv cmd/lsrv.go


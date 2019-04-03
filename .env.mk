# Find project directory using make builtins
PROJECT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
NAME := github.com/gregzuro/service

# the project root is also the GOPATH
# workaround for https://github.com/golang/go/issues/14566
# to allow build self-contained within PWD with vendor/ dir
# should be unnecessary in Go 1.8
export GOPATH := $(PROJECT_DIR).gopath
BIN := $(GOPATH)/bin
PLATFORM := $(strip $(shell uname))

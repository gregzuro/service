Go: [![wercker status](https://app.wercker.com/status/47dd08d8524c45ff7f587d4b913a561b/s/master "wercker status")](https://app.wercker.com/project/byKey/47dd08d8524c45ff7f587d4b913a561b) Go+Android: [![Build Status](https://travis-ci.com/gregzuro/service.svg?token=K36oiyCbzU9cpzC1ffws&branch=master)](https://travis-ci.com/gregzuro/service)

# service
all Service Entities

## Setup

### Prerequisites

You'll need:

 1. Go 1.7.x https://golang.org/doc/install. Other Go versions should work, but 1.7 is the primary target.
 2. Git
 3. Docker
   - [Linux (plain docker engine)](https://docs.docker.com/engine/installation/linux/)
   - [Docker for Mac](https://docs.docker.com/docker-for-mac/)
   - [Docker for Windows](https://docs.docker.com/docker-for-windows/). Note: running the development environment directly on Windows isn't supported. Please use a Linux environment such as Alpine Linux or Ubuntu under a virtualization system like Virtualbox. TODO: specific instructions for setting up a windows dev env.

### Install dependencies

Once you have the bootstrap dependencies installed, you can get everything else by running

```
make
``` 

from the top level `service` directory. This will make:

 - `clean`
 - `vendor` (golang packages)
 - `bindeps`
   - `bin/go-bindata`, used to generate `.go` files which embed binary data.
   - `bin/protoc-gen-go`, `protoc-gen-grpc-gateway`, `protoc-gen-swagger`, `protoc` output plugins.
   - `bin/protoc`, a non-Go binary dependency, downloaded rather than built. Base `.proto` files are placed in `include/`.

## Run

Components are deployed and run, both in development and in production, as Docker containers. In development, there are several make targets which aid in starting and stopping containers. See [switch/README.md](./switch/README.md) for details.

## Develop

Code is stored in Git, as a single repository containing multiple projects (sometimes called a "monorepo").

### Go package dependencies

Dependencies are stored in your local `vendor/` directory. Run `glide install` to refresh dependencies to match the versions specified in `Glide.lock` after pulling down changes (or just run `make`).

Updating or adding a dependency is a bit more complicated. Glide updates *ALL* dependencies, rather than allowing you to single out just one. This creates some pain since we generally don't want to update all dependencies simultaneously (https://github.com/Masterminds/glide/issues/252). To work around it, you must perform the following steps:

 - `glide get somepackage`
 - Use `git add -i Glide.lock` (advanced) or a graphical equivalent like [Github Desktop](https://desktop.github.com/) (easy) to stage just the lock lines for `somepackage`.
 - `git checkout Glide.lock` to revert the other changes.
 
Updating a package version is similar, but starts with `glide update` instead.

### Shell environment

You may want to make the following changes to your shell/editor environment while working on the project:

```bash
export PATH=$PWD/bin:PATH
export GOPATH=$PWD/.gopath
```

This will allow you to use the local versions of various bin commands (`glide`, `protoc`, `gomobile`, etc) and ensure that the `go` command can find project packages.

### vendor vs $GOPATH

All Go source dependencies are installed into the `vendor/` directory at the top level of the project.

The project uses a private `$GOPATH` at `.gopath/`, which allows a self-contained build. The various Makefiles use this GOPATH automatically. More info at [.gopath/README.md](.gopath/README.md).

#### Branches

We have a few standard branches:

- master (production)
- staging
- test-\<test-cluster-name-suffix\>
- dev-yymmdd-\<description\>
- issueNNN-\<description\>

where \<test-cluster-name\> is the name suffix of the test cluster upon which the branch is meant to be deployed.
So a branch with the name `test-aces` is destined to be on the cluster names `test-aces`;

and where \<description\> is a description of the issue or ticket or story or whatever that the branch incorporates.

## Understand

### Go

The language that everything is written in.

https://golang.org/doc/install

### Glide

Vendors our dependencies.

https://github.com/Masterminds/glide

### Protobuf and Go Support

Protobuf is the compiler that ingests your `.proto` (~IDL) files and creates stubs for various languages:

https://github.com/google/protobuf

Go isn't supported out of the box, so you need to install the Go support plug-in as well:

https://github.com/golang/protobuf

### Go-gRPC gateway

The gateway is what allows us to have *both* gRPC endpoints *and* REST ones.
This component generates the REST ones from the `.proto` files (it's also a protobuf compiler plug-in):

https://github.com/grpc-ecosystem/grpc-gateway



 - https://golang.org/doc/
 - https://github.com/google/protobuf

## wercker

Wercker is our continuous integration service. It automatically builds any branches pushed to Github. The steps to execute a build are in `wercker.yaml`. You can [download](http://wercker.com/downloads/) the command line tool to run builds locally.

`wercker build` tries to connect to a local Docker instance via TCP by default. On Linux, Docker often [uses a Unix socket](http://devcenter.wercker.com/docs/faq/troubleshooting.html) instead. Try:

`wercker build --docker-host 'unix:///var/run/docker.sock'`

or set `DOCKER_HOST=unix:///var/run/docker.sock` in your shell.

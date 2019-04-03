include .env.mk

deps: vendor bindeps

clean: clean-bin clean-vendor

$(BIN)/glide:
	go get github.com/Masterminds/glide

glide.lock: glide.yaml
	$(BIN)/glide update

vendor: $(BIN)/glide glide.lock
  # vendor is also our $GOPATH/src
	$(BIN)/glide install

bindeps: $(BIN)/protoc-gen-go $(BIN)/protoc-gen-grpc-gateway $(BIN)/protoc-gen-swagger $(BIN)/go-bindata $(BIN)/protoc $(BIN)/gomobile

clean-bin:
	rm -f $(BIN)/*
	rm -rf ./include

clean-vendor:
	rm -rf vendor

VENDOR := $(NAME)/vendor
$(BIN)/protoc-gen-go:
	go install $(VENDOR)/github.com/golang/protobuf/protoc-gen-go

$(BIN)/protoc-gen-grpc-gateway:
	go install $(VENDOR)/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

$(BIN)/protoc-gen-swagger:
	go install $(VENDOR)/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

$(BIN)/go-bindata:
	go install $(VENDOR)/github.com/jteeuwen/go-bindata/go-bindata

$(BIN)/gomobile:
	go install $(VENDOR)/golang.org/x/mobile/cmd/gomobile


$(BIN)/protoc:
	version=3.1.0 ;\
	case "${PLATFORM}" in \
	Darwin) \
		platform=osx-x86_64 ;;\
	Linux) \
		platform=linux-x86_64 ;;\
	*) \
		echo "Unsupported platform ${PLATFORM}" ; exit 1 ;;\
	esac ;\
	TMP=$$(mktemp -d /tmp/tmp.XXXXX) ;\
	echo https://github.com/google/protobuf/releases/download/v$${version}/protoc-$${version}-$${platform}.zip ;\
	curl -L https://github.com/google/protobuf/releases/download/v$${version}/protoc-$${version}-$${platform}.zip > $${TMP}/protoc.zip ;\
	unzip $${TMP}/protoc.zip -d $${TMP} ;\
	mkdir ./include ;\
	cp $${TMP}/bin/protoc $(BIN)/ ;\
	cp -R $${TMP}/include/google ./include ;\


include ../../.env.mk

main := sgms.proto
out := .
protoc := $(BIN)/protoc

all: sgms.pb.go sgms.swagger.json sgms.pb.gw.go 

.pb.go: $(main) 

sgms.pb.go: .pb.go $(protoc) $(BIN)/protoc-gen-go 
	$(protoc) \
		--plugin=$(BIN)/protoc-gen-go \
		--go_out=\
Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,\
Mgoogle/protobuf/empty.proto=github.com/golang/protobuf/ptypes/empty,\
Mgoogle/protobuf/timestamp.proto=github.com/golang/protobuf/ptypes/timestamp,\
plugins=grpc:$(out) \
		sgms.proto

sgms.pb.gw.go: $(main) $(protoc) $(BIN)/protoc-gen-grpc-gateway
	$(protoc) \
		--plugin=$(BIN)/protoc-gen-grpc-gateway \
		--grpc-gateway_out=logtostderr=true:$(out) \
		$<

sgms.swagger.json: $(main) $(protoc) $(BIN)/protoc-gen-swagger scripts/includetxt.go
	$(protoc) \
		--plugin=$(BIN)/protoc-gen-swagger \
		--swagger_out=logtostderr=true:$(out) \
		$<
	go generate .
 
clean:
	-rm sgms.pb.go sgms.pb.gw.go sgms.swagger.json 

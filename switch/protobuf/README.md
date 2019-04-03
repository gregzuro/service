Here is a graph showing the swagger and protobuf dependencies for your edification:

![Swagger](http://g.gravizo.com/g?
  digraph G {
    "protoc[--go-out]"-> "sgms.pb.go" [label=creates];
    "protoc[--grpc-gateway-out]" -> "sgms.pb.gw.go" [label=creates];
    "protoc[--swagger-out]" -> "sgms.swagger.json" [label=creates];
    "protobuf/make" -> "protoc[--go-out]" [label=executes];
    "protobuf/make" -> "protoc[--grpc-gateway-out]" [label=executes];
    "protobuf/make" -> "protoc[--swagger-out]" [label=executes];
    "protobuf/make" -> "go generate" [label=executes];
    "scripts/includetxt.go" -> "go generate" [label="processed by"];
    "go generate" -> "swagger.pb.go" [label=creates];
    "hack/build-ui.sh" -> "go-bindata" [label=executes]
    "third_party/swagger-ui/..." -> "go-bindata" [label="processed by"]
    "go-bindata" -> "datafile.go" [label=creates]
    "sgms.swagger.json" -> "go generate" [label="processed-by"]
    "sgms.pb.go" -> "go build" [label="read by"]
    "sgms.pb.gw.go" -> "go build" [label="read by"]
    "swagger.pb.go" -> "go build" [label="read by"]
    "datafile.go" -> "go build" [label="read by"]
    })

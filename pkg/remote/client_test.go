package remote

import (
	"golang.org/x/net/context"
	// "reflect"
	// "sync"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/switch/protobuf"
	"google.golang.org/grpc"
)

// type SGMSServiceClient interface {
// 	SendSGMS(ctx context.Context, in *SGMS, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
// 	StreamSGMS(ctx context.Context, opts ...grpc.CallOption) (SGMSService_StreamSGMSClient, error)
// }

type mockSGMSServiceClient struct{}

func (c *mockSGMSServiceClient) SendSGMS(ctx context.Context, in *protobuf.SGMS, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, nil
}
func TestStreamClient(t *testing.T) {
	send := make(chan message.Message)
	down := ClientStreamer(mockSGMSServiceClient{}, send)

}

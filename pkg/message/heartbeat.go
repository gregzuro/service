package message

import (
	"time"

	"github.com/gregzuro/service/switch/protobuf"
	"golang.org/x/net/context"
)

func HeartBeat(ctx context.Context, entityId string) <-chan Message {
	out := make(chan Message)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		var c uint64
		for {
			select {
			case <-ticker.C:
				out <- Wrap(&protobuf.SGMS{
					HeartBeat: &protobuf.HeartBeat{
						Common: &protobuf.Common{
							EntityId: entityId,
						},
						Counter: c,
					},
				})
				c++
			case <-ctx.Done():
				break
			}
		}
	}()
	return out
}

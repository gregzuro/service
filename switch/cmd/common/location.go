package common

import (
	"time"

	google_protobuf1 "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/gregzuro/service/switch/protobuf"
	"golang.org/x/net/context"
)

// SendLocation sends a location message
func SendLocation(client protobuf.SGMSServiceClient, id string, _time time.Time, gpsLocation protobuf.Location) {
	var c protobuf.Common
	c.EntityId = id
	var ts google_protobuf1.Timestamp
	ts.Seconds = _time.Unix()
	ts.Nanos = int32(_time.Nanosecond())
	c.Timestamp = &ts
	l := gpsLocation
	l.Common = &c

	var msg protobuf.SGMS
	msg.Location = append(msg.Location, &l)

	client.SendSGMS(context.Background(), &msg)

}

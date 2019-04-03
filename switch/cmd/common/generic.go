package common

import (
	"time"

	google_protobuf1 "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/gregzuro/service/switch/protobuf"
	"golang.org/x/net/context"
)

// SendGeneric send a generic SGMS message
func SendGeneric(client protobuf.SGMSServiceClient, id string, _time time.Time, measurement string, tags map[string]string, fieldsInt64 map[string]int64, fieldsDouble map[string]float64) {
	var c protobuf.Common
	c.EntityId = id
	var ts google_protobuf1.Timestamp
	ts.Seconds = _time.Unix()
	ts.Nanos = int32(_time.Nanosecond())
	c.Timestamp = &ts
	var g protobuf.Generic
	g.Common = &c
	g.Measurement = measurement
	g.Tags = tags
	g.FieldsInt64 = fieldsInt64
	g.FieldsDouble = fieldsDouble

	var msg protobuf.SGMS
	msg.Generic = append(msg.Generic, &g)

	client.SendSGMS(context.Background(), &msg)

}

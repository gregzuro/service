package event

import (
	"github.com/Workiva/go-datastructures/trie/ctrie"
	google_protobuf1 "github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/gregzuro/service/plugin/cmd/locationeventspb"
	dt "github.com/gregzuro/service/plugin/cmd/pkg/decisiontree"
	"google.golang.org/grpc/grpclog"
)

/*	google_protobuf1 "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/Workiva/go-datastructures/trie/ctrie"
	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/pkg/pluginclient"
	pb "github.com/gregzuro/service/plugin/cmd/locationeventspb"
	dt "github.com/gregzuro/service/plugin/cmd/pkg/decisiontree"
	ev "github.com/gregzuro/service/plugin/cmd/pkg/event"
	"github.com/gregzuro/service/plugin/cmd/pkg/messageprocess"
	"github.com/gregzuro/service/plugin/cmd/pkg/queops"
	"github.com/tidwall/rtree"*/

type Event struct {
	EventId          string                      `protobuf:"bytes,1,opt,name=eventId" json:"eventId,omitempty"`
	ClientId         string                      `protobuf:"bytes,2,opt,name=clientId" json:"clientId,omitempty"`
	Name             string                      `protobuf:"bytes,3,opt,name=name" json:"name,omitempty"`
	Description      string                      `protobuf:"bytes,4,opt,name=description" json:"description,omitempty"`
	EndTimestamp     *google_protobuf1.Timestamp `protobuf:"bytes,5,opt,name=endTimestamp" json:"endTimestamp,omitempty"`
	StartTimestamp   *google_protobuf1.Timestamp `protobuf:"bytes,6,opt,name=startTimestamp" json:"startTimestamp,omitempty"`
	DecisionElements *[]dt.DecisionElement       `protobuf:"bytes,7,rep,name=decisionElements" json:"decisionElements,omitempty"`
	DeviceId         []string                    `protobuf:"bytes,8,rep,name=deviceId" json:"deviceId,omitempty"`
}

func StoreEventData(event pb.EventData, ect ctrie.Ctrie) {
	idt := dt.ConvertPBDecisionTreeToInternalDecisionTree(event.DecisionElements)
	dt.DumpDecisionTreeElements(idt)
	nev := Event{EventId: event.EventId, Name: event.Name, Description: event.Description,
		StartTimestamp: event.StartTimestamp, EndTimestamp: event.EndTimestamp,
		DecisionElements: &idt}
	_, ok := ect.Lookup([]byte(event.EventId))
	if ok {
		grpclog.Printf("Found Replace: %v", event.EventId)
		_, ok = ect.Remove([]byte(event.EventId))
		ect.Insert([]byte(event.EventId), nev)
	} else {
		grpclog.Printf("Insert: %v", event.EventId)
		ect.Insert([]byte(event.EventId), nev)
	}

}

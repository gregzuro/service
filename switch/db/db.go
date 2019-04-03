package db

import "github.com/gregzuro/service/switch/protobuf"

type TSDB interface {
	//	ShowDatabases() (string, error)
	SaveSGMS(*protobuf.SGMS)
	SaveError()
	SaveAudit()
	SaveSubscription()
	SaveLocation([]*protobuf.Location)
	SaveGeneric([]*protobuf.Generic)
}

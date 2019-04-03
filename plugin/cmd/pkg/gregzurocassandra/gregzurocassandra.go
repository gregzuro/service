package gregzurocassandra

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	pb "github.com/gregzuro/service/plugin/cmd/locationeventspb"
	"github.com/tidwall/rtree"

	"github.com/Workiva/go-datastructures/trie/ctrie"

	"github.com/gregzuro/service/plugin/cmd/pkg/deviceinfo"
	"github.com/gregzuro/service/plugin/cmd/pkg/event"
	"github.com/gregzuro/service/plugin/cmd/pkg/landmark"
)

//    "github.com/gregzuro/service/plugin/cmd/locationeventspb"
type DecisionElementType struct {
	DecisionElement []struct {
		Variable      []string          `json:"variables"`
		Parameters    map[string]string `json:"parameters"`
		Sequence      int32             `json:"sequence"`
		Level         int64             `json:"level"`
		Property      string            `json:"property"`
		Operation     string            `json:"operation"`
		Value         string            `json:"value"`
		Branchid      int64             `json:"branch_id"`
		Inbranch      int32             `json:"in_branch"`
		TrueBeElement struct {
			BranchName       string `json:"branch_name"`
			DtElementIndex   int32  `json:"dte_element_index"`
			FunctionCallName string `json:"function_call_name"`
		} `json:"true_be_element"`
		FalseBeElement struct {
			BranchName       string `json:"branch_name"`
			DteElementIndex  int32  `json:"dte_element_index"`
			FunctionCallName string `json:"function_call_name"`
		} `json:"false_be_element"`
	} `json:"decision_element"`
}

/*type WifiLandmarkDef struct {
	Mac       string `json:"mac"`
	Ssid      string `json:"ssid"`
	RssiLow   int32  `json:"rssilow"`
	RssiHigh  int32  `json:"rssihigh"`
	RadioType int32  `json:"radioType"`
}
type IbeaconLandmarkDef struct {
	RadioType int32  `json:"radioType"`
	Uuid      string `json:"Uuid"`
	Major     int32  `json:"major"`
	Minor     int32  `json:"minor"`
	Proximity int32  `json:"proximity"`
}*/

//type Polygon []polygon_points

type LandmarkUdt struct {
	PoiId   string   `json:"poi_id"`
	Events  []string `json:"events"`
	Polypts []struct {
		Points []struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"points"`
	} `json:"polygon_points"`
	IbeaconInfo      []pb.IBeaconLandmarkDef `json:"ibeacon_info"`
	WifiLandmarkInfo []pb.WifiLandmarkDef    `json:"wifi_info"`
}

//"polygon_points": [{"points":
var cluster *gocql.ClusterConfig

//var session *gocql.Session

func Initialize() *gocql.Session {
	// connect to the cluster
	cluster := gocql.NewCluster("ec2-54-89-159-28.compute-1.amazonaws.com")
	//cluster.Keyspace = "gregzuro"
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "root",
		Password: "gregzuro_1025!",
	}
	cluster.ProtoVersion = 4
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = time.Second * 2
	cluster.NumConns = 4

	session, _ := cluster.CreateSession()
	return session
}
func RetrieveDevices(dt *ctrie.Ctrie) {
	session := Initialize()
	var device_id string
	device_group_id := new([]string)
	events := new([]string)
	device_attribute := new([]string)

	iter := session.Query(`select device_id,device_attribute,device_group_id,events from gregzuro.device`).Iter()
	for iter.Scan(&device_id, device_attribute, device_group_id, events) {
		fmt.Println(device_id)
		dh, ok := dt.Lookup([]byte(device_id))
		var dhc *deviceinfo.DeviceHistory
		if !ok {
			dhc = new(deviceinfo.DeviceHistory)
			dt.Insert([]byte(device_id), dhc)
		} else {
			dhc = dh.(*deviceinfo.DeviceHistory)
		}
		dhc.DeviceGroupId = device_group_id
		dhc.Events = events
		dhc.DeviceAttributes = device_attribute
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
	session.Close()
}
func RetrieveEvents(ect *ctrie.Ctrie) {
	var eventId string
	var clientId string
	var decisionElementJSON string
	var description string

	var startTimestamp time.Time
	var endTimestamp time.Time
	var name string

	//var startTimestamp = new(timestamp.Timestamp)
	//var endTimestamp = new(timestamp.Timestamp)
	var eds []*pb.EventData
	session := Initialize()
	iter := session.Query(`select event_id,client_id,description,start_timestamp,end_timestamp,name from gregzuro.event_data`).Iter()

	//iter := session.Query(`select JSON decision_element from gregzuro.event_data`).Iter()

	for iter.Scan(&eventId, &clientId, &description, &startTimestamp, &endTimestamp, &name) {
		//for iter.Scan(&event_id, &client_id, &test_decision_element, &description, &end_timestamp, &name, &start_timestamp) {
		evt := new(pb.EventData)
		evt.EventId = eventId
		evt.ClientId = clientId
		evt.Description = description
		//evt.StartTimestamp = timestamp.Timestamp(&startTimestamp)
		//evt.EndTimestamp = timestamp.Timestamp(&endTimestamp)
		evt.Name = name
		eds = append(eds, evt)
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}

	for _, v := range eds {
		session.Close()
		session = Initialize()

		strQuery := `select JSON decision_element from gregzuro.event_data where event_id = '` + v.EventId + `'`
		fmt.Println(strQuery)
		var decisionElement DecisionElementType
		iter := session.Query(strQuery).Iter()
		// Copying is done since protobuf json names are generated and do not match cassandra database names
		for iter.Scan(&decisionElementJSON) {
			byt := []byte(decisionElementJSON)
			fmt.Println(decisionElementJSON)
			if err := json.Unmarshal(byt, &decisionElement); err != nil {
				log.Fatal(err)
			}
			fmt.Println(decisionElement)
			for i := 0; i < len(decisionElement.DecisionElement); i++ {
				v.DecisionElements = append(v.DecisionElements, new(pb.DecisionElement))
				v.DecisionElements[i].BranchID = decisionElement.DecisionElement[i].Branchid
				v.DecisionElements[i].Inbranch = decisionElement.DecisionElement[i].Inbranch
				v.DecisionElements[i].Level = decisionElement.DecisionElement[i].Level
				v.DecisionElements[i].Operation = decisionElement.DecisionElement[i].Operation
				v.DecisionElements[i].Property = decisionElement.DecisionElement[i].Property
				v.DecisionElements[i].Sequence = decisionElement.DecisionElement[i].Sequence
				v.DecisionElements[i].Value = decisionElement.DecisionElement[i].Value
				for _, vv := range decisionElement.DecisionElement[i].Variable {
					v.DecisionElements[i].Variable = append(v.DecisionElements[i].Variable, vv)
				}
				v.DecisionElements[i].Parameters = make(map[string]string)
				for key, value := range decisionElement.DecisionElement[i].Parameters {
					v.DecisionElements[i].Parameters[key] = value //fmt.Println("Key:", key, "Value:", value)
				}
				v.DecisionElements[i].TrueBElement = new(pb.BranchElement)
				v.DecisionElements[i].TrueBElement.BranchName = decisionElement.DecisionElement[i].TrueBeElement.BranchName
				v.DecisionElements[i].TrueBElement.DTElementIndex = decisionElement.DecisionElement[i].TrueBeElement.DtElementIndex
				v.DecisionElements[i].TrueBElement.FunctionCallName = decisionElement.DecisionElement[i].TrueBeElement.FunctionCallName
				v.DecisionElements[i].FalseBElement = new(pb.BranchElement)
				v.DecisionElements[i].FalseBElement.BranchName = decisionElement.DecisionElement[i].FalseBeElement.BranchName
				v.DecisionElements[i].FalseBElement.DTElementIndex = decisionElement.DecisionElement[i].FalseBeElement.DteElementIndex
				v.DecisionElements[i].FalseBElement.FunctionCallName = decisionElement.DecisionElement[i].FalseBeElement.FunctionCallName

			}

		}
		if err := iter.Close(); err != nil {
			log.Fatal(err)
		}

		event.StoreEventData(*v, *ect)

	}
	session.Close()

}
func RetrieveLandmarks(rt *rtree.RTree) {
	var jsonElement string
	var lmUdt LandmarkUdt
	session := Initialize()

	polyIter := session.Query(`select JSON poi_id, events, polygon_points, ibeacon_info, wifi_info from gregzuro.landmark_data`).Iter()
	for polyIter.Scan(&jsonElement) {
		var lm = new(pb.LandmarkData)
		fmt.Println(jsonElement)
		byt := []byte(jsonElement)

		if err := json.Unmarshal(byt, &lmUdt); err != nil {
			log.Fatal(err)
		}
		fmt.Println(lmUdt)
		lm.PoiId = lmUdt.PoiId
		lm.Events = nil
		for i := 0; i < len(lmUdt.Events); i++ {
			ee := new(pb.EventElement)
			ee.EventID = lmUdt.Events[i]
			lm.Events = append(lm.Events, ee)
		}

		lm.IbeaconInfo = nil
		for i := 0; i < len(lmUdt.IbeaconInfo); i++ {
			lm.IbeaconInfo = append(lm.IbeaconInfo, &lmUdt.IbeaconInfo[i])
		}
		lm.WifiInfo = nil
		for i := 0; i < len(lmUdt.WifiLandmarkInfo); i++ {
			lm.WifiInfo = append(lm.WifiInfo, &lmUdt.WifiLandmarkInfo[i])
		}
		lm.PolygonPoints = nil
		for i := 0; i < len(lmUdt.Polypts); i++ {
			//lmUdt.Polypts
			newPolyGon := new(pb.Polygon)
			for ii := 0; ii < len(lmUdt.Polypts[i].Points); ii++ {
				newPoint := new(pb.Point)
				newPoint.X = lmUdt.Polypts[i].Points[ii].X
				newPoint.Y = lmUdt.Polypts[i].Points[ii].Y
				newPolyGon.PolygonPoints = append(newPolyGon.PolygonPoints, newPoint)
				//				lm.PolygonPoints[i].PolygonPoints[ii].X = lmUdt.Polypts[i].Points[ii].X
				//				lm.PolygonPoints[i].PolygonPoints[ii].Y = lmUdt.Polypts[i].Points[ii].Y
			}
			lm.PolygonPoints = append(lm.PolygonPoints, newPolyGon)
		}

		landmark.StoreLandmarkData(lm, *rt)

	}
	if err := polyIter.Close(); err != nil {
		log.Fatal(err)
	}
	session.Close()

}

//var polygon polygon_points

/*	deStrQuery := "select JSON decision_element from gregzuro.event_data where event_id='" + eventId + "'"
	var decisionElement DecisionElementType
	iter = session.Query(deStrQuery).Iter()

	for iter.Scan(&decisionElementJSON) {
		//	for iter.Scan(&event_id, &client_id, &decision_element, &description, &end_timestamp, &name, &start_timestamp) {
		byt := []byte(decisionElementJSON)

		if err := json.Unmarshal(byt, &decisionElement); err != nil {
			log.Fatal(err)
		}
		fmt.Println(decisionElement)

	}*/

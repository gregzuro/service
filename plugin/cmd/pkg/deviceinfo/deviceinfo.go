package deviceinfo

import (
	"errors"
	"time"

	"github.com/Workiva/go-datastructures/trie/ctrie"
	pb "github.com/gregzuro/service/plugin/cmd/locationeventspb"
)

//	cass "github.com/gregzuro/service/plugin/cmd/pkg/cassandra"

var ct *ctrie.Ctrie
var numLocationsToKeep int32 = 5

type DeviceStateType int8

const (
	InArea DeviceStateType = iota + 1
	OutOfArea
	EnterArea
	ExitArea
	InLandmark
	OutOfLandmark
	Unknown
)

type DeviceStateByEvent struct {
	DeviceState DeviceStateType
	Event       string
}
type DeviceHistoryElement struct {
	DeviceLocation      pb.LocationData
	TimeStamp           time.Time
	InAreaLandMarksList []string
	DeviceStateBasedOn  []string
	LastSequence        int64
	DeviceID            string
	DeviceState         []DeviceStateByEvent
}
type DeviceHistory struct {
	DevHistory       []DeviceHistoryElement
	NumLastLocations int32
	CurrentIndex     int
	DeviceGroupId    *[]string
	Events           *[]string
	DeviceAttributes *[]string
}

func Initialize(devt *ctrie.Ctrie) {
	ct = devt
}
func SlideingWindowGPSInArea(device string) float64 {
	//dl, err := GetDeviceLocations(device)
	//if err == nil {

	//}
	return 0.0

}
func AddLandmarkstoDeviceState(Landmark string, device string, event string, state DeviceStateType) error {

	dh, ok := ct.Lookup([]byte(device))

	if ok {
		dhc := dh.(*DeviceHistory)
		dhe := dhc.DevHistory[dhc.CurrentIndex]
		dhe.InAreaLandMarksList = append(dhe.InAreaLandMarksList, Landmark)
		dtbe := new(DeviceStateByEvent)
		dtbe.Event = event
		dtbe.DeviceState = state
		dhe.DeviceState = append(dhe.DeviceState, *dtbe)
	} else {
		return errors.New("Device Not Found")
	}

	return nil
}
func AddDeviceState(loc *pb.LocationData, device string) error {
	//t := time.Now()
	dh, ok := ct.Lookup([]byte(device))
	var dhc *DeviceHistory
	if !ok {
		dhc = new(DeviceHistory)
		ct.Insert([]byte(device), dhc)
	} else {
		dhc = dh.(*DeviceHistory)
	}

	dhe := new(DeviceHistoryElement)
	dhe.DeviceLocation = *loc

	dhe.DeviceID = device

	dhc.NumLastLocations++
	dhc.DevHistory = append(dhc.DevHistory, *dhe)
	if dhc.NumLastLocations > numLocationsToKeep {
		//write to database
		dhc.DevHistory = dhc.DevHistory[1:]
		dhc.CurrentIndex = len(dhc.DevHistory)
	}

	return nil
}
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func GetDeviceLocations(device string) (*DeviceHistory, error) {

	dh, ok := ct.Lookup([]byte(device))

	if ok {
		dhc := dh.(*DeviceHistory)
		return dhc, nil
	}
	return nil, errors.New("No Location Information")
}
func GetDeviceHistory(device string) (*DeviceHistory, error) {

	dh, ok := ct.Lookup([]byte(device))

	if ok {
		dhc := dh.(*DeviceHistory)
		return dhc, nil
	}
	return nil, errors.New("No Location Information")
}

/*func CalcualeDeviceStateByEvent(device string, event string) error {
	dh, ok := ct.Lookup([]byte(device))

	if ok {
		dhc := dh.(*DeviceHistory)
		for i := 0; i < len(dhc.DevHistory); i++{
			var lastState  DeviceStateType = Unknown
			for ii := 0; len(dhc.DevHistory[i].DeviceState); ii++{
				ds := dhc.DevHistory[i].DeviceState[ii]
				if event == ds.Event {

	              if lastState == InArea && ds.DeviceState = InLandmark{
					  ds.DeviceState = InArea
					  lastState = InArea
				  }

				}
 			}
		}
		return dhc, nil
	}
	return nil, errors.New("No Location Information")

}

*/
/*func GetDeviceState(device string) (int8, error) {

	dh, ok := ct.Lookup([]byte(device))
	dhc := dh.(DeviceHistory)
	if ok {
		return dhc.DeviceState, nil
	}
	return 0, errors.New("No Location Information")
}*/

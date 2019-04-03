package messageprocess

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/Workiva/go-datastructures/trie/ctrie"
	pb "github.com/gregzuro/service/plugin/cmd/locationeventspb"
	dt "github.com/gregzuro/service/plugin/cmd/pkg/decisiontree"
	"github.com/gregzuro/service/plugin/cmd/pkg/deviceinfo"
	evp "github.com/gregzuro/service/plugin/cmd/pkg/event"
	"github.com/gregzuro/service/plugin/cmd/pkg/queops"
	pip "github.com/pip-go"
	"github.com/tidwall/rtree"
)

//	"github.com/pip-go"
//   "github.com/gregzuro/service/plugin/cmd/pkg/queops"
// "github.com/Workiva/go-datastructures/queue"

var log = logrus.New()

func init() {
	log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter) // default
	log.Level = logrus.DebugLevel

}

type TRect struct {
	RectPoints []float64
	LandMark   *pb.LandmarkData
}

func (r *TRect) Arr() []float64 {
	return []float64(r.RectPoints)
}
func (r *TRect) Rect(ctx interface{}) (min, max []float64) {
	return r.Arr()[:len(r.Arr())/2], r.Arr()[len(r.Arr())/2:]
}

func (r *TRect) String() string {
	min, max := r.Rect(nil)
	return fmt.Sprintf("%v,%v", min, max)
}

func LocationUpdateThread(lq *queue.PriorityQueue, rt *rtree.RTree, ect *ctrie.Ctrie, devt *ctrie.Ctrie) {

	continueProcessing := true
	//deviceinfo.Initialize
	deviceinfo.Initialize(devt)
	log.WithFields(logrus.Fields{
		"Operation": "Start Priority Queue",
		"Function":  "LocationUpdateThread",
	}).Debug("Start")

	for continueProcessing == true {
		item, err := queops.ReadFromPriorityQue(lq)
		log.WithFields(logrus.Fields{
			"Operation": "Received Message",
			"Function":  "LocationUpdateThread",
		}).Debug("Process")

		if err == nil {
			var actionableEvents []string
			locationInt := *item.StructPtr
			location := locationInt.(*pb.LocationData)
			//deviceinfo.AddDeviceLocation(location, location.Descriptor)
			var rectPts []float64

			rectPts = append(rectPts, location.Latitude)
			rectPts = append(rectPts, location.Longitude) // x,y
			rectPts = append(rectPts, location.Latitude)
			rectPts = append(rectPts, location.Longitude)

			tr := TRect{
				LandMark:   nil,
				RectPoints: rectPts,
			}

			gc := new(dt.GContext)
			deviceinfo.AddDeviceState(location, location.DeviceID)
			dc, _ := deviceinfo.GetDeviceLocations(location.DeviceID)
			gc.DeviceHistory = dc
			gc.Device = location.DeviceID
			// First searh for devices that are in an area
			rt.Search(&tr, func(item rtree.Item) bool {
				landmarkRect := item.(*TRect)
				log.WithFields(logrus.Fields{
					"Operation": "Search",
					"Rectangle": landmarkRect.LandMark.PoiId,
					"Function":  "LocationUpdateThread",
				}).Debug("Process")

				point := pip.Point{X: location.Latitude, Y: location.Longitude}
				var landmarkPolygon = new(pip.Polygon)
				var isInPoly = false
				for _, v := range landmarkRect.LandMark.PolygonPoints {
					for _, vi := range v.PolygonPoints {
						var landmarkPoint = new(pip.Point)
						landmarkPoint.X = vi.X
						landmarkPoint.Y = vi.Y
						landmarkPolygon.Points = append(landmarkPolygon.Points, *landmarkPoint)

					}
					if pip.PointInPolygon(point, *landmarkPolygon) {
						isInPoly = true
						log.WithFields(logrus.Fields{
							"Operation": "Search",
							"Rectangle": "In Polygon",
							"Function":  "LocationUpdateThread",
						}).Debug("Process")
					}

				}
				if isInPoly {
					for _, v := range landmarkRect.LandMark.Events {
						//todo is device part of event
						actionableEvents = append(actionableEvents, v.EventID)
						_, eventExists := ect.Lookup([]byte(v.EventID))

						if eventExists {

							//ev := evr.(evp.Event)

							deviceinfo.AddLandmarkstoDeviceState(landmarkRect.LandMark.PoiId, gc.Device, v.EventID, deviceinfo.InLandmark)

						}
					}
				}

				return true
			})
			//todo add blacklist of events that have been processed
			dh, err := deviceinfo.GetDeviceHistory(gc.Device)
			if err == nil {

				for _, evid := range *(dh.Events) {

					evr, eventExists := ect.Lookup([]byte(evid))
					if eventExists {
						ev := evr.(evp.Event)
						log.WithFields(logrus.Fields{
							"Operation": "Search",
							"Event":     ev.EventId,
							"Function":  "EvaluateTree",
						}).Debug("Process")

						if deviceinfo.StringInSlice(evid, actionableEvents) {
							deviceinfo.AddLandmarkstoDeviceState("", gc.Device, evid, deviceinfo.OutOfLandmark)
						}

						dt.DumpDecisionTreeElements(*ev.DecisionElements)
						dt.EvaluateTree(*ev.DecisionElements, *gc)
					}

				}
			}

		}

	}

}

/*	TODO add context history				if exists {
		var gc dt.GContext

		gcr, devExists := devt.Lookup([]byte(location.DeviceID))

		  if !devexists {
			grpclog.Printf("Insert: %v", ev.EventId)
			gc = new(decisiontree.GContext)
			gc.Device = location.DeviceID
			devt.Insert([]byte(location.DeviceID), gc)
		}

	}
*/

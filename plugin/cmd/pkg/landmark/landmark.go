package landmark

import (
	pb "github.com/gregzuro/service/plugin/cmd/locationeventspb"
	"github.com/gregzuro/service/plugin/cmd/pkg/messageprocess"
	"github.com/tidwall/rtree"
)

func getlatLongMBR(pp []*pb.Point) []float64 {

	var minX float64 = 90.0
	var minY float64 = 180.0
	var maxX float64 = -90.0
	var maxY float64 = -180.0
	var pointArray []float64

	for _, pv := range pp {
		if pv.X < minX {
			minX = pv.X
		}
		if pv.X > maxX {
			maxX = pv.X
		}
		if pv.Y < minY {
			minY = pv.Y
		}
		if pv.Y > maxY {
			maxY = pv.Y
		}
	}
	pointArray = append(pointArray, minX)
	pointArray = append(pointArray, minY)
	pointArray = append(pointArray, maxX)
	pointArray = append(pointArray, maxY)
	return pointArray
}

func StoreLandmarkData(landmark *pb.LandmarkData, rt rtree.RTree) {
	for _, v := range landmark.PolygonPoints {
		rectPts := getlatLongMBR(v.PolygonPoints)
		var found = false
		tr := messageprocess.TRect{
			LandMark:   landmark,
			RectPoints: rectPts,
		}
		rt.Search(&tr, func(item rtree.Item) bool {
			landmarkRect := item.(*messageprocess.TRect)
			if landmarkRect.LandMark.PoiId == landmark.PoiId {
				landmarkRect.LandMark = landmark
				found = true
			}
			return true
		})

		if !found {
			rt.Insert(&tr)
		}
	}

}

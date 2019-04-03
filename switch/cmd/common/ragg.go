package common

import "github.com/gregzuro/service/switch/protobuf"

// countryHandler takes a lat & lng query params and returns a string
// with the 'country' of the coordinate
func (s *RAGGServer) countryHandler(l protobuf.Location) string {

	data := s.Query(float64(l.Latitude), float64(l.Longitude))
	if len(data) == 0 {
		return "nowhere!"
	}

	return "in " + data["iso_a2"] + "/" + data["name"]
}

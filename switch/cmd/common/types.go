package common

import (
	"fmt"

	ragg "github.com/akhenakh/regionagogo"
	"github.com/gregzuro/service/switch/protobuf"
)

// PolygonCL sets() multiple) GA GF(s) from command line
type PolygonCL []string

// String is for CL arg mgmt
func (i *PolygonCL) String() string {
	return "!a GeoAffinityCL!"
}

// Set is for CL arg mgmt
func (i *PolygonCL) Set(value string) error {
	fmt.Println("got polygon: ", value)
	*i = append(*i, value)
	return nil
}

// Type is for CL arg mgmt
func (i *PolygonCL) Type() string {
	return "string"
}

// GeoAffinityCL defines geo affinities passed on the command line
type GeoAffinityCL struct {
	Include PolygonCL
	Exclude PolygonCL
}

// CommandArgs keeps values for all the arguements that were passed on the command line
type CommandArgs struct {
	MasterAddress  string
	MasterPort     int64
	ParentAddress  string
	ParentPort     int64
	ListenPort     int64
	ShortName      string // set by the `--id` command-line argument
	StreamMessages bool

	GeoAffinityCL          GeoAffinityCL `json:"geoaffinity_cl,omitempty"`
	ConfigFile             string
	SlaveGeoAffinitiesFile string `json:"slave_geoaffinities_file,omitempty"`

	PortMappedCluster bool

	NumTestDevices int
}

// RPCClients stores the client handles for the various endpoints
type RPCClients struct {
	SGMS    protobuf.SGMSServiceClient
	Status  protobuf.StatusServiceClient
	Contact protobuf.InitialContactServiceClient
}

// RAGGServer is for PIP kludge
type RAGGServer struct {
	*ragg.GeoSearch
}

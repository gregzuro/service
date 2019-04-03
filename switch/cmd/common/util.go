package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gregzuro/service/switch/protobuf"
	geojson "github.com/paulmach/go.geojson"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	google_protobuf1 "github.com/golang/protobuf/ptypes/timestamp"
)

const (
	port = 7467
)

// AddCommFlags adds flags related to network communication
func AddCommFlags(Cmd *cobra.Command, commandArgs *CommandArgs) {
	addListenPortFlag(Cmd, commandArgs)
	AddMasterFlags(Cmd, commandArgs)
	addIDFlag(Cmd, commandArgs)
	Cmd.Flags().BoolVarP(&commandArgs.StreamMessages, "stream", "S", false, "make persistent streaming connections to send messages")
}

// AddGeoAffinityFlags add flags for specifying include and exclude geoaffinity
func AddGeoAffinityFlags(Cmd *cobra.Command, commandArgs *CommandArgs) {
	Cmd.Flags().Var(&commandArgs.GeoAffinityCL.Include, "iga", "a polygon defining (an included) geofence using lat,lng{;lat,lng}...")
	Cmd.Flags().Var(&commandArgs.GeoAffinityCL.Exclude, "ega", "a polygon defining (an excluded) geofence using lat,lng{;lat,lng}...")
}

func addIDFlag(Cmd *cobra.Command, commandArgs *CommandArgs) {
	Cmd.Flags().StringVarP(&commandArgs.ShortName, "short-name", "s", "none", "a short name for this switch")
}

func addListenPortFlag(Cmd *cobra.Command, commandArgs *CommandArgs) {
	Cmd.Flags().Int64VarP(&commandArgs.ListenPort, "listenport", "l", port, "the port upon which this node should listen")
}

// AddMasterFlags adds flags for this entity's master info
func AddMasterFlags(Cmd *cobra.Command, commandArgs *CommandArgs) {
	addMasterAddressFlag(Cmd, commandArgs)
	addMasterPortFlag(Cmd, commandArgs)
}

func addMasterAddressFlag(Cmd *cobra.Command, commandArgs *CommandArgs) {
	Cmd.Flags().StringVarP(&commandArgs.MasterAddress, "master-address", "a", "", "the master switch ip address to which this switch should (initially) connect")
}

func addMasterPortFlag(Cmd *cobra.Command, commandArgs *CommandArgs) {
	Cmd.Flags().Int64VarP(&commandArgs.MasterPort, "master-port", "p", port, "the master switch ip port to which this switch should (initially) connect")
}

// ReadConfigFile reads a config file
func ReadConfigFile(cfgFile string, Cmd *cobra.Command, commandArgs *CommandArgs) error {

	viper.SetDefault("DockerAddress", "unix:///var/run/docker.sock")
	viper.SetDefault("influxdb-host", "http://influxdb_master1:8086")
	viper.SetDefault("influxdb-dbname", "sgms")
	viper.SetDefault("influxdb-user", "sgms")
	viper.SetDefault("log-level", "info")

	if cfgFile == "" { // if a config file was not found, then construct a name
		deployEnv := os.Getenv("DEPLOY_ENV")
		if deployEnv == "" {
			viper.SetDefault("influxdb-password", "dev") // NB: this will fail when not in dev unless password is specified elsewhere
			commandArgs.ConfigFile = "./configs/" + Cmd.Use + ".yaml"
		} else {
			viper.SetDefault("influxdb-password", deployEnv) // NB: this will likely fail unless password is specified elsewhere
			commandArgs.ConfigFile = "./configs/" + Cmd.Use + "/" + Cmd.Use + "-" + deployEnv + ".yaml"
		}
	} else {
		// If a config file was specified, then use it
		commandArgs.ConfigFile = cfgFile
	}
	viper.SetConfigFile(commandArgs.ConfigFile)

	c := viper.ReadInConfig()

	// config file may have specified a log level..
	l, _ := log.ParseLevel(viper.Get("log-level").(string))
	log.SetLevel(l)

	return c
}

// ToGoTime converts protobuf timestamps to Go times
func ToGoTime(pt *google_protobuf1.Timestamp) time.Time {
	return time.Unix(pt.Seconds, int64(pt.Nanos))
}

// ToPbTime converts Go times to protobuf timestamps
func ToPbTime(gt time.Time) *google_protobuf1.Timestamp {
	return &google_protobuf1.Timestamp{Seconds: int64(gt.Unix()), Nanos: int32(gt.Nanosecond())}
}

type ll struct {
	lat  float64
	long float64
}

// ParseJSONGeoAffinities converts a geojson featureCollection into our (interim?) GA format
func ParseJSONGeoAffinities(ga *[]*protobuf.GeoAffinity, staticGeoAffinities []byte) (int, error) {

	wantCovering := 0

	featureCollection, err := geojson.UnmarshalFeatureCollection(staticGeoAffinities)
	if err != nil {
		return wantCovering, fmt.Errorf("error unmarshalling JsonGeoAffinities")
	}

	for _, feature := range featureCollection.Features {
		if feature.Geometry.IsPolygon() {
			polygon := protobuf.Polygon{}
			for _, point := range feature.Geometry.Polygon[0] {
				polygon.Points = append(polygon.Points, &protobuf.Point{Latitude: point[1], Longitude: point[0]})
			}

			geofence := protobuf.GeoFence{Polygon: &polygon}
			geoAffinity := protobuf.GeoAffinity{
				GeoFence:     &geofence,
				Exclude:      false,
				WantCovering: 1,
				Name:         feature.Properties["regionName"].(string),
			}
			*ga = append(*ga, &geoAffinity)
			wantCovering += int(geoAffinity.WantCovering)
		} else {
			log.WithFields(log.Fields{
				"context": "Unmarshal()"}).Infof("non-polygon found in staticGeoAffinities: (regionName: %v, Geometry: %v)",
				feature.Properties["regionName"], feature.Geometry.Type)
		}
	}
	fmt.Println(featureCollection.Features[0].Properties["regionName"], featureCollection.Features[0].Geometry)

	return wantCovering, nil
}

// ParseCLGeoAffinities splits up the command line's string versions of the GAs into proper ones for Entity
func (gacl *GeoAffinityCL) ParseCLGeoAffinities(ga *[]*protobuf.GeoAffinity) (int, error) {
	// wantCovering, err := gacl.parseCLGeoAffinities(ga, false)
	// if err != nil {
	// 	return wantCovering, err
	// }
	return gacl.parseCLGeoAffinities(ga, false)

}

// parseCLGeoAffinities splits up the command line's string versions of the GAs into proper ones for Entity for either include or exclude
func (gacl *GeoAffinityCL) parseCLGeoAffinities(ga *[]*protobuf.GeoAffinity, exclude bool) (int, error) {

	var polygonF64 [][]float64

	var fooClude PolygonCL
	if exclude {
		fooClude = gacl.Exclude
	} else {
		fooClude = gacl.Include
	}

	wantCovering := 0
	for _, fooCludeV := range fooClude {
		err := json.Unmarshal([]byte(fooCludeV), &polygonF64)
		if err != nil {
			log.WithFields(log.Fields{
				"context": "Unmarshal()"}).Error(err)
			return wantCovering, err
		}

		var p protobuf.Polygon
		for _, pointV := range polygonF64 {
			p.Points = append(p.Points, &protobuf.Point{Latitude: pointV[1], Longitude: pointV[0]})
		}

		// TODO(greg) simplify creation of these intermediate items
		gf := protobuf.GeoFence{Polygon: &p}

		err = ValidatePolygon(gf.Polygon)
		if err != nil {
			log.WithFields(log.Fields{
				"context": "ValidatePolygon()"}).Error(err)
			return wantCovering, err
		}

		// TODO(greg) pull want_covering from command line
		newGa := protobuf.GeoAffinity{WantCovering: 1, Exclude: exclude, GeoFence: &gf}
		*ga = append(*ga, &newGa)
		wantCovering += int(newGa.WantCovering)
	}

	return wantCovering, nil
}

// ValidatePolygon makes sure that your polygon is kosher
func ValidatePolygon(p *protobuf.Polygon) error {

	if len(p.Points) < 3 {
		return errors.New("bad polygon - less than three 3 vertices")
	}
	return nil
}

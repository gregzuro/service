package db

import (
	"fmt"
	"time"

	"github.com/gregzuro/service/switch/protobuf"
	"github.com/influxdata/influxdb/client/v2"
)

const (
	MINIMUM_WRITE_INTERVAL_IN_MS = 1000
	MAXIMUM_POINTS_TO_CACHE      = 1000
)

type InfluxTSDB struct {
	Host         string
	DBName       string
	Config       client.HTTPConfig
	Client       client.Client
	BPC          client.BatchPointsConfig
	BP           client.BatchPoints
	WriteChannel chan *client.Point
}

// NewInfluxTSDB gets and returns a client connection to the database
func NewInfluxTSDB(host, dbName, user, password string) (*InfluxTSDB, error) {

	// TODO(greg) get dbname, user, password, etc. from env RIGHT HERE

	var influxTSDB InfluxTSDB
	influxTSDB.DBName = dbName

	influxTSDB.Config = client.HTTPConfig{
		Addr:     host,
		Username: user,
		Password: password,
	}

	var err error
	influxTSDB.Client, err = client.NewHTTPClient(influxTSDB.Config)
	if err != nil {
		return &influxTSDB, err
	}

	// create a batch for writing
	influxTSDB.BPC = client.BatchPointsConfig{
		Database:        influxTSDB.DBName,
		RetentionPolicy: "autogen",
	}
	influxTSDB.BP, err = client.NewBatchPoints(influxTSDB.BPC)
	if err != nil {
		return &influxTSDB, err
	}

	// create a channel to transmit the points that we want to save
	influxTSDB.WriteChannel = make(chan *client.Point, 100)

	// kick off the routine that saves points peridically
	go influxTSDB.savePoints()

	return &influxTSDB, nil
}

// ShowDatabases does that (for testing)
func (t *InfluxTSDB) ShowDatabases() (string, error) {
	q := client.Query{Command: `show databases`}
	response, err := t.Client.Query(q)
	if err != nil {
		return "not okay!", err
	}
	for _, v := range response.Results {
		for _, v := range v.Series {
			fmt.Printf("\t")
			for _, v := range v.Columns {
				fmt.Printf("\t\t%v\n", v)
			}
			for k, v := range v.Values {
				fmt.Printf("\t\t%04v:\t%v\n", k, v)
			}
		}
	}

	//	fmt.Printf("results: %v\n", response.Results)
	return "okay!", err
}

func (t *InfluxTSDB) SaveGeneric(generic []*protobuf.Generic) {

	// send the message to be enqueued
	// fmt.Println(generic[0].Common.EntityId)
	for _, v := range generic {

		c := v.Common
		fields := make(map[string]interface{})
		for k, v := range v.FieldsInt64 {
			fields[k] = v
		}
		for k, v := range v.FieldsDouble {
			fields[k] = v
		}
		err := t.enqueueSavePoint(c.EntityId, v.Measurement, v.Tags, fields)
		if err != nil {
			fmt.Println("error in enqueueSavePoint: " + err.Error())
		}
	}
}

func (t *InfluxTSDB) SaveLocation(gpsLocation []*protobuf.Location) {
	// send the message to be enqueued
	// fmt.Println(generic[0].Common.EntityId)
	for _, v := range gpsLocation {

		c := v.Common
		fields := make(map[string]interface{})

		fields["latitude"] = v.Latitude
		fields["longitude"] = v.Longitude
		fields["velocity"] = v.Velocity
		fields["course"] = v.Course

		err := t.enqueueSavePoint(c.EntityId, "location", map[string]string{"EntityId": v.Common.EntityId}, fields)
		if err != nil {
			fmt.Println("error in enqueueSavePoint: " + err.Error())
		}
	}
}

// enqueueSavePoint builds a point and sends it to the queue
func (t *InfluxTSDB) enqueueSavePoint(entityId string, measurement string, tags map[string]string, fields map[string]interface{}) error {

	// make a proper point

	point, err := client.NewPoint(
		measurement,
		tags,
		fields,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}

	// send to saver via channel
	t.WriteChannel <- point

	return nil
}

// savePoints queues and saves points to the database periodically
func (t *InfluxTSDB) savePoints() {

	lastWriteAttemptTime := time.Now().UTC()
	waitUntil := lastWriteAttemptTime.Add(MINIMUM_WRITE_INTERVAL_IN_MS * time.Millisecond)
	for {
		// calculate how much (more time to wait)
		timeToWait := waitUntil.Sub(time.Now().UTC())

		select {
		case point := <-t.WriteChannel:
			//fmt.Printf("got point: %v\n", point)

			// add point to batch
			t.BP.AddPoint(point)

			if len(t.BP.Points()) >= MAXIMUM_POINTS_TO_CACHE {
				//				fmt.Printf(">MAXIMUM_POINTS: %v\n", point)
				// TODO(greg) lock BP
				err := t.Client.Write(t.BP)
				if err != nil {
					fmt.Println("error in write: " + err.Error())
				}
				t.BP, err = client.NewBatchPoints(t.BPC)
				// TODO(greg) check for leaks here
				if err != nil {
					fmt.Printf("error NewBatchPoints: %v\n", err.Error())
				}
				// TODO(greg) unlock BP
				if err != nil {
					fmt.Printf("error Write: %v\n", err.Error())
				}
				lastWriteAttemptTime = time.Now().UTC()
				waitUntil = lastWriteAttemptTime.Add(MINIMUM_WRITE_INTERVAL_IN_MS * time.Millisecond)
			}

		case <-time.After(timeToWait):
			if len(t.BP.Points()) > 0 {
				//				fmt.Printf("writing %v points\n", len(t.BP.Points()))
				// TODO(greg) lock BP
				err := t.Client.Write(t.BP)
				if err != nil {
					fmt.Println("error in write: " + err.Error())
				}
				// TODO(greg) check for leaks here
				t.BP, err = client.NewBatchPoints(t.BPC)
				if err != nil {
					fmt.Printf("error NewBatchPoints: %v\n", err.Error())
				}
				// TODO(greg) unlock BP
				if err != nil {
					fmt.Printf("error Write: %v\n", err.Error())
				}
			}
			lastWriteAttemptTime = time.Now().UTC()
			waitUntil = lastWriteAttemptTime.Add(MINIMUM_WRITE_INTERVAL_IN_MS * time.Millisecond)

		}

	}

}

func (t *InfluxTSDB) SaveSGMS(s *protobuf.SGMS) {

}

func (t *InfluxTSDB) SaveError() {

}

func (t *InfluxTSDB) SaveAudit() {

}

func (t *InfluxTSDB) SaveSubscription() {

}

func newClient() {

}
func open() {

}

/*
 *
 * Copyright 2015, Google Inc.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above
 * copyright notice, this list of conditions and the following disclaimer
 * in the documentation and/or other materials provided with the
 * distribution.
 *     * Neither the name of Google Inc. nor the names of its
 * contributors may be used to endorse or promote products derived from
 * this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

// Package main implements a simple gRPC client that demonstrates how to use gRPC-Go libraries
// to perform unary, client streaming, server streaming and full duplex RPCs.
//
// It interacts with the route guide service whose definition can be found in proto/route_guide.proto.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"

	tspb "github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/gregzuro/service/plugin/cmd/locationeventspb"
	"github.com/gregzuro/service/plugin/cmd/pkg/decisiontree"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "testdata/ca.pem", "The file containning the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:10011", "The server address in the format of host:port")
	emailFilePath      = flag.String("email_file_path", "", "The file path to the json file")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

const (
	// Seconds field of the earliest valid Timestamp.
	// This is time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).Unix().
	minValidSeconds = -62135596800
	// Seconds field just after the latest valid Timestamp.
	// This is time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC).Unix().
	maxValidSeconds = 253402300800
)

func randomPoint(r *rand.Rand) *pb.Point {
	lat := float64((r.Int31n(180) - 90) * 1e7)
	long := float64((r.Int31n(360) - 180) * 1e7)
	return &pb.Point{lat, long}
}
func validateTimestamp(ts *tspb.Timestamp) error {
	if ts == nil {
		return errors.New("timestamp: nil Timestamp")
	}
	if ts.Seconds < minValidSeconds {
		return fmt.Errorf("timestamp: %v before 0001-01-01", ts)
	}
	if ts.Seconds >= maxValidSeconds {
		return fmt.Errorf("timestamp: %v after 10000-01-01", ts)
	}
	if ts.Nanos < 0 || ts.Nanos >= 1e9 {
		return fmt.Errorf("timestamp: %v: nanos not in range [0, 1e9)", ts)
	}
	return nil
}

func TimestampProto(t time.Time) (*tspb.Timestamp, error) {
	seconds := t.Unix()
	nanos := int32(t.Sub(time.Unix(seconds, 0)))
	ts := &tspb.Timestamp{
		Seconds: seconds,
		Nanos:   nanos,
	}
	if err := validateTimestamp(ts); err != nil {
		return nil, err
	}
	return ts, nil
}
func addDecisionTree(filename string) []*pb.DecisionElement {

	dat, err := ioutil.ReadFile(filename)
	var pbde []*pb.DecisionElement
	if err == nil {

		var tmpJ json.RawMessage

		json.Unmarshal(dat, &tmpJ)
		dte := decisiontree.CreateDecisionTree(tmpJ)
		dte.DecisionArray = decisiontree.AddLevel(dte.DecisionArray)
		dte.BranchArray = decisiontree.Addjumps(dte.DecisionArray, dte.BranchArray)
		dte.DecisionArray = decisiontree.AddBranchestoDecisionTree(dte.BranchArray, dte.DecisionArray)
		var gc decisiontree.GContext
		gc.Device = ""
		dte.DecisionArray = decisiontree.AddVariables(dte.DecisionArray, gc)

		decisiontree.DumpDecisionTree(dte.DecisionArray, dte.BranchArray)

		i := 0
		var isOk bool
		for _, v := range dte.DecisionArray {
			de := new(pb.DecisionElement)
			pbde = append(pbde, de)

			if v.Level == 0 {
				continue
			}

			pbde[i].BranchID = v.BranchID
			pbde[i].Level = v.Level
			ib, isOK := v.Inbranch.(float64)

			if !isOK {
				return nil
			}
			pbde[i].Inbranch = int32(ib)
			pbde[i].Operation, isOk = v.Operation.(string)
			if !isOk {
				return nil
			}
			pbde[i].Property, isOk = v.Property.(string)
			if !isOk {
				return nil
			}
			pbde[i].Level = v.Level
			if !isOk {
				return nil
			}
			pbde[i].Value, isOk = v.Value.(string)
			if !isOk {
				pbde[i].Value = strconv.FormatFloat(v.Value.(float64), 'f', -1, 64)
			}
			pbde[i].TrueBElement = new(pb.BranchElement)
			pbde[i].FalseBElement = new(pb.BranchElement)
			pbde[i].TrueBElement.DTElementIndex = v.TrueBElement.DTElementIndex
			pbde[i].FalseBElement.DTElementIndex = v.FalseBElement.DTElementIndex
			if v.TrueBElement.FunctionCallName != nil {

				pbde[i].TrueBElement.FunctionCallName, isOk = v.TrueBElement.FunctionCallName.(string)
				if !isOk {
					return nil
				}

			}
			if v.FalseBElement.FunctionCallName != nil {

				pbde[i].FalseBElement.FunctionCallName, isOk = v.FalseBElement.FunctionCallName.(string)
				if !isOk {
					return nil
				}
			}
			for _, vv := range v.Variables {
				pbde[i].Variable = append(pbde[i].Variable, vv)

			}
			i++

		}
	}
	return pbde
}
func addEvent(ev pb.EventData) pb.EventData {

	//p(t.Format(time.RFC3339))
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-01-01T10:00:20.021-05:00")
	t2, _ := time.Parse(
		time.RFC3339,
		"2020-01-01T10:00:20.021-05:00")

	ev.ClientId = "d700ac0f-70ef-49c7-ac37-ec94d8aace57"
	//ev.DecisionElements
	ev.Description = ""

	ev.EndTimestamp, _ = TimestampProto(t1)
	ev.StartTimestamp, _ = TimestampProto(t2)
	ev.EventId = "228304ea-088c-44b7-9ff7-b45fd4bbff17"
	evtemp := addDecisionTree(*emailFilePath)
	for _, v := range evtemp {
		ev.DecisionElements = append(ev.DecisionElements, v)

	}
	return ev
}
func addLandmark(lm pb.LandmarkData) pb.LandmarkData {

	//p(t.Format(time.RFC3339))

	lm.PoiId = "ef57cfb6-cd20-48ca-a585-9345f978ed92"

	lm.Events = append(lm.Events, &pb.EventElement{EventID: "228304ea-088c-44b7-9ff7-b45fd4bbff17", PresenceCounter: 1})
	lm.ClientId = "ef57cfb6-cd20-48ca-a585-9345f978ed93"
	//	ll := pb.Point{X: -1.0, Y: -1.0}
	//	ul := pb.Point{X: -1.0, Y: 1.0}
	//	ur := pb.Point{X: 1.0, Y: 1.0}
	//	lr := pb.Point{X: 1.0, Y: -1.0}
	lm.PolygonPoints[0] = new(pb.Polygon)
	//lm.PolygonPoints = append(lm.PolygonPoints[0].PolygonPoints, &ll, &ul, &ur, &lr)
	//ev.DecisionElements
	return lm
}
func runUpdateEvent(client pb.LocationEventClient) {
	// Create a random number of random points

	ev := pb.EventData{}
	ev = addEvent(ev)

	stream, err := client.UpdateEvent(context.Background())

	if err != nil {
		grpclog.Fatalf("%v.UpdateEvent(_) = _, %v", client, err)
	}
	for i := 0; i < 10; i++ {
		if err := stream.Send(&ev); err != nil {
			grpclog.Fatalf("%v.Send(%v) = %v", stream, ev, err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		grpclog.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	grpclog.Printf("UpdateEvent summary: %v", reply)
}

// runRecordRoute sends a sequence of points to server and expects to get a RouteSummary from server.
func runUpdateLandmark(client pb.LocationEventClient) {
	// Create a random number of random points
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//landmarkPointCount := int(r.Int31n(100)) + 2 // Traverse at least two points

	lm := pb.LandmarkData{}

	lm.ClientId = "ef57cfb6-cd20-48ca-a585-9345f978ed93"
	lm.PoiId = "ddd7cfb6-cd20-48ca-a585-9345f978ed93"
	//lm.PolygonPoints = make([]*pb.Point, 4)

	//lm.PolygonPoints[0] = &pb.Point{X: -1.0, Y: -1.0}
	//lm.PolygonPoints[1] = &pb.Point{X: -1.0, Y: 1.0}
	//lm.PolygonPoints[2] = &pb.Point{X: 1.0, Y: 1.0}
	//lm.PolygonPoints[3] = &pb.Point{X: 1.0, Y: -1.0}

	lm = addLandmark(lm)
	/*for i := 0; i < landmarkPointCount; i++ {
		lm.PolygonPoints = append(lm.PolygonPoints, randomPoint(r))
	}*/
	stream, err := client.UpdateLandmark(context.Background())
	if err != nil {
		grpclog.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}
	for i := 0; i < 10; i++ {
		if err := stream.Send(&lm); err != nil {
			grpclog.Fatalf("%v.Send(%v) = %v", stream, lm.PolygonPoints, err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		grpclog.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	grpclog.Printf("Route summary: %v", reply)
}
func runUpdateLocation(client pb.LocationEventClient) {
	// Create a random number of random points
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//landmarkPointCount := int(r.Int31n(100)) + 2 // Traverse at least two points

	lm := pb.LocationData{}
	//43.206518574277965 -77.49562394711833
	//43.19812145996094, -77.504884002736
	lm.Latitude = 43.19812145996094
	lm.Longitude = -77.504884002736
	//[0]:43.19812145996094
	//[1]:-77.504884002736
	//[2]:43.20712145996094
	//[3]:-77.49253726191243
	//[0]:43.19812145996094
	//[1]:-77.504884002736
	//[2]:43.20712145996094
	//[3]:-77.49253726191243

	//grpclog.Printf("Traversing %d points.", len(lm.PolygonPoints))
	stream, err := client.UpdateLocation(context.Background())
	if err != nil {
		grpclog.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}
	for i := 0; i < 1; i++ {
		if err := stream.Send(&lm); err != nil {
			grpclog.Fatalf("%v.Send(lm) = %v", stream, err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		grpclog.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	grpclog.Printf("Route summary: %v", reply)
}

func main() {
	flag.Parse()

	//b, err := ioutil.ReadFile(*emailFilePath)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(string(b))

	var opts []grpc.DialOption
	if *tls {
		var sn string
		if *serverHostOverride != "" {
			sn = *serverHostOverride
		}
		var creds credentials.TransportCredentials
		if *caFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(*caFile, sn)
			if err != nil {
				grpclog.Fatalf("Failed to create TLS credentials %v", err)
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, sn)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewLocationEventClient(conn)
	//	runUpdateEvent(client)
	//	runUpdateLandmark(client)

	runUpdateLocation(client)

}

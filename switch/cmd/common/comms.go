package common

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/gregzuro/service/switch/insecure"
	"github.com/gregzuro/service/switch/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

var (
	// KeyPair is a pointer to a TLS keypair
	KeyPair *tls.Certificate
	// CertPool is an x509 thingie
	CertPool *x509.CertPool
)

// GetCertPool is a getter
func GetCertPool() *x509.CertPool {
	return CertPool
}

// GetKeyPair is a getter
func GetKeyPair() *tls.Certificate {
	return KeyPair
}

// findGoodSlave returns a child that's a slave and that has had a heartbeat recently
func findGoodSlave(slaves map[string]protobuf.Entity) (protobuf.Entity, error) {

	// just taking the first item does NOT result in a uniform distribution,
	// so we iterate through the list a random number of times, instead.

	liveOnes := make(map[string]protobuf.Entity)
	for k, v := range slaves {
		if time.Since(ToGoTime(v.LastSeen)) < time.Second*10 {
			liveOnes[k] = v
		}
	}

	if len(liveOnes) == 0 {
		return protobuf.Entity{}, errors.New("no live slaves")
	}
	rand.Seed(int64(time.Now().Nanosecond())) // Try changing this number!

	var err error
	i := 0
	var v protobuf.Entity
	want := rand.Intn(len(liveOnes))
	for _, v = range liveOnes {
		if i == want {
			break
		}
		i++
	}

	return v, err
}

// determineGeoAffinityForSlave returns the index to the GA that's appropriate for a (the?) new slave
func determineGeoAffinityForSlave(masterGA []*protobuf.GeoAffinity, slaves map[string]protobuf.Entity) (int, error) {

	// return index to the GA that has the lowest 'coverage'
	var deficit int32
	var possibleK int
	for k, v := range masterGA {
		if v.WantCovering-int32(len(v.Coverers)) > deficit {
			deficit = v.WantCovering - int32(len(v.Coverers))
			possibleK = k
		}
	}

	if deficit > 0 {
		return possibleK, nil
	}

	return 0, errors.New("no slaves needed - all GA covered as wanted")

	// if len(slaves) == 0 {

	// 	return masterGA, nil
	// } else {
	// 	return []*protobuf.GeoAffinity{}, errors.New("don't know what to do WRT GA for new slave")
	// }

	// TODO(greg)
	// if there are slaves that need help, then pick one

	// if there are no slaves that need help, then tell the slave to quit

	// set the GA for the new slave to some portion of the existing slave's GA (this new GA should be provided by the existing slave)

	// kick off a split management action if necessary

	// return

}

// GrpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func GrpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(tamird): point to merged gRPC code rather than a PR.
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

// AllowCORS allows Cross Origin Resource Sharing from any origin.
// Don't do this without consideration in production systems.
func AllowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	grpclog.Printf("preflight request for %s", r.URL.Path)
	return
}

// SetUpDialer establishes connections to all of the parent's services
func SetUpDialer(address string, port int64) (*grpc.ClientConn, *RPCClients, error) {

	var clients RPCClients
	var opts []grpc.DialOption
	creds := credentials.NewClientTLSFromCert(CertPool, "dev") // TODO(greg) allow different keys
	opts = append(opts, grpc.WithTransportCredentials(creds))
	server := fmt.Sprintf("%s:%d", address, port)
	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
		return conn, &clients, err
	}

	// get all the client handles
	clients.SGMS = protobuf.NewSGMSServiceClient(conn)
	clients.Status = protobuf.NewStatusServiceClient(conn)
	clients.Contact = protobuf.NewInitialContactServiceClient(conn)
	return conn, &clients, nil
}

// GetParentForDevice determines the parent (slave) for a device node by inquiring with the specified master node
func GetParentForDevice(args CommandArgs, entity protobuf.Entity) (string, int64, error) {
	var clients RPCClients
	var opts []grpc.DialOption
	creds := credentials.NewClientTLSFromCert(CertPool, "dev") // TODO(greg) allow different keys
	opts = append(opts, grpc.WithTransportCredentials(creds))
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d",
		args.MasterAddress,
		args.MasterPort), opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
		return "", 0, err
	}

	// get the client handle
	clients.Contact = protobuf.NewInitialContactServiceClient(conn)

	// TODO(greg) need to transition from master to generic parent for plugins ans such
	_goto, err := clients.Contact.InitialContact(context.Background(),
		&protobuf.Hello{
			CallerEntity: &protobuf.Entity{Id: entity.Id, Kind: entity.Kind, Address: entity.Address, Port: entity.Port},
			CalledEntity: &protobuf.Entity{Kind: "master", Address: args.MasterAddress, Port: args.MasterPort},
		})

	if err != nil {
		log.WithField("context", "InitialContact").Fatal(err)
	}

	//	output := fmt.Sprixxntf("address: %v, port: %v, err: %v\n", _goto.Address, _goto.Port, err)
	log.WithFields(log.Fields{
		"context": "InitialContact",
		"address": _goto.Address,
		"port":    _goto.Port}).Info()

	return _goto.Address, _goto.Port, err
}

func init() {
	var err error
	pair, err := tls.X509KeyPair([]byte(insecure.DevCert), []byte(insecure.DevKey))
	if err != nil {
		log.Fatal(err)
	}
	KeyPair = &pair
	CertPool = x509.NewCertPool()
	ok := CertPool.AppendCertsFromPEM([]byte(insecure.DevCert))
	if !ok {
		log.Fatal("bad certs")
	}
}

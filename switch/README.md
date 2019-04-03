Serial Global Message Switch (SGMS)

# Development Environment

Before you get started, follow "Set up your environment" in the [main README.md](../README.md).  A development environment with at least 2GB of RAM is recommended to avoid low memory errors.  

## Quick start

Run `local-cluster-up` to bring up the whole environment, including log, db, grafana, and master switch. This also stops any running nodes.

After this, verify that all nodes are running with `docker ps`. It should list

 - `sgms_master1`
 - `fluentd_master1`
 - `grafana_master1`
 - `influxdb_master1`
 
Once these are running, you can rebuild/start just the switch with `make local-switch-up`.

## Make tasks

Read the [Makefile](./Makefile) to see what's going on, but the basic high-level targets are:

 - `local-cluster-up`: bring down: everything; then bring up: log, db, grafana, master and device switches
 - `local-other-up`: bring down: log, db, and grafana down and then up: log, db, and grafana
 - `local-switches-up`: bring down: master, slave, and device switch nodes; then: rebuild the go binaries and docker images; then bring up: master switch nodes
 - `switches`: build the native binaries which you can run outside of Docker. Useful for one-off testing.
 - `master`: build the master switch binary
 - `master-container`: build the master switch docker container
 - `containers`: build the master, slave, and device containers

## Web console

The master node includes a web UI, from which you can browse the cluster.

https://localhost:7456/ui

## Starting slaves

The master node is responsible for starting new slave nodes but currently doesn't start one automatically.

You'll need to instruct it to start one by attaching to the master console and issuing a command:

```
docker attach sgms_master1
amp> start slave
```

`Ctrl-p` `Ctrl-q` (detach without terminating the process)

## Test devices

After starting a slave node as above, you can run `make local-devices-up` to start a number (1 by default, setable in the Makefile) of test device nodes generating "random walk" data.

### Connecting external clients to your dev cluster

The Dockerized local dev cluster runs in a private bridge network, with service ports forwarded to all interfaces on the host machine. To connect to a cluster running with forwarded ports from outside (from the host itself, or via a LAN or internet address):

 - Pass the `-m` (`--port-mapped`) flag to `switch device`.
 - If instantiating a client in code, set the corresponding `PortMappedCluster: true` flag.
 - If building a custom gRPC client in another language, implement a dev mode which ignores the `address` field in the `Goto` message returned from the `InitialContact` endpoint, and reuses the master address instead.

### Interacting with the API

Although the main protocol for the API is Protobuf/gRPC, you can interact with it from a web client via the built-in grpc-gateway. For example:

```
curl -k --data '{"value":"health"}'  https://localhost:7467/v1/health
```

### API console

Each node also has an embedded API browser at `/swagger-ui`. You can access this on the master1 node of the above cluster setup at https://localhost:7467/swagger-ui/ .

## Database

Messages are logged to InfluxDB. In the local environment, you can interact with it via:

 - InfluxDB admin console: http://localhost:8083/ (influxdb admin console)
 - Grafana: http://localhost:3000/
 
Example queries:

 - `select mean(velocity) from gps_location where time > "2016-01-01" group by time(1h)`

## Logging

You can find the log for the master instance at `../var/logs/log.DATE.instanceID`.

Slave instances are only logging to stdout currently. You can examine their logs with `docker logs NAME`.

[Fluentd](http://docs.fluentd.org/) listens to Docker containers' stdout and can send logs to downstream services. Currently it is listening to the master node's output and writing it to the above file.

## Tips for Working with the Go code

- Go comes with a suite of excellent static analysis tools, which enable editors like Sublime Text, VS Code, and Emacs to provide "Jump to definition", "find uses", etc. Check your editor's go package docs.
- To browse package documentation:
  1. Fire up a godoc instance: `godoc -http=6060`
  2. Visit http://localhost:6060/pkg/github.com/gregzuro/service/switch

## Hacking on the web UI

To make changes to the web UI:

 1. Get the Elm compiler.
 2. Run `make` from the UI directory. It's not necessary to restart the node to see your changes.
 3. When you are happy with your changes, submit the compiled JS output along with the elm source.
 
```bash
$ cd ...go/src/github.com/gregzuro/service/switch 
$ compiledaemon -recursive=true -build="make local-switch-up" -exclude-dir=.git
```

## CompileDaemon (optional)

I like to have things rebuilt automatically when I save, so you can get `CompileDaemon` from https://github.com/gregzuro/CompileDaemon .

If you do: 

then the entire switch will be rebuilt and (re-)deployed to your local machine (docker) every time you save a Go source file.
If there are any Go build errors, they will be shown in that console window.

## GKE Testing

For testing in Google Container Engine (GKE):

### Prerequisites

#### GKE credentials

See here: https://cloud.google.com/container-engine/, but there is a company account that 

#### gcloud

https://cloud.google.com/sdk/gcloud/

Use

```bash
$ gcloud config set project <project>
```

where \<project\> is { `gregzuro-dev` | `gregzuro-test` }, to select the GCE 'project'.

You'll almost be using `gregzuro-dev` during your development efforts.

#### kubectl

This will be installed as part of gcloud except for Debian and Ubuntu.
See here: https://cloud.google.com/sdk/downloads#apt-get

### Procedure

#### Create a Cluster

```bash
$ cd ...go/src/github.com/gregzuro/service/switch 
$ TEST_CLUSTER_NAME=<cluster-name> make test-create-cluster
```

For example:

```bash
$ TEST_CLUSTER_NAME=test TEST_CLUSTER_NODES=3 make test-create-cluster 
# create cluster (this blocks)
gcloud container clusters create test --zone us-central1-c --machine-type n1-standard-2 --num-nodes 2 --password test 
Creating cluster test...done.
Created [https://container.googleapis.com/v1/projects/gregzuro-1280/zones/us-central1-c/clusters/test].
kubeconfig entry generated for test.
NAME  ZONE           MASTER_VERSION  MASTER_IP     MACHINE_TYPE   NODE_VERSION  NUM_NODES  STATUS
test  us-central1-c  1.2.4           104.154.39.3  n1-standard-2  1.2.4         3          RUNNING
gcloud container clusters get-credentials test
Fetching cluster endpoint and auth data.
kubeconfig entry generated for test.
Kubernetes master is running at https://104.154.39.3
GLBCDefaultBackend is running at https://104.154.39.3/api/v1/proxy/namespaces/kube-system/services/default-http-backend
Heapster is running at https://104.154.39.3/api/v1/proxy/namespaces/kube-system/services/heapster
KubeDNS is running at https://104.154.39.3/api/v1/proxy/namespaces/kube-system/services/kube-dns
kubernetes-dashboard is running at https://104.154.39.3/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
```

will take a few minutes to create a cluster of 3 nodes that's named `test`.
In addition, a `get-credentials` is done so that you can immediately use `kubectl` on the cluster.
The credentials for logging in to the cluster (dashboard, etc) will be `admin/$TEST_CLUSTER_NAME`.

Note that, for now, the cluster will not autoscale.
If you want auto-scaling, you can manually go into the GCE control panel and edit the instance group for the cluster.

#### Delete a Cluster

```bash
$ TEST_CLUSTER_NAME=<cluster-name> make test-delete-cluster
```

For example: 

```bash
$ TEST_CLUSTER_NAME=test make test-delete-cluster 
gcloud container clusters delete test
The following clusters will be deleted.
 - [test] in [us-central1-c]

Do you want to continue (Y/n)?  y

Deleting cluster test...done.
Deleted [https://container.googleapis.com/v1/projects/gregzuro-1280/zones/us-central1-c/clusters/test].
```

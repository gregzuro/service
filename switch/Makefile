include ../.env.mk
SHELL=/bin/bash

.PHONY: all switch build local-nodes-down test local-db-down local-db-down test-cluster-up test-cluster-down local-up local-down switch-container

all: build/linux_amd64/switch

SWITCHDEPS=protobuf/* cmd/common/*.go ui
ifndef SS
	SS = 0
endif
SP= $(shell expr 7500 + $(SS) )
SN= $$(printf 'sgms_slave%03d' "$(SS)")

build/linux_amd64/switch: $(SWITCHDEPS) db/*.go cmd/*.go cmd/master/*.go cmd/slave/*.go cmd/device/*.go ui protobuf/sgms.pb.go ../pkg/remote/*.go ../pkg/message/*.go
	mkdir -p $$(dirname build/linux_amd64)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/linux_amd64/switch 

../plugin/cmd/locationevents/locationevents_linux_amd64: $(SWITCHDEPS) ../plugin/cmd/*/*.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $@ $(NAME)/plugin/cmd/locationevents

FORCE:

protobuf/sgms.pb.go: protobuf/sgms.proto protobuf/scripts/includetxt.go
	make -C protobuf

ui: ui/legacy

ui/legacy: ui/legacy/*
	cd ui && elm-make Main.elm

# docker doesn't need sudo on Darwin
ifeq ($(strip $(shell uname)), Darwin)
dockercmd = docker
else
dockercmd = sudo docker
endif

var ?= $(PWD)/../var
$(var):
	mkdir $@
	mkdir $@/lib
	mkdir $@/log

local-cluster-up: local-other-up local-nodes-up

lcu: local-cluster-up

terraform-cluster-up: server-containers
	terraform apply -var "cluster_name=$(CLUSTER_PREFIX)-$$(date +%Y%m%d%H%M%S)"

tcu: terraform-cluster-up

local-cluster-down: local-nodes-down local-other-down

lcd: local-cluster-down

terraform-cluster-down: 
	terraform destroy -var 'cluster_name=$(CLUSTER_NAME)'

tcd: terraform-cluster-down

local-other-up: local-other-down local-db-up local-grafana-up local-log-up local-registry-up

lou: local-other-up

local-other-down: local-db-down local-grafana-down local-log-down local-registry-down

lod: local-other-down

local-nodes-up: local-nodes-down local-masters-up 

lnu: local-nodes-up

local-nodes-down: local-devices-down local-slaves-down local-masters-down

lnd: local-nodes-down

local-registry-up: local-registry-down
	-$(dockercmd) run -d -p 5000:5000 --restart=always --name registry registry:2.6.0

lru: local-registry-up 

local-registry-down: 
	-$(dockercmd) stop registry
	-$(dockercmd) rm registry

lrd: local-registry-down

local-db-up: local-db-down $(var)
	-$(dockercmd) network create --driver bridge sgms
	$(dockercmd) run --net=sgms --name influxdb_master1 -d -p 8083:8083 -p 8086:8086 -p 8089:8089 \
	-v $(var)/lib/influxdb:/var/lib/influxdb \
	-v $(PWD)/db/influxdb.conf:/etc/influxdb/influxdb.conf:ro \
	influxdb:1.1.0 \
	-config /etc/influxdb/influxdb.conf
	sleep 2
	-$(dockercmd) exec influxdb_master1 influx -execute "create database sgms"
	-$(dockercmd) exec influxdb_master1 influx -execute "create database logs"
	-$(dockercmd) exec influxdb_master1 influx -execute "create user sgms with password 'local'"
	-$(dockercmd) exec influxdb_master1 influx -execute "grant all on sgms to sgms"

local-db-down:
	-$(dockercmd) stop influxdb_master1
	-$(dockercmd) rm influxdb_master1

local-grafana-up: local-grafana-down
	$(dockercmd) run -d --net=sgms --name=grafana_master1 -p 3000:3000 \
	-v $(var)/lib/grafana:/var/lib/grafana \
	grafana/grafana

local-grafana-down:
	-$(dockercmd) stop grafana_master1
	-$(dockercmd) rm grafana_master1

local-log-up: local-log-down $(var)
	$(dockercmd) build -t gregzuro/fluentd ./fluentd
	$(dockercmd) run --net=sgms --name fluentd_master1 -d -p 24224:24224 \
	-v $(var)/log:/fluentd/log \
	gregzuro/fluentd

local-log-down:
	-$(dockercmd) stop fluentd_master1
	-$(dockercmd) rm fluentd_master1

containers: switch-container

server-containers: switch-container

switch-container: tmp_switch-container-built

tmp_switch-container-built: cmd/Dockerfile build/linux_amd64/switch cmd/master/configs/* cmd/slave/configs/* cmd/device/configs/*
	# make the switch container
	$(dockercmd) build --build-arg DEPLOY_ENV=local -t gregzuro/switch -f cmd/Dockerfile .
	touch tmp_switch-container-built
	# push to our registry
	## $(dockercmd) tag gregzuro/switch-master gregzuro/switch-master
	#$(dockercmd) push gregzuro/switch-master 

locationevents-container: ../plugin/cmd/locationevents/locationevents_linux_amd64 ../plugin/cmd/locationevents/Dockerfile
	$(dockercmd) build -t gregzuro/plugin/locationevents -f ../plugin/cmd/locationevents/Dockerfile ../plugin/cmd/locationevents/

local-masters-up: local-masters-down local-slaves-down switch-container
	# start the master
	$(dockercmd) run -d -it \
	-p 7467:7467 \
	--net=sgms \
	--name sgms_master1 \
	-v $(PWD)/ui:/srv/ui \
	--log-driver=fluentd \
	--log-opt tag=service.master \
	gregzuro/switch \
	./switch master -l 7467 --short-name sgms_master1 --iga '[[-123,45],[-123,46],[-122,46],[-122,45]]' 
	# -iga '[[-124,45],[-124,46],[-123,46],[-123,45]]' --iga '[[-123,44],[-123,45],[-122,45],[-122,44]]' --iga '[[-124,44],[-124,45],[-123,45],[-123,44]]' 	

lmu: local-masters-up

local-masters-down:
	-$(dockercmd) stop $$($(dockercmd) ps -q --filter name=sgms_master)
	-$(dockercmd) rm $$($(dockercmd) ps -a -q --filter name=sgms_master)

local-slave-start: switch-container
	# start a slave
	$(dockercmd) run -d \
	-p $(SP):$(SP) \
	--net=sgms \
	--name $(SN) \
	--log-driver=fluentd \
	--log-opt tag=service.slave \
	gregzuro/switch \
	./switch slave -a sgms_master1 -p 7467 -l $(SP) --short-name $(SN)

lss: local-slave-start

local-slaves-down:
	-$(dockercmd) stop $$($(dockercmd) ps -q --filter name=sgms_slave)
	-$(dockercmd) rm $$($(dockercmd) ps -a -q --filter name=sgms_slave)

lsd: local-slaves-down

n = 1  # <- threads * 10 devices * 10 Hz = (maximum of) 100/second

local-devices-up: local-devices-down containers
	# start a bunch of devices
	$(dockercmd) run -d \
	-p 7601:7601 \
	--net=sgms \
	--name sgms_device001 \
	--log-driver=fluentd \
	--log-opt tag=service.device \
	gregzuro/switch \
	./switch device -a sgms_master1 -p 7467 -l 7601 --short-name sgms_device001 -n $(n)

ldu: local-devices-up

local-devices-down: 
	-$(dockercmd) stop $$($(dockercmd) ps -q --filter name=sgms_device)
	-$(dockercmd) rm $$($(dockercmd) ps -a -q --filter name=sgms_device)

ldd: local-devices-down

local-plugin-up: local-plugins-down locationevents-container
	$(dockercmd) run -d -it \
	-p 10011:10011 \
	--net=sgms \
	--name sgms_plugin001 \
	--log-driver=fluentd \
	--log-opt tag=plugin.locationevents \
	gregzuro/plugin/locationevents \
	./locationevents -saddr sgms_slave000 -sport 7500

local-plugins-down:
	-$(dockercmd) stop $$($(dockercmd) ps -q --filter name=sgms_plugin)
	-$(dockercmd) rm $$($(dockercmd) ps -a -q --filter name=sgms_plugin)

test:
	go test ./...

###### not used:

test-cluster-up:
	# create cluster (this blocks)
	gcloud container clusters create $(TEST_CLUSTER_(NAME)) --zone us-central1-c --machine-type n1-standard-2 --num-nodes $(TEST_CLUSTER_NODES) --password $(TEST_CLUSTER_(NAME)) 
	gcloud container clusters get-credentials $(TEST_CLUSTER_(NAME))
	kubectl cluster-info

test-cluster-down:
	gcloud container clusters delete $(TEST_CLUSTER_(NAME))


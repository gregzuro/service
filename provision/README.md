# Terraform

## Prerequisites

- triton CLI
- json tool
- terraform
- packer

## Set up

### Triton

Currently, the terraform scripts in dev/ will make a cluster on Joyent's Triton cloud service.

- Get a Joyent account set up. The Triton account name shall be of the form `gregzuro_<foo>.
- Add your ssh key to that account (or create a new one) using the Joyent UI: https://my.joyent.com/main/#!/account .
- Check `Use Container Name Service` near the bottom of this page: https://my.joyent.com/main/#!/account/edit .

Install triton as described here: https://github.com/joyent/node-triton .

Create a Triton profile on your local machine with:

```
$ triton profile create
```

Configure your current shell environment variables to allow communication with Triton's docker endpoint:

```
$ eval "$(triton env)"
```

We use Hashicorp's Terraform to create a cluster.
The Triton profile name must match the name used as the root of the terraform var-file name.
This profile must match - account-wise - the one specified in the Terraform var-file.

[This is necessary for the provisioning of the influxdb docker container (creation of databases, users, and permissions) through the use of `local-exec` Terraform provisioners which means that your local Docker must be able to connect to Triton's Docker implementation (sdc-docker).]

Clusters are isolated by using seperate Triton accounts: so each Triton account can have a single cluster.
The Terraform scripts depend on a particular format for the Triton account name: `gregzuro_<foo>{_<bar>}` , where <foo> is used as a subdomain name for the cluster:

Running the Terraform scripts with a Triton account having the name `gregzuro_greg1` will result in a cluster with components accessible via (DNS): `<component>.greg1.dev.gregzuro.io`.
Where component is one of:

- influxdb
- fluentd
- master
- slave_xxx
- consul_xx

- grafana
- registry
- vault

### Google

In order to run the scripts, you need to have a file in the secrets/ directory called `gregzuro-dev-*.json` that looks something like:

```
{
  "type": "service_account",
  "project_id": "gregzuro-dev",
  "private_key_id": "373914f17...5670eb3cf3677fd6",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDMDjFzS99Omt3x\nOMV8TLYhlUMTzfFl1Wcs7h1mmf5OOc2uNjmCxyaM+Wd...
  L7LNwKBgEjdAPkpHS2xLxcYn+ln\nguAi++gjHBbDy7OPvtBabIVixZW7ns3FXD1Hasip71Xl4Ai0sHFa/I8T3pyzTN3f\njcoEg5m7DZWi5gaPCpkeaZGJYTOu4M4Cd8XWSAyTy2QiSEn3gbR7bwFO9NsbhwAB\nsIXjT8EDSYENXEDAwYCdddA8\n-----END PRIVATE KEY-----\n",
  "client_email": "589224277445-compute@developer.gserviceaccount.com",
  "client_id": "113917210144324923376",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/589224277445-compute%40developer.gserviceaccount.com"
}

```

This file is created via the Google Developers Console as described at the bottom of this page: https://www.terraform.io/docs/providers/google/ .
The name of the generated file must be specified using the 

You also need to have a file in the dev/ directory named like `gregzuro_greg1.triton.joyent.us-west-1.dev.tfvars` that looks something like:

```
account_name = "gregzuro_greg1"
account_uuid = "e581a016-3711-6af9-cbfd-9c379b218327"
triton_url = "https://us-west-1.api.joyentcloud.com"
triton_key_path = "~/.ssh/joyent2_id_rsa"
triton_key_id = "19:29:e2:d5:29:75:0e:b7:d4:53:f3:05:d7:67:2a:12"
data_center = "us-west-1"
dns_domain = "joyent.com"
docker_cert_path = "/.triton/docker/gregzuro_greg1@us-west-1_api_joyent_com"
docker_host = "tcp://us-west-1.docker.joyent.com:2376"
google_creds_file = "../secrets/gregzuro-dev-45c1d404d431.json"
```

Create or modify the cluster with:
```
$ cd provision
$ ./mtf dev switches gregzuro_greg1.triton.joyent.us-west-1.dev plan
...
$ ./mtf dev switches gregzuro_greg1.triton.joyent.us-west-1.dev apply
```

## Background

The generic services (listed above) are available in pre-built docker containers.  
We simply use those.
We use the Terraform providers for those services to configure them once they are running.

We manually create two docker containers: `vault` at https://hub.docker.com/r/gregzuro/vault/ and `fluentd` at <TBD> because: reasons.  

Reasons:

- vault wouldn't start 'cause it wants to setcap, but that won't work if docker is backed by aufs or Triton. I've 'fixed' it here: https://github.com/gregzuro/docker-vault/commit/66d8bbcd30b38c49750692ad1f735314f26f83ed
- fluentd needs to have the influxdb plugin, so we add that as shown here:

```
FROM fluent/fluentd:latest-onbuild
RUN fluent-gem install fluent-plugin-influxdb
EXPOSE 24284
CMD fluentd -c /fluentd/etc/$FLUENTD_CONF -p /fluentd/plugins $FLUENTD_OPT
```

These are pushed to the public docker registry so they can be used by Terraform/Triton during the cluster bootstrap process.

# Packer

## Creating the image

Once you have a Triton account of the form ```gregzuro_foo```, create a packer template like packer/dev/switch.json that contains the correct key info.

Run packer

```
$ packer  build -only=gregzuro_greg1.triton.joyent.us-west-1.dev.switch ./switch.json
gregzuro_greg1.triton.joyent.us-west-1.dev.switch output will be in this color.

==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Creating source machine...
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Waiting for source machine to become available...
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Waiting for SSH to become available...
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Connected to SSH!
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Provisioning with shell script: /var/folders/41/s0qrql4j07l0kwyhtb2n7gvh0000gp/T/packer-shell813867525
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Uploading ./cmd/master/configs/ => /srv/configs/master/
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Uploading ./cmd/slave/configs/ => /srv/configs/slave/
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Uploading ./cmd/device/configs/ => /srv/configs/device/
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Uploading ./ui/ => /srv/ui/
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Uploading ./build/linux_amd64/switch => /srv/
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Uploading ./third_party/ => /srv/third_party/
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Stopping source machine (1f5aa7f6-0e4b-6fa8-9a1f-bd758f944b44)...
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Waiting for source machine to stop (1f5aa7f6-0e4b-6fa8-9a1f-bd758f944b44)...
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Creating image from source machine...
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Waiting for image to become available...
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Deleting source machine...
==> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Waiting for source machine to be deleted...
Build 'gregzuro_greg1.triton.joyent.us-west-1.dev.switch' finished.

==> Builds finished. The artifacts of successful builds are:
--> gregzuro_greg1.triton.joyent.us-west-1.dev.switch: Image was created: af144cae-187d-42da-8d62-30f4e36111be
```

## Provisioning nodes with the image

Now, you can create nodes of any flavor from this image.

### Master 

Create a master node:

```
$ cd provision/
$ triton instance create af144cae-187d-42da-8d62-30f4e36111be g4-highcpu-512M --script=../switch/cmd/master/start.sh -n master -t triton.cns.services=master
Creating instance master (83b571f7-5930-661b-ab4e-a85ce4f4ebf6, switch@170301052207)
```

### Slave

Create a slave node:

```
$ triton instance create af144cae-187d-42da-8d62-30f4e36111be g4-highcpu-512M --script=../switch/cmd/slave/start.sh -n slave001 -t triton.cns.services=slave -m master=master.inst.e581a016-3711-6af9-cbfd-9c379b218327.us-west-1.triton.zone
Creating instance slave001 (2eab2905-3262-60be-ab56-d535152a1ec3, switch@170301052207)
```


### Device

```
$ triton instance create af144cae-187d-42da-8d62-30f4e36111be g4-highcpu-512M --script=../switch/cmd/device/start.sh -n device001 -t triton.cns.services=device -m master=master.inst.e581a016-3711-6af9-cbfd-9c379b218327.us-west-1.triton.zone
Creating instance device001 (c1908d13-ba25-ed77-def1-901a97a782eb, switch@170301052207)
```

## Notes

Though the instances are created by the commands above, the switch process is run by the specified script.
Look in those scripts for more insight into that.

To see all the images that you may have created, try:

```
$ triton images public=false
SHORTID   NAME    VERSION       FLAGS  OS     TYPE        PUBDATE
19768a76  switch  170223025748  I      linux  lx-dataset  2017-02-23
a01d7675  switch  170223033855  I      linux  lx-dataset  2017-02-23
af144cae  switch  170301052207  I      linux  lx-dataset  2017-03-01
```

The VERSION value is of the form YYMMDDHHMMSS based on the build time.

You may use the SHORTID form of the image identifier in your ```triton instance create``` command.

The Joyent Triton UI ```my.joyent.com``` does not reliably show running instances when you go to "Compute->Instances".
Instead use ```$ triton instances``` to see what is running.

## Artifacts

To get the uuid of the image that's created, you can do something like:

```
$ cat manifest.json | json builds| json -c 'this.name == "gregzuro_greg1.triton.joyent.us-west-1.dev.switch"' | json -a build_time artifact_id|sort -n|tail -1
```

This depends on the use of Packer's `manifest` post-processor which updates the manifect file with the details of the created artifacts.
This should be passed to Terraform (it's manually set in local.tfvars at the moment) in an automated fashion.

## Terraform

The starting of these instances will be done by the terraform scripts for which these examples serve as a model.









provider "google" {
  credentials = "${file("../secrets/gregzuro-dev-45c1d404d431.json")}"
  project     = "gregzuro-dev"
  region      = "us-west1"
}

# triton_account_name must be of the form: "foo_bar{_baz}" where bar is used as the sub-domain and the rest are ignored
resource "google_dns_managed_zone" "subzone" {
  name     = "${element(split("_",var.account_name),1)}"
  dns_name = "${element(split("_",var.account_name),1)}.dev.gregzuro.io."
}

resource "google_dns_record_set" "parent-zone" {
  name         = "${google_dns_managed_zone.subzone.dns_name}"
  type         = "NS"
  ttl          = 300
  managed_zone = "dev-gregzuro-io"
  rrdatas      = ["${google_dns_managed_zone.subzone.name_servers}"]
}

provider "triton" {
  account      = "${var.account_name}"
  key_material = "${file("${var.triton_key_path}")}"
  key_id       = "${var.triton_key_id}"
  url          = "${var.triton_url}"
}

provider "docker" {
  host      = "${var.docker_host}"
  cert_path = "${var.home}${var.docker_cert_path}/"
}

resource "docker_image" "consul" {
  name         = "autopilotpattern/consul:0.7r0.7"
  keep_locally = true
}

resource "docker_container" "consul" {
  # entrypoint = ["/usr/local/bin/containerpilot"]
  command = [
    "/usr/local/bin/containerpilot",
    "/bin/consul",
    "agent",
    "-server",
    "-bootstrap-expect",
    "3",
    "-config-dir=/etc/consul",
    "-ui-dir",
    "/ui",
  ]

  env        = ["CONSUL=consul.svc.${var.account_uuid}.${var.data_center}.cns.${var.dns_domain}"]
  count      = "${var.consul_count}"
  name       = "${format("consul-%01d", count.index)}"
  image      = "${docker_image.consul.latest}"
  must_run   = true
  restart    = "always"
  memory     = 1024
  log_driver = "json-file"

  log_opts = {
    max-size = "1m"
    max-file = 2
  }

  # ports {
  #   internal = 53
  #   external = 53
  # }
  ports {
    internal = 8300
    external = 8300
  }

  ports {
    internal = 8301
    external = 8301
    protocol = "udp"
  }

  ports {
    internal = 8302
    external = 8302
    protocol = "udp"
  }

  ports {
    internal = 8400
    external = 8400
  }

  ports {
    internal = 8500
    external = 8500
  }

  ports {
    internal = 8600
    external = 8600
    protocol = "udp"
  }

  labels = {
    "triton.cns.services" = "consul"
  }
}

resource "google_dns_record_set" "consul" {
  count        = "${var.consul_count}"
  name         = "${element(docker_container.consul.*.name, count.index)}.${google_dns_managed_zone.subzone.dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas      = ["${element(docker_container.consul.*.ip_address, count.index)}"]
}

resource "triton_machine" "master" {
  name    = "master"
  package = "g4-highcpu-512M"

  tags = {
    "triton.cns.services" = "amp.master"
  }

  image = "${var.switch_image_uuid}"

  #root_authorized_keys = 
  #user_data = ""
  user_script = "${var.master_script}"
}

resource "google_dns_record_set" "master" {
  name         = "${triton_machine.master.name}.${google_dns_managed_zone.subzone.dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas      = ["${triton_machine.master.primaryip}"]
}

/*









resource "docker_image" "influxdb" {
  name = "influxdb:1.0.2-alpine"
  keep_locally = true
}
resource "docker_container" "influxdb" {
  count = 1
  name = "influxdb"
  image = "${docker_image.influxdb.latest}"
  must_run = true
  restart = "always"
  memory = 1024
  log_driver = "json-file"
  log_opts = {
    max-size = "1m"
    max-file = 2
  }
  ports {
    internal = 8083
    external = 8083
  }
  ports {
    internal = 8086
    external = 8086
  }
  ports {
    internal = 8089
    external = 8089
  }
  labels = { "triton.cns.services" = "influxdb"}
}
resource "google_dns_record_set" "influxdb" {
  name = "${docker_container.influxdb.name}.${google_dns_managed_zone.subzone.dns_name}"
  type = "A"
  ttl  = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas = ["${docker_container.influxdb.ip_address}"]
}








resource "docker_image" "grafana" {
  name = "grafana/grafana:4.0.1"
  keep_locally = true
}
resource "docker_container" "grafana" {
  count = 1
  name = "grafana"
  image = "${docker_image.grafana.latest}"
  must_run = true
  restart = "always"
  memory = 1024
  log_driver = "json-file"
  log_opts = {
    max-size = "1m"
    max-file = 2
  }
  ports {
    internal = 3000
    external = 3000
  }
  labels = { "triton.cns.services" = "grafana"}
}
resource "google_dns_record_set" "grafana" {
  name = "${docker_container.grafana.name}.${google_dns_managed_zone.subzone.dns_name}"
  type = "A"
  ttl  = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas = ["${docker_container.grafana.ip_address}"]
}










resource "docker_image" "vault" {
  name = "gregzuro/vault:0.6.4.1"
  keep_locally = true
}
resource "docker_container" "vault" {
  command = []
  count = 1
  name = "vault"
  image = "${docker_image.vault.latest}"
  must_run = true
  restart = "always"
  memory = 1024
  log_driver = "json-file"
  log_opts = {
    max-size = "1m"
    max-file = 2
  }
  ports {
    internal = 8200
    external = 8200
  }
  labels = { "triton.cns.services" = "vault"}
}
resource "google_dns_record_set" "vault" {
  name = "${docker_container.vault.name}.${google_dns_managed_zone.subzone.dns_name}"
  type = "A"
  ttl  = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas = ["${docker_container.vault.ip_address}"]
}




resource "docker_image" "registry" {
  name = "registry:2.5"
  keep_locally = true
}
resource "docker_container" "registry" {
  command = ["/etc/docker/registry/config.yml"]
  count = 1
  name = "registry"
  image = "${docker_image.registry.latest}"
  must_run = true
  restart = "always"
  memory = 1024
  log_driver = "json-file"
  log_opts = {
    max-size = "1m"
    max-file = 2
  }
  ports {
    internal = 5000
    external = 5000
  }
  labels = { "triton.cns.services" = "registry"}
}
resource "google_dns_record_set" "registry" {
  name = "${docker_container.registry.name}.${google_dns_managed_zone.subzone.dns_name}"
  type = "A"
  ttl  = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas = ["${docker_container.registry.ip_address}"]
}




# resource "docker_image" "fluentd" {
#   name = "gregzuro/fluentd"
#   keep_locally = true
# }

 

*/


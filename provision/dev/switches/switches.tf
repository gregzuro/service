provider "google" {
  credentials = "${file("../secrets/gregzuro-dev-45c1d404d431.json")}"
  project     = "gregzuro-dev"
  region      = "us-west1"
}

# account_name must be of the form: "foo_bar{_baz}" where bar is used as the sub-domain and the rest are ignored
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

# 

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

resource "docker_image" "influxdb" {
  name         = "influxdb:1.0.2-alpine"
  keep_locally = true
}

resource "docker_container" "influxdb" {
  count      = 1
  name       = "influxdb"
  image      = "${docker_image.influxdb.latest}"
  must_run   = true
  restart    = "always"
  memory     = 1024
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

  labels = {
    "triton.cns.services" = "influxdb"
  }

  upload {
    content = "${file("${var.influxdb_conf}")}"
    file    = "/etc/influxdb/influxdb.conf"
  }

  provisioner "local-exec" {
    command = "sleep 3"
  }

  provisioner "local-exec" {
    command = "${var.docker_command} exec ${docker_container.influxdb.name} influx -execute \"create database sgms\""
  }

  provisioner "local-exec" {
    command = "${var.docker_command} exec ${docker_container.influxdb.name} influx -execute \"create database logs\""
  }

  provisioner "local-exec" {
    command = "${var.docker_command} exec ${docker_container.influxdb.name} influx -execute \"create user sgms with password 'local'\""
  }

  provisioner "local-exec" {
    command = "${var.docker_command} exec ${docker_container.influxdb.name} influx -execute \"grant all on sgms to sgms\""
  }
}

resource "google_dns_record_set" "influxdb" {
  name         = "${docker_container.influxdb.name}.${google_dns_managed_zone.subzone.dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas      = ["${docker_container.influxdb.ip_address}"]
}

# fluentd
#
#
#

resource "triton_machine" "master" {
  name    = "master"
  package = "g4-highcpu-512M"

  tags = {
    "triton.cns.services" = "master"
  }

  image = "${var.switch_image_uuid}"

  #root_authorized_keys = 
  #user_data = ""
  user_script = "${file("${var.master_script}")}"
}

resource "google_dns_record_set" "master" {
  name         = "${triton_machine.master.name}.${google_dns_managed_zone.subzone.dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas      = ["${triton_machine.master.primaryip}"]
}

/*
resource "docker_image" "registry" {
  name         = "registry:2.5"
  keep_locally = true
}

resource "docker_container" "registry" {
  command    = ["/etc/docker/registry/config.yml"]
  count      = 1
  name       = "registry"
  image      = "${docker_image.registry.latest}"
  must_run   = true
  restart    = "always"
  memory     = 1024
  log_driver = "json-file"

  log_opts = {
    max-size = "1m"
    max-file = 2
  }

  ports {
    internal = 5000
    external = 5000
  }

  labels = {
    "triton.cns.services" = "registry"
  }
}

resource "google_dns_record_set" "registry" {
  name         = "${docker_container.registry.name}.${google_dns_managed_zone.subzone.dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = "${google_dns_managed_zone.subzone.name}"
  rrdatas      = ["${docker_container.registry.ip_address}"]
}
*/


/*

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

*/


# resource "docker_image" "registry" {


#   name         = "registry:2.5"


#   keep_locally = true


# }


# resource "docker_container" "registry" {


#   command    = ["/etc/docker/registry/config.yml"]


#   count      = 1


#   name       = "registry"


#   image      = "${docker_image.registry.latest}"


#   must_run   = true


#   restart    = "always"


#   memory     = 1024


#   log_driver = "json-file"


#   log_opts = {


#     max-size = "1m"


#     max-file = 2


#   }


#   ports {


#     internal = 5000


#     external = 5000


#   }


#   labels = {


#     "triton.cns.services" = "registry"


#   }


# }


# resource "google_dns_record_set" "registry" {


#   name         = "${docker_container.registry.name}.${google_dns_managed_zone.subzone.dns_name}"


#   type         = "A"


#   ttl          = 300


#   managed_zone = "${google_dns_managed_zone.subzone.name}"


#   rrdatas      = ["${docker_container.registry.ip_address}"]


# }


# resource "docker_image" "fluentd" {


#   name = "gregzuro/fluentd"


#   keep_locally = true


# }


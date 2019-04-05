variable "gcloud_project" {
}

variable "gcloud_zone" {
}

variable "gcloud_credentials" {
}

provider "google" {
  credentials = "${var.gcloud_credentials}"
  project = "${var.gcloud_project}"
  zone = "${var.gcloud_zone}"
}

resource "google_pubsub_topic" "falco_alerts" {
  name = "falco-alerts"
}

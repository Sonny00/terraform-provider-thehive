terraform {
  required_providers {
    thehive = {
      source = "registry.terraform.io/Sonny00/thehive"
    }
  }
}

provider "thehive" {
  url     = "http://localhost:9000"
  api_key = var.api_key
}

variable "api_key" {
  type      = string
  sensitive = true
}

# --- Test validators with bad values ---
resource "thehive_case_template" "bad" {
  name     = "bad-template"
  severity = 99
  tlp      = 10
  pap      = 5
}

resource "thehive_user" "bad" {
  login        = "bad@test.com"
  name         = "Bad"
  profile      = "analyst"
  organisation = "admin"
  type         = "blabla"
}

resource "thehive_custom_field" "bad" {
  name  = "bad_field"
  type  = "invalid"
  group = "default"
}
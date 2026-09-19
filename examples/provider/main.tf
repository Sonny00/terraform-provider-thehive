terraform {
  required_providers {
    thehive = {
      source = "Sonny00/thehive"
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

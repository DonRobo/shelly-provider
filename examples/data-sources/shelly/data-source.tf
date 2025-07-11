terraform {
  required_providers {
    shelly = {
      source = "github.com/DonRobo/shelly-provider"
    }
  }
}


provider "shelly" {
  ip = "192.168.1.169"
}

data "shelly_version" "example" {}

output "shelly_fw_version" {
  value = data.shelly_version.example.version
}
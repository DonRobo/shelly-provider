terraform {
  required_providers {
    shelly = {
      source = "github.com/DonRobo/shelly-provider"
    }
  }
}


provider "shelly" {
}

data "shelly_device" "example" {
    ip = "192.168.1.169"
}

output "shelly_fw_version" {
  value = data.shelly_device.example.version
}
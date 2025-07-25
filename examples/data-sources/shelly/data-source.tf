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
  ip = "192.168.1.100"
}

output "device_mac" {
  value = data.shelly_device.example.mac
}

output "device_version" {
  value = data.shelly_device.example.version
}
terraform {
  required_providers {
    shelly = {
      source = "github.com/DonRobo/shelly-provider"
    }
  }
}

provider "shelly" {
  # Device IP has to be set in each resource or data source
}
terraform {
  required_providers {
    tfcoremock = {
      source  = "hashicorp/tfcoremock"
      version = "0.1.1"
    }
  }
}

provider "tfcoremock" {}

resource "tfcoremock_map" "map" {
  id = "50E1A46E-E64A-4C1F-881C-BA85A5440964"
  map = {
    "zero" : "6B044AF7-172B-495B-BE11-B9546C12C3BD",
    "one" : "682672C7-0918-4448-8342-887BAE01062A",
    "two" : "212FFBF6-40FE-4862-B708-E6AA508E84E0",
  }
}

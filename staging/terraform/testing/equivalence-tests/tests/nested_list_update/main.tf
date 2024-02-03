terraform {
  required_providers {
    tfcoremock = {
      source  = "hashicorp/tfcoremock"
      version = "0.1.1"
    }
  }
}

provider "tfcoremock" {}

resource "tfcoremock_nested_list" "nested_list" {
  id = "DA051126-BAD6-4EB2-92E5-F0250DAF0B92"

  lists = [
    ["44E1C623-7B70-4D78-B4D3-D9CFE8A6D982"],
    ["8B031CD1-01F7-422C-BBE6-FF8A0E18CDFD"],
    ["13E3B154-7B85-4EAA-B3D0-E295E7D71D7F"],
  ]
}

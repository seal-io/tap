terraform {
  required_providers {
    tfcoremock = {
      source = "hashicorp/tfcoremock"
      version = "0.1.1"
    }
  }
}

provider "tfcoremock" {}

resource "tfcoremock_nested_set" "nested_set" {
  id = "510598F6-83FE-4090-8986-793293E90480"

  sets = [
    ["29B6824A-5CB6-4C25-A359-727BAFEF25EB"],
    ["9373D62D-1BF0-4F17-B100-7C0FBE368ADE"],
    ["7E90963C-BE32-4411-B9DD-B02E7FE75766"],
  ]
}

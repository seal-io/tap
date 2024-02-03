terraform {
  required_providers {
    tfcoremock = {
      source  = "hashicorp/tfcoremock"
      version = "0.1.1"
    }
  }
}

provider "tfcoremock" {}

resource "tfcoremock_list" "list" {
  id = "985820B3-ACF9-4F00-94AD-F81C5EA33663"
  list = [
    "9C2BE420-042D-440A-96E9-75565341C994",
    "D01290F6-2D3A-45FA-B006-DAA80F6D31F6",
    "9B9F3ADF-8AD4-4E8C-AFE4-7BC2413E9AC0",
  ]
}

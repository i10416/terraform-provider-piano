terraform {

  required_providers {
    piano = {
      source  = "registry.terraform.io/i10416/piano"
      version = "~> 0.1.3"
    }
  }

}

provider "piano" {
  endpoint  = "https://sandbox.piano.io/api/v3"
  api_token = var.PIANO_API_TOKEN
  app_id = var.PIANO_APP_ID
}

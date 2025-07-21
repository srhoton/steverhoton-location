terraform {
  backend "s3" {
    bucket = "steve-rhoton-tfstate"
    key    = "sr-location-sandbox/terraform.tfstate"
    region = "us-west-2"
  }
}